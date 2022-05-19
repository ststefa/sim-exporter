package cmd

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"git.mgmt.innovo-cloud.de/obs/sim-exporter/pkg/errors"
	"git.mgmt.innovo-cloud.de/obs/sim-exporter/pkg/metrics"
)

var (
	port_help = "TCP port on which the exporter should listen"
	port      = 8080

	path_help = "URI (with leading '/') on which the exporter should listen"
	path      = "/metrics"

	refreshTime_help = "After how many seconds the metrics are refreshed"
	refreshTime, _   = time.ParseDuration("15s")

	serveCmd = &cobra.Command{
		Use:     "serve <file.yaml>",
		Short:   "Serve simulated prometheus metrics defined in <file.yaml>",
		Long:    "Start the exporter and serve prometheus metrics read from <file.yaml> to the configured port and path.",
		Args:    cobra.ExactArgs(1),
		PreRunE: validateServe,
		Run:     doServe,
	}
)

func init() {
	serveCmd.PersistentFlags().IntVarP(&port, "port", "p", port, port_help)
	serveCmd.PersistentFlags().StringVar(&path, "path", path, path_help)
	serveCmd.PersistentFlags().DurationVarP(&refreshTime, "refresh", "r", refreshTime, refreshTime_help)

	rootCmd.AddCommand(serveCmd)
}

func validateServe(cmd *cobra.Command, args []string) error {

	// Validate path
	re, err := regexp.Compile(`^(/.+)+`)
	if err != nil {
		return err
	}

	if !re.MatchString(path) {
		return fmt.Errorf("path %q does not match required regex %q", path, re.String())
	}

	// Validate port
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q. Must be in range 1-65535", strconv.FormatInt(int64(port), 10))
	}

	return nil
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doServe(cmd *cobra.Command, args []string) {
	helpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, "<html><body>\n")
		io.WriteString(w, "<h1>Available paths</h1>\n")
		io.WriteString(w, "<a href='"+path+"'>"+path+"</a>\n")
		io.WriteString(w, "</body></html>\n")
	})

	collection, err := metrics.FromYamlFile(args[0])
	if err != nil {
		panic(&errors.SimulationError{Err: err.Error()})
	}

	err = metrics.SetupMetricsCollection(collection)
	if err != nil {
		panic(&errors.SimulationError{Err: err.Error()})
	}
	metrics.StartMetricsCollection(collection, time.Duration(refreshTime))

	http.Handle("/", helpHandler)
	http.Handle(path, promhttp.Handler())

	log.Printf("Serving metrics on *:%d%v", port, path)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
