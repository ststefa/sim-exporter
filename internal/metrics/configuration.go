package metrics

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"

	"git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/pkg/errors"
	"github.com/docker/go-units"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/guregu/null.v4"
	"gopkg.in/yaml.v2"
)

var (
	// How a value-range must look like: a single value or "value-value" where
	// value is either plain digits or golang scientific notation (e.g. 3.123e+10)
	// or human readable (e.g. "20GB")
	// The resulting regex will have a mandatory "from" and an optional "to" group
	regexpValueRange = regexp.MustCompile(`(?P<from>[0-9a-zA-Z.+]+)(-(?P<to>[0-9a-zA-Z.+]*))?`)
)

type Configuration struct {
	Version string                         `yaml:"version"`
	Metrics map[string]ConfigurationMetric `yaml:"metrics"`
}

type ConfigurationMetric struct {
	Name string `yaml:"name"`
	Help string `yaml:"help"`
	Type string `yaml:"type"`

	Labels []string `yaml:"labels"`

	Items []ConfigurationMetricItem `yaml:"items"`

	prometheus struct {
		gauge     *prometheus.GaugeVec
		counter   *prometheus.CounterVec
		summary   *prometheus.SummaryVec
		histogram *prometheus.HistogramVec
	}
}

func (m *ConfigurationMetric) Init() error {

	for index := range m.Items {
		item := &m.Items[index]

		if item.Value != "" {
			err := item.parseValue()
			if err != nil {
				return err
			}

		}
	}
	return nil
}

type ConfigurationMetricItem struct {
	Value     string `yaml:"value"`
	value     null.Float
	rangeFrom null.Float
	rangeTo   null.Float
	Labels    map[string]string `yaml:"labels"`
}

func (m *ConfigurationMetricItem) parseValue() error {
	matchMap := createMatchMap(regexpValueRange, &m.Value)
	if len(matchMap) >= 1 {
		rangeFrom, err := parseFloatFromString(matchMap["from"])
		if err != nil {
			return err
		}

		// The "to" capture group is optional. If absent, the metric is constant
		if matchMap["to"] == "" {
			m.value = rangeFrom
		} else {
			rangeTo, err := parseFloatFromString(matchMap["to"])
			if err != nil {
				return err
			}
			m.rangeFrom = rangeFrom
			m.rangeTo = rangeTo
		}
		return nil
	} else {
		return fmt.Errorf("value %v not parsable with regex %v", m.Value, regexpValueRange)
	}
}

func (m *ConfigurationMetricItem) generateValue() null.Float {
	if m.value.Valid {
		return m.value
	} else if m.rangeFrom.Valid && m.rangeTo.Valid {
		result := (rand.Float64() * (m.rangeTo.ValueOrZero() - m.rangeFrom.ValueOrZero())) + m.rangeFrom.ValueOrZero()
		return null.FloatFrom(result)
	}
	return null.NewFloat(0, false)
}

func LoadAndValidateConfiguration(filename string) (*Configuration, error) {
	var config Configuration
	rawBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	log.Debugf("rawBytes from %v: %s\n", filename, rawBytes)

	err = yaml.Unmarshal([]byte(rawBytes), &config)
	if err != nil {
		return nil, err
	}
	log.Debugf("config: %v\n", config)

	var validationErrors []string

	if len(config.Metrics) == 0 {
		validationErrors = append(validationErrors, "metrics must have one or more elements")
	}
	for metricName, metric := range config.Metrics {
		if !(metric.Type == "gauge" || metric.Type == "counter" || metric.Type == "summary" || metric.Type == "histogram") {
			validationErrors = append(validationErrors, fmt.Sprintf("metrics.%v.type: Unknown metric type %q. Must be one of gauge, counter, summary, histogram", metricName, metric.Type))
		}
		if len(metric.Items) == 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("metrics.%v.items: Must have at least one element", metricName))
		}
		if len(metric.Labels) > 0 {
			for _, item := range metric.Items {
				var itemLabelSlice []string
				for itemLabelName, _ := range item.Labels {
					itemLabelSlice = append(itemLabelSlice, itemLabelName)
					if isNotInSlice(itemLabelName, metric.Labels) {
						validationErrors = append(validationErrors, fmt.Sprintf("metrics.%v.items: Item label %q not declared in metric labels", metricName, itemLabelName))
					}
				}
				for _, metricLabelName := range metric.Labels {
					if isNotInSlice(metricLabelName, itemLabelSlice) {
						validationErrors = append(validationErrors, fmt.Sprintf("metrics.%v.items: Metric label %q missing from item labels", metricName, metricLabelName))
					}
				}

			}

		}
	}

	if len(validationErrors) > 0 {
		errorMessage := ""
		for i := 0; i < len(validationErrors); i++ {
			errorMessage += fmt.Sprintf("%v. %v", i+1, validationErrors[i]) + "; "
		}
		return nil, &errors.SimulationError{Err: fmt.Sprintf("%v has validation errors: %v", filename, errorMessage)}
	}

	return &config, nil
}

func parseFloatFromString(value string) (null.Float, error) {
	ret, err := strconv.ParseFloat(value, 64)
	if err != nil {
		tmp, err := units.FromHumanSize(value)
		if err != nil {
			return null.NewFloat(0, false), err
		}
		ret = float64(tmp)
	}

	return null.FloatFrom(ret), nil
}
