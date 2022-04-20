package metrics

import (
	"os"
	"regexp"

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

// Test whether searchString is an element of slice
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
