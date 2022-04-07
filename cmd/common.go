package cmd

import (
	"container/list"
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
	// Defaults might be overridden at link time

	// Default listening port
	port = 9041

	// Default config file name
	configFile = "sim-exporter.yaml"

	// Default URI path to metrics
	metricsPath = "metrics"

	// The proper version is automatically set to the contents of `version.txt` at link-time, see Makefile
	version = "dev"

	// Whether to enable debug output
	debug = false

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
	log.Debugf("config: %s\n", yamlBytes)
	if err != nil {
		return config, errors.Wrap(err, "Cannot parse config")
	}

	validationErrors := list.New()
	var metricNames []string
	for _, metric := range config.Metrics {
		metricNames = append(metricNames, metric.Name)
	}
	var instanceTypeNames []string
	for _, instanceType := range config.InstanceTypes {
		metricNames = append(instanceTypeNames, instanceType.Name)
	}

	for _, metric := range config.Metrics {
		if metric.Fuzzy < 0 {
			validationErrors.PushBack(fmt.Sprintf("metric.%v: Fuzzy must be >= 0", metric.Name))
		}
		if metric.LowerLimit >= metric.UpperLimit {
			validationErrors.PushBack(fmt.Sprintf("metric.%v: lowerLimit must be smaller than upperLimit", metric.Name))
		}
	}
	// validate metric refs in instance_types
	for _, instanceType := range config.InstanceTypes {
		for _, metricRef := range instanceType.MetricRefs {
			if isNotInSlice(metricRef.Ref, metricNames) {
				validationErrors.PushBack(fmt.Sprintf("metricref.%v: Referencing non-existing metric %v", instanceType.Name, metricRef.Ref))
			}
		}
	}

	// validate kinds in fleet
	for _, fleetMember := range config.Fleet {
		if isNotInSlice(fleetMember.Kind, instanceTypeNames) {
			validationErrors.PushBack(fmt.Sprintf("metricref.%v: Referencing non-existing instance type %v", fleetMember.Kind, fleetMember.Kind))
		}
	}
	//if valid.HasErrors() {
	//	return config, errors.Errorf("%v", valid.Errors)
	//}
	if validationErrors.Len() > 0 {
		errorMessage := ""
		for e := validationErrors.Front(); e != nil; e = e.Next() {
			errorMessage += fmt.Sprint(e.Value) + ";"
		}
		return config, &ValidationError{errorMessage}
	}
	return config, nil
}

func Execute() error {
	return rootCmd.Execute()
}
