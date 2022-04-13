package cmd

import "regexp"

// Used to indicate handled failure
type SimulationError struct {
	err string
}

func (e *SimulationError) Error() string {
	return e.err
}

func isInSlice(searchString string, slice []string) bool {
	for _, sliceItem := range slice {
		if sliceItem == searchString {
			return true
		}
	}
	return false
}
func isNotInSlice(searchString string, slice []string) bool {
	return !isInSlice(searchString, slice)
}

// Create a dictionary from regex capture groups
func createMatchMap(regexp *regexp.Regexp, line *string) map[string]string {

	valueList := regexp.FindStringSubmatch(*line)
	result := make(map[string]string)

	if len(valueList) > 0 {
		for i, name := range regexp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = valueList[i]
			}
		}
	}
	return result
}

// based on https://gist.github.com/lelandbatey/a5c957b537bed39d1d6fb202c3b8de06
//func setField(item interface{}, fieldName string, value interface{}) error {
//	v := reflect.ValueOf(item).Elem()
//	if !v.CanAddr() {
//		return fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
//	}
//	fieldNames := map[string]int{}
//	for i := 0; i < v.NumField(); i++ {
//		typeField := v.Type().Field(i)
//		fieldNames[typeField.Name] = i
//	}
//
//	fieldNum, ok := fieldNames[fieldName]
//	if !ok {
//		return fmt.Errorf("field %s does not exist within the provided item", fieldName)
//	}
//	fieldVal := v.Field(fieldNum)
//	fieldVal.Set(reflect.ValueOf(value))
//	return nil
//}
