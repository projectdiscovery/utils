package reflectutil

import (
	"strings"
	"testing"

	"github.com/projectdiscovery/utils/reflect/tests"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	FirstOption    string
	SecondOption   int
	privateOption3 string
}

func TestToMap(t *testing.T) {
	testStruct := TestStruct{
		FirstOption:    "test",
		SecondOption:   10,
		privateOption3: "ignored",
	}
	// testing normal fields
	tomap, err := ToMap(testStruct, nil, false)
	require.Nilf(t, err, "error while parsing: %s", err)
	m := map[string]interface{}{"first_option": "test", "second_option": 10}
	require.EqualValues(t, m, tomap, "objects are not equal")

	// testing with non exported ones
	tomap, err = ToMap(testStruct, nil, true)
	require.Nilf(t, err, "error while parsing: %s", err)
	m = map[string]interface{}{"first_option": "test", "second_option": 10, "private_option3": "ignored"}
	require.EqualValues(t, m, tomap, "objects are not equal")

	// testing with custom stringify function
	fu := func(s string) string {
		return strings.ToLower(s)
	}
	tomap, err = ToMap(testStruct, fu, false)
	require.Nilf(t, err, "error while parsing: %s", err)
	m = map[string]interface{}{"firstoption": "test", "secondoption": 10}
	require.EqualValues(t, m, tomap, "objects are not equal")
}

func TestUnexportedField(t *testing.T) {
	// create a pointer instance to a struct with an "unexported" field
	testStruct := &tests.Test{}
	SetUnexportedField(testStruct, "unexported", "test")
	value := GetUnexportedField(testStruct, "unexported")
	require.Equal(t, value, "test")
}

// Test taken from https://github.com/DmitriyVTitov/size/blob/v1.5.0/size_test.go
func TestSizeOf(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want int
	}{
		{
			name: "Array",
			v:    [3]int32{1, 2, 3}, // 3 * 4  = 12
			want: 12,
		},
		{
			name: "Slice",
			v:    make([]int64, 2, 5), // 5 * 8 + 24 = 64
			want: 64,
		},
		{
			name: "String",
			v:    "ABCdef", // 6 + 16 = 22
			want: 22,
		},
		{
			name: "Map",
			// (8 + 3 + 16) + (8 + 4 + 16) = 55
			// 55 + 8 + 10.79 * 2 = 84
			v:    map[int64]string{0: "ABC", 1: "DEFG"},
			want: 84,
		},
		{
			name: "Struct",
			v: struct {
				slice     []int64
				array     [2]bool
				structure struct {
					i int8
					s string
				}
			}{
				slice: []int64{12345, 67890}, // 2 * 8 + 24 = 40
				array: [2]bool{true, false},  // 2 * 1 = 2
				structure: struct {
					i int8
					s string
				}{
					i: 5,     // 1
					s: "abc", // 3 * 1 + 16 = 19
				}, // 20 + 7 (padding) = 27
			}, // 40 + 2 + 27 = 69 + 6 (padding) = 75
			want: 75,
		},
		{
			name: "Pointer",
			v:    new(int64), // 8
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SizeOf(tt.v); got != tt.want {
				t.Errorf("Of() = %v, want %v", got, tt.want)
			}
		})
	}
}
