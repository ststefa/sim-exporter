package metrics

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
)

type Collection struct {
	Version string    `yaml:"version"`
	Metrics []*Metric `yaml:"metrics"`
}

func (c *Collection) AddMetric(m Metric) error {
	if m.Name == "" {
		return fmt.Errorf("metric name is required")
	}
	if _, ok := c.GetMetric(m.Name); ok {
		return fmt.Errorf("metric %q already in collection", m.Name)
	}
	m.parent = c
	c.Metrics = append(c.Metrics, &m)
	return nil
}

func (c *Collection) GetMetric(n string) (*Metric, bool) {
	for i := range c.Metrics {
		if c.Metrics[i].Name == n {
			return c.Metrics[i], true
		}
	}
	return nil, false
}

//func (c *Collection) DeleteMetric(n string) {
//	for i := range c.Metrics {
//		if c.Metrics[i].Name == n {
//			// Reslicing is required to remove an element from a slice
//			c.Metrics = append(c.Metrics[:i], c.Metrics[i+1:]...)
//		}
//	}
//}

/* Abandoned attempt to implement custom unmarshaling
func (c *Collection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	//isInSlice := func(searchString reflect.Value, slice []reflect.Value) bool {
	//	for _, sliceItem := range slice {
	//		if sliceItem == searchString {
	//			return true
	//		}
	//	}
	//	return false
	//}

	var validationErrors []string

	if _, ok := raw["version"]; !ok {
		validationErrors = append(validationErrors, "missing version")
	} else {
		c.Version = reflect.ValueOf(raw["version"]).String()
	}

	if _, ok := raw["metrics"]; ok {
		if raw["metrics"] == nil {
			validationErrors = append(validationErrors, "one or more metrics are required")
		} else {
			c.Metrics = make(map[string]Metric)
			metricsIter := reflect.ValueOf(raw["metrics"]).MapRange()
			for metricsIter.Next() {
				k := metricsIter.Key().Elem().String()
				v := metricsIter.Value().Elem().String()
				fmt.Printf("k:%v v:%v\n", k, v)
			}
		}
	}

	if len(c.Metrics) == 0 {
		validationErrors = append(validationErrors, "one or more metrics are required")
	}
	if len(validationErrors) > 0 {
		errorMessage := ""
		for i := 0; i < len(validationErrors); i++ {
			errorMessage += fmt.Sprintf("%v. %v", i+1, validationErrors[i]) + "; "
		}
		return fmt.Errorf("input has %v validation errors: %v", len(validationErrors), errorMessage)
	}

	return nil
}
*/

type Metric struct {
	Name   string   `yaml:"name"`
	Help   string   `yaml:"help"`
	Type   string   `yaml:"type"`
	Labels []string `yaml:"labels"`

	Items []*MetricItem `yaml:"items"`

	parent *Collection

	// min and max values of all child items
	//min float64
	//max float64

	prometheus struct {
		gauge     *prometheus.GaugeVec
		counter   *prometheus.CounterVec
		summary   *prometheus.SummaryVec
		histogram *prometheus.HistogramVec
	}
}

func (m *Metric) AddItem(i MetricItem) error {

	i.parent = m

	var labelNames []string
	for name := range i.Labels {
		labelNames = append(labelNames, name)
	}
	sort.Strings(labelNames)

	if m.Labels == nil {
		m.Labels = labelNames
	} else {
		if !stringSlicesEqual(m.Labels, labelNames) {
			return fmt.Errorf("label mismatch, want %q, got %q", m.Labels, labelNames)
		}
	}

	//if i.Min < m.min {
	//	m.min = i.Min
	//}
	//if i.Max > m.max {
	//	m.max = i.Max
	//}

	m.Items = append(m.Items, &i)

	return nil
}

func (m *Metric) ParentCollection() *Collection {
	return m.parent
}

type MetricItem struct {
	Min      float64           `yaml:"min"`
	Max      float64           `yaml:"max"`
	Func     string            `yaml:"func"`
	Interval time.Duration     `yaml:"interval"`
	Labels   map[string]string `yaml:"labels"`

	parent *Metric
}

func (i *MetricItem) ParentMetric() *Metric {
	return i.parent
}

