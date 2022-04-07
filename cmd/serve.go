package cmd

import (
	"io"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags().IntVar(&port, "port", port, "TCP port on which the exporter should listen")
	serveCmd.PersistentFlags().StringVar(&metricsPath, "metrics_path", metricsPath, "TCP metrics_path on which the exporter should listen")

	rootCmd.AddCommand(serveCmd)
}

func serve(cmd *cobra.Command, args []string) {
	helpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "Available paths:\n")
		io.WriteString(w, "/"+metricsPath+"\n")
	})
	http.Handle("/", helpHandler)

	http.Handle(metricsPath, promhttp.Handler())
	log.Printf("Serving metrics on *:%d/%v", port, metricsPath)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve simulated metrics",
	Long:  "Start the exporter and serve metrics on the configured port and path.",
	Run:   serve,
}
