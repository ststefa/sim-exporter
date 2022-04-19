package cmd

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	// Default listening port
	port = 8080

	// Default metrics URI
	path = "/metrics"

	// Default metrics refresh time
	refreshTime, _ = time.ParseDuration("15s")

	serveCmd = &cobra.Command{
		Use:   "serve <file.yaml>",
		Short: "Serve simulated prometheus metrics defined in <file.yaml>",
		Long:  "Start the exporter and serve prometheus metrics read from <file.yaml> to the configured port and path.",
		Args:  cobra.ExactArgs(1),
		Run:   doServe,
	}
)

func init() {
	serveCmd.PersistentFlags().IntVar(&port, "port", port, "TCP port on which the exporter should listen")
	serveCmd.PersistentFlags().StringVar(&path, "path", path, "URI (with leading '/') on which the exporter should listen")
	serveCmd.PersistentFlags().DurationVar(&refreshTime, "refresh", refreshTime, "After how many seconds the metrics are refreshed")

	rootCmd.AddCommand(serveCmd)
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doServe(cmd *cobra.Command, args []string) {
	helpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "Available paths:\n")
		io.WriteString(w, path+"\n")
	})

	config, err := loadAndValidateConfiguration(args[0])
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	err = setupMetricsCollection(config)
	if err != nil {
		panic(&SimulationError{err.Error()})
	}
	startMetricsCollection(config, time.Duration(refreshTime))

	http.Handle("/", helpHandler)
	http.Handle(path, promhttp.Handler())

	log.Printf("Serving metrics on *:%d%v", port, path)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

}
