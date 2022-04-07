package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

/*
	To avoid confusion (or until I have a better understanding of go scoping)
	I decided to put anything which is shared between multiple files in  a
	directory/package into this here `common.go`
*/

var (
	log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
)

// Metric Main configuration struct
type Metric struct {
	Name       string `yaml:"name"`
	Func       string `yaml:"func"`
	Fuzzy      int64  `yaml:"fuzzy"`
	LowerLimit int64  `yaml:"lowerLimit"`
	UpperLimit int64  `yaml:"upperLimit"`
}
type MetricRef struct {
	Ref string `yaml:"ref"`
}
type InstanceType struct {
	Name       string      `yaml:"name"`
	MetricRefs []MetricRef `yaml:"metricRefs"`
}
type Fleet struct {
	Kind string `yaml:"kind"`
	Num  int64  `yaml:"num"`
}
type Config struct {
	Metrics       []Metric       `yaml:"metrics"`
	InstanceTypes []InstanceType `yaml:"instanceTypes"`
	Fleet         []Fleet        `yaml:"fleet"`
}

type ValidationError struct {
	err string
}

func (e *ValidationError) Error() string {
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

func readAndValidateConfig(filename string) (Config, error) {
	var config Config
	yamlBytes, err := os.ReadFile(filename)
	if err != nil {
		return config, errors.Wrap(err, "Cannot read config")
	}
	log.Debugf("yamlBytes from %v: %s\n", filename, yamlBytes)
	err = yaml.Unmarshal(yamlBytes, &config)
	log.Debugf("config: %v\n", config)
	if err != nil {
		return config, errors.Wrap(err, "Cannot parse config")
	}

	var validationErrors []string
	var metricNames []string

	// Generate name slices for validation
	for _, metric := range config.Metrics {
		metricNames = append(metricNames, metric.Name)
	}
	var instanceTypeNames []string
	for _, instanceType := range config.InstanceTypes {
		instanceTypeNames = append(instanceTypeNames, instanceType.Name)
	}

	// Validate metrics
	if len(config.Metrics) == 0 {
		validationErrors = append(validationErrors, "metrics must have one or more elements")
	}
	for _, metric := range config.Metrics {
		if metric.Fuzzy < 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("metrics.[name=%v].fuzzy: Must be >= 0", metric.Name))
		}
		if metric.LowerLimit >= metric.UpperLimit {
			validationErrors = append(validationErrors, fmt.Sprintf("metrics.[name=%v].lowerLimit: Must be smaller than upperLimit", metric.Name))
		}
	}
	// Validate instanceTypes
	if len(config.InstanceTypes) == 0 {
		validationErrors = append(validationErrors, "instanceTypes must have one or more elements")
	}
	for _, instanceType := range config.InstanceTypes {
		for _, metricRef := range instanceType.MetricRefs {
			if isNotInSlice(metricRef.Ref, metricNames) {
				validationErrors = append(validationErrors, fmt.Sprintf("instanceTypes.[name=%v].metricrefs.[ref=%v]: Referencing non-existing metric name %v", instanceType.Name, metricRef.Ref, metricRef.Ref))
			}
		}
	}

	// Validate fleet
	if len(config.Fleet) == 0 {
		validationErrors = append(validationErrors, "fleet must have one or more elements")
	}
	for _, fleetMember := range config.Fleet {
		if fleetMember.Num < 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("fleet.[kind=%v].num: Must be >= 0", fleetMember.Kind))
		}
		if isNotInSlice(fleetMember.Kind, instanceTypeNames) {
			validationErrors = append(validationErrors, fmt.Sprintf("fleet.[kind=%v]: Referencing non-existing instance type %v", fleetMember.Kind, fleetMember.Kind))
		}
	}
	if len(validationErrors) > 0 {
		errorMessage := ""
		for i := 0; i < len(validationErrors); i++ {
			errorMessage += fmt.Sprintf("%v. %v", i+1, validationErrors[i]) + "; "
		}
		return config, &ValidationError{errorMessage}
	}
	return config, nil
}
