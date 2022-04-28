package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Create and setup metrics and collection
func SetupMetricsCollection(config *Collection) error {

	for i := range config.Metrics {
		metric := config.Metrics[i]

		switch metric.Type {
		case "gauge":
			vec := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: metric.Name,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.gauge = vec
			prometheus.MustRegister(vec)
		case "counter":
			vec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: metric.Name,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.counter = vec
			prometheus.MustRegister(vec)
		case "summary":
			vec := prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: metric.Name,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.summary = vec
			prometheus.MustRegister(vec)

		case "histogram":
			vec := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: metric.Name,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.histogram = vec
			prometheus.MustRegister(vec)
		default:
			return fmt.Errorf("metric %v: type %q not defined", i, metric.Type)
		}

		config.Metrics[i] = metric

	}
	return nil
}

// Start async metrics refresh in intervals
func StartMetricsCollection(c *Collection, refresh time.Duration) {
	startTime := time.Now()
	go func() {
		for {
			go func() error {
				err := refreshMetricsCollection(c, startTime)
				if err != nil {
					return err
				}
				return nil
			}()
			time.Sleep(refresh)
		}
	}()
}

func refreshMetricsCollection(c *Collection, startTime time.Time) error {
	var wg sync.WaitGroup

	callbackChannel := make(chan func() error)

	for i := range c.Metrics {
		metric := c.Metrics[i]

		for _, metricItem := range metric.Items {
			newVal, _ := metricItem.generateValue(startTime)
			//TODO error handling not working. Should not abort refresh process
			//if err != nil {
			//	return err
			//}

			switch metric.Type {
			case "gauge":
				metric.prometheus.gauge.With(metricItem.Labels).Set(newVal)
			case "summary":
				metric.prometheus.summary.With(metricItem.Labels).Observe(newVal)
			case "histogram":
				metric.prometheus.histogram.With(metricItem.Labels).Observe(newVal)
			case "counter":
				metric.prometheus.counter.With(metricItem.Labels).Add(newVal)
			}
		}
	}

	go func() {
		var callbackList []func() error
		for callback := range callbackChannel {
			callbackList = append(callbackList, callback)
		}

		for _, callback := range callbackList {
			err := callback()
			if err != nil {
				fmt.Printf("error: %v", err)
			}

		}

		log.Debug("run: finished")
	}()

	wg.Wait()
	close(callbackChannel)

	return nil
}