// Compute duration since start modulo interval (i.e. the interval repeats endlessly)
func interval(start time.Time, interval time.Duration) (time.Duration, error) {
	result := time.Duration(0)
	passed := int64(time.Since(start))
	intv := int64(interval)
	if intv == 0 {
		return result, fmt.Errorf("interval cannot be 0")
	}
	sec := float64(passed % intv)
	result = time.Duration(sec)
	return result, nil
}

// Generate a new value upon refresh
func (i *MetricItem) generateValue(start time.Time) (float64, error) {
	var result float64
	if i.Min == i.Max {
		result = i.Min
	} else {
		elapsed, err := interval(start, i.Interval) // elapsed duration in interval
		if err != nil {
			return 0, err
		}

		switch i.Func {
		case "rand":
			{
				result = i.Min + ((i.Max - i.Min) * rand.Float64())
			}
		case "asc":
			{
				intervalFactor := float64(elapsed) / float64(i.Interval) // 0 at start of interval, 1 at end
				result = i.Min + ((i.Max - i.Min) * intervalFactor)
			}
		case "desc":
			{
				intervalFactor := float64(elapsed) / float64(i.Interval)
				result = i.Max - ((i.Max - i.Min) * intervalFactor)
			}
		case "sin":
			{
				intervalFactor := (float64(elapsed) / float64(i.Interval)) * 2 * math.Pi // 0 at start of interval, 2*pi at end
				mean := (i.Min + i.Max) / 2
				result = mean + (((i.Max - i.Min) / 2) * math.Sin(intervalFactor))
			}
		default:
			{
				//return 0, fmt.Errorf("unknown function %q", i.Func)
				//TODO workaround for missing error flow. Case should be impossible (except when it happens ;) )
				panic(fmt.Sprintf("unknown function %q", i.Func))
			}
		}
	}

	return result, nil
}

// Build a *Collection from a file
func FromYamlFile(filename string) (*Collection, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	c := Collection{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal %v: %v", filename, err)
	}

	var validationErrors []string

	if c.Version == "" {
		validationErrors = append(validationErrors, "missing version")
	}

	if len(c.Metrics) == 0 {
		validationErrors = append(validationErrors, "metrics must have one or more elements")
	}

	for i := range c.Metrics {
		metric := c.Metrics[i]
		metric.parent = &c

		if !isInSlice(metric.Type, validMetricTypes) {
			validationErrors = append(validationErrors, fmt.Sprintf("metric %v: Unknown type %q. Must be one of gauge, counter, summary, histogram", metric.Name, metric.Type))
		}

		if len(metric.Items) == 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("metric %v: Must have at least one metricitem", metric.Name))
		} else {
			for j := range metric.Items {
				item := metric.Items[j]
				item.parent = metric

				if item.Min > item.Max {
					validationErrors = append(validationErrors, fmt.Sprintf("metric %v: min > max", c.Metrics[i].Name))
				}

				if !isInSlice(item.Func, validFunctions) {
					validationErrors = append(validationErrors, fmt.Sprintf("metric %v: Unknown func %q. Must be one of rand, asc, desc, sin", c.Metrics[i].Name, item.Func))
				}
				if item.Interval == 0 {
					validationErrors = append(validationErrors, fmt.Sprintf("metric %v: Invalid interval. Must be 1s or longer", c.Metrics[i].Name))
				}
				var keys []string
				for _, key := range reflect.ValueOf(item.Labels).MapKeys() {
					keys = append(keys, key.String())
				}
				sort.Strings(keys)
				sort.Strings(metric.Labels)
				if !stringSlicesEqual(keys, metric.Labels) {
					validationErrors = append(validationErrors, fmt.Sprintf("metric %v: Label mismatch. Item=%v, matric=%v", c.Metrics[i].Name, keys, metric.Labels))
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		errorMessage := ""
		for i := 0; i < len(validationErrors); i++ {
			errorMessage += fmt.Sprintf("%v. %v", i+1, validationErrors[i]) + "; "
		}
		return nil, fmt.Errorf("input has %v validation errors: %v", len(validationErrors), errorMessage)
	}

	return &c, nil
}
