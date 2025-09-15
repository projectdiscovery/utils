package structs

import (
	"errors"
	"reflect"
	"strings"

	mapsutil "github.com/projectdiscovery/utils/maps"
)

// CallbackFunc on the struct field
// example:
// structValue := reflect.ValueOf(s)
// ...
// field := structValue.Field(i)
// fieldType := structValue.Type().Field(i)
type CallbackFunc func(reflect.Value, reflect.StructField)

// Walk traverses a struct and executes a callback function on each field in the struct.
// The interface{} passed to the function should be a pointer to a struct
func Walk(s interface{}, callback CallbackFunc) {
	structValue := reflect.ValueOf(s)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}
	if structValue.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structValue.Type().Field(i)
		if !fieldType.IsExported() {
			continue
		}
		if field.Kind() == reflect.Struct {
			Walk(field.Addr().Interface(), callback)
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			Walk(field.Interface(), callback)
		} else {
			callback(field, fieldType)
		}
	}
}

func walkFilteredFields[T any](input T, includeFields, excludeFields []string, walker func(field reflect.StructField, value reflect.Value)) error {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}

	includeMap := make(map[string]bool)
	excludeMap := make(map[string]bool)

	for _, field := range includeFields {
		includeMap[strings.ToLower(field)] = true
	}
	for _, field := range excludeFields {
		excludeMap[strings.ToLower(field)] = true
	}

	typeOfStruct := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typeOfStruct.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := strings.ToLower(field.Name)
		fieldValue := val.Field(i)

		if (len(includeMap) == 0 || includeMap[fieldName]) && !excludeMap[fieldName] {
			walker(field, fieldValue)
		}
	}
	return nil
}

// FilterStruct filters the struct based on include and exclude fields and returns a new struct.
// - input: the original struct.
// - includeFields: list of fields to include (if empty, includes all).
// - excludeFields: list of fields to exclude (processed after include).
func FilterStruct[T any](input T, includeFields, excludeFields []string) (T, error) {
	var zeroValue T
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	filteredStruct := reflect.New(val.Type()).Elem()

	walker := func(field reflect.StructField, value reflect.Value) {
		filteredStruct.FieldByName(field.Name).Set(value)
	}

	if err := walkFilteredFields(input, includeFields, excludeFields, walker); err != nil {
		return zeroValue, err
	}

	return filteredStruct.Interface().(T), nil
}

func FilterStructToMap[T any](input T, includeFields, excludeFields []string) (*mapsutil.OrderedMap[string, any], error) {
	resultMap := mapsutil.NewOrderedMap[string, any]()

	walker := func(field reflect.StructField, value reflect.Value) {
		jsonTag := field.Tag.Get("json")
		jsonKey := strings.Split(jsonTag, ",")[0]

		if jsonKey == "" || jsonKey == "-" {
			return
		}

		fieldValue := value.Interface()
		if strings.Contains(jsonTag, "omitempty") && value.IsZero() {
			return
		}

		resultMap.Set(jsonKey, fieldValue)
	}

	if err := walkFilteredFields(input, includeFields, excludeFields, walker); err != nil {
		return nil, err
	}

	return &resultMap, nil
}

// GetStructFields returns all the top-level field names from the given struct.
// - input: the original struct.
// Returns a slice of field names or an error if the input is not a struct.
func GetStructFields[T any](input T) ([]string, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct")
	}

	fields := make([]string, 0, val.NumField())
	typeOfStruct := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fields = append(fields, strings.ToLower(typeOfStruct.Field(i).Name))
	}

	return fields, nil
}
