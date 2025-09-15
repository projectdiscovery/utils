package structs

import (
	"reflect"
	"testing"

	mapsutil "github.com/projectdiscovery/utils/maps"
)

type TestStruct struct {
	Name    string
	Age     int
	Address string
}

type NestedStruct struct {
	Basic    TestStruct
	PtrField *TestStruct
}

type MapTestStruct struct {
	Name     string `json:"name"`
	Age      int    `json:"age,omitempty"`
	Password string `json:"-"`
	IsActive bool   `json:"is_active"`
	Address  string `json:"address"`
	Country  string `json:"country,omitempty"`
}

func TestFilterStruct(t *testing.T) {
	s := TestStruct{
		Name:    "John",
		Age:     30,
		Address: "New York",
	}

	tests := []struct {
		name          string
		input         any
		includeFields []string
		excludeFields []string
		want          TestStruct
		wantErr       bool
	}{
		{
			name:          "include specific fields",
			input:         s,
			includeFields: []string{"name", "Age"},
			excludeFields: []string{},
			want: TestStruct{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:          "exclude specific fields",
			input:         s,
			includeFields: []string{},
			excludeFields: []string{"address"},
			want: TestStruct{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:          "non-struct input",
			input:         "not a struct",
			includeFields: []string{},
			excludeFields: []string{},
			want:          TestStruct{},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterStruct(tt.input, tt.includeFields, tt.excludeFields)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("FilterStruct() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFilterStructToMap(t *testing.T) {
	s := MapTestStruct{
		Name:     "John",
		Age:      30,                // To test omitempty on a non zero value
		Password: "secret-password", // To test an ignored tag (json:"-")
		IsActive: true,
		Address:  "New York",
		Country:  "", // To test omitempty on a zero value
	}

	tests := []struct {
		name          string
		input         any
		includeFields []string
		excludeFields []string
		want          *mapsutil.OrderedMap[string, any]
		wantErr       bool
	}{
		{
			name:          "no filtering",
			input:         s,
			includeFields: nil,
			excludeFields: nil,
			want: func() *mapsutil.OrderedMap[string, any] {
				om := mapsutil.NewOrderedMap[string, any]()
				om.Set("name", "John")
				om.Set("age", 30)
				om.Set("is_active", true)
				om.Set("address", "New York")
				return &om
			}(),
			wantErr: false,
		},
		{
			name:          "include specific fields",
			input:         s,
			includeFields: []string{"Name", "Address"},
			excludeFields: []string{},
			want: func() *mapsutil.OrderedMap[string, any] {
				om := mapsutil.NewOrderedMap[string, any]()
				om.Set("name", "John")
				om.Set("address", "New York")
				return &om
			}(),
			wantErr: false,
		},
		{
			name:          "exclude specific fields",
			input:         s,
			includeFields: []string{},
			excludeFields: []string{"Address", "IsActive"},
			want: func() *mapsutil.OrderedMap[string, any] {
				om := mapsutil.NewOrderedMap[string, any]()
				om.Set("name", "John")
				om.Set("age", 30)
				return &om
			}(),
			wantErr: false,
		},
		{
			name:          "include and exclude",
			input:         s,
			includeFields: []string{"Name", "Age", "Address"},
			excludeFields: []string{"Age"},
			want: func() *mapsutil.OrderedMap[string, any] {
				om := mapsutil.NewOrderedMap[string, any]()
				om.Set("name", "John")
				om.Set("address", "New York")
				return &om
			}(),
			wantErr: false,
		},
		{
			name:          "non-struct input",
			input:         "not a struct",
			includeFields: []string{},
			excludeFields: []string{},
			want:          nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterStructToMap(tt.input, tt.includeFields, tt.excludeFields)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterStructToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterStructToMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterStructToMapOrder(t *testing.T) {
	type OrderTestStruct struct {
		FirstField  string `json:"first_field"`
		SecondField int    `json:"second_field"`
		ThirdField  bool   `json:"third_field"`
		FourthField string `json:"fourth_field"`
	}

	s := OrderTestStruct{
		FirstField:  "first",
		SecondField: 42,
		ThirdField:  true,
		FourthField: "fourth",
	}

	t.Run("field order follows struct definition", func(t *testing.T) {
		result, err := FilterStructToMap(s, nil, nil)
		if err != nil {
			t.Fatalf("FilterStructToMap() error = %v", err)
		}

		keys := result.GetKeys()
		expectedOrder := []string{"first_field", "second_field", "third_field", "fourth_field"}

		if len(keys) != len(expectedOrder) {
			t.Errorf("Expected %d keys, got %d", len(expectedOrder), len(keys))
		}

		for i, expectedKey := range expectedOrder {
			if i >= len(keys) {
				t.Errorf("Missing key at index %d: expected %s", i, expectedKey)
				continue
			}
			if keys[i] != expectedKey {
				t.Errorf("Key at index %d: expected %s, got %s", i, expectedKey, keys[i])
			}
		}
	})

	t.Run("field order preserved with include filter", func(t *testing.T) {
		result, err := FilterStructToMap(s, []string{"ThirdField", "FirstField"}, nil)
		if err != nil {
			t.Fatalf("FilterStructToMap() error = %v", err)
		}

		keys := result.GetKeys()
		expectedOrder := []string{"first_field", "third_field"} // Struct order, not include order

		if len(keys) != len(expectedOrder) {
			t.Errorf("Expected %d keys, got %d", len(expectedOrder), len(keys))
		}

		for i, expectedKey := range expectedOrder {
			if i >= len(keys) {
				t.Errorf("Missing key at index %d: expected %s", i, expectedKey)
				continue
			}
			if keys[i] != expectedKey {
				t.Errorf("Key at index %d: expected %s, got %s", i, expectedKey, keys[i])
			}
		}
	})

	t.Run("field order preserved with exclude filter", func(t *testing.T) {
		result, err := FilterStructToMap(s, nil, []string{"SecondField"})
		if err != nil {
			t.Fatalf("FilterStructToMap() error = %v", err)
		}

		keys := result.GetKeys()
		expectedOrder := []string{"first_field", "third_field", "fourth_field"}

		if len(keys) != len(expectedOrder) {
			t.Errorf("Expected %d keys, got %d", len(expectedOrder), len(keys))
		}

		for i, expectedKey := range expectedOrder {
			if i >= len(keys) {
				t.Errorf("Missing key at index %d: expected %s", i, expectedKey)
				continue
			}
			if keys[i] != expectedKey {
				t.Errorf("Key at index %d: expected %s, got %s", i, expectedKey, keys[i])
			}
		}
	})
}

func TestGetStructFields(t *testing.T) {
	s := TestStruct{
		Name:    "John",
		Age:     30,
		Address: "New York",
	}

	tests := []struct {
		name    string
		input   any
		want    []string
		wantErr bool
	}{
		{
			name:    "valid struct",
			input:   s,
			want:    []string{"name", "age", "address"},
			wantErr: false,
		},
		{
			name:    "non-struct input",
			input:   "not a struct",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStructFields(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStructFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetStructFields() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
