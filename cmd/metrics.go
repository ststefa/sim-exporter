package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Create and setup metrics and collection
func setupMetricsCollection(config *Configuration) error {

	for metricName := range config.Metrics {
		metric := config.Metrics[metricName]
		err := metric.Init()
		if err != nil {
			return err
		}

		switch metric.Type {
		case "gauge":
			vec := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: metricName,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.gauge = vec
			prometheus.MustRegister(vec)
		case "counter":
			vec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: metricName,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.counter = vec
			prometheus.MustRegister(vec)
		case "summary":
			vec := prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: metricName,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.summary = vec
			prometheus.MustRegister(vec)

		case "histogram":
			vec := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: metricName,
					Help: metric.Help,
				},
				metric.Labels,
			)

			metric.prometheus.histogram = vec
			prometheus.MustRegister(vec)
		default:
			return fmt.Errorf("metric %v: type '%v' not defined", metricName, metric.Type)
		}

		config.Metrics[metricName] = metric

	}
	return nil
}

// Start async metrics refresh in intervals
func startMetricsCollection(config *Configuration, refresh time.Duration) {
	go func() {
		for {
			go func() {
				refreshMetricsCollection(config)
			}()
			time.Sleep(refresh)
		}
	}()
}

func refreshMetricsCollection(config *Configuration) {
	var wg sync.WaitGroup

	callbackChannel := make(chan func())

	for metricName := range config.Metrics {
		metric := config.Metrics[metricName]

		for _, metricItem := range metric.Items {
			switch metric.Type {
			case "gauge":
				metric.prometheus.gauge.With(metricItem.Labels).Set(metricItem.GenerateValue())
			case "summary":
				metric.prometheus.summary.With(metricItem.Labels).Observe(metricItem.GenerateValue())
			case "histogram":
				metric.prometheus.histogram.With(metricItem.Labels).Observe(metricItem.GenerateValue())
			case "counter":
				metric.prometheus.counter.With(metricItem.Labels).Add(metricItem.GenerateValue())
			}
		}
	}

	go func() {
		var callbackList []func()
		for callback := range callbackChannel {
			callbackList = append(callbackList, callback)
		}

		for _, callback := range callbackList {
			callback()
		}

		log.Debug("run: finished")
	}()

	wg.Wait()
	close(callbackChannel)
}
