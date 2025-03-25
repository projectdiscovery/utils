package structs

import (
	"reflect"
	"testing"
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

func TestFilterStruct(t *testing.T) {
	s := TestStruct{
		Name:    "John",
		Age:     30,
		Address: "New York",
	}

	tests := []struct {
		name          string
		input         interface{}
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

func TestGetStructFields(t *testing.T) {
	s := TestStruct{
		Name:    "John",
		Age:     30,
		Address: "New York",
	}

	tests := []struct {
		name    string
		input   interface{}
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
