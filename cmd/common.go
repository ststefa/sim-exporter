package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
)

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
