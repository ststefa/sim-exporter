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

	// How a prometheus metric line must look like
	regexpMetricItem = *regexp.MustCompile(`^(?P<name>\w+)\s*(?:|{(?P<labels>[^}]*)})\s+(?P<value>[^\s]*).*$`)

	// Valid functions
	// This should really be a constant but golang will not let me
	validFunctions = []string{"rand", "asc", "desc", "sin"}

	// Valid prometheus metric types
	// This should also be a constant
	validMetricTypes = []string{"gauge", "counter", "summary", "histogram"}
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

// Test whether two string slices are equal. Note that order matters. To
// ensure specific order use e.g. sort.Strings
func stringSlicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Create a dictionary from regex capture groups
func createMatchMap(regexp regexp.Regexp, line string) map[string]string {

	valueList := regexp.FindStringSubmatch(line)
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
