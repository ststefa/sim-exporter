package cmd

import (
	"io"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	// Default listening port
	port = 9041

	// Default URI path to metrics
	path = "metrics"

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve simulated prometheus metrics",
		Long:  "Start the exporter and serve prometheus metrics on the configured port and path.",
		Args:  cobra.ExactArgs(1),
		Run:   serve,
	}
)

func init() {
	serveCmd.PersistentFlags().IntVar(&port, "port", port, "TCP port on which the exporter should listen")
	serveCmd.PersistentFlags().StringVar(&path, "path", path, "URI (without leading '/') on which the exporter should listen")

	rootCmd.AddCommand(serveCmd)
}

func serve(cmd *cobra.Command, args []string) {
	helpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "Available paths:\n")
		io.WriteString(w, "/"+path+"\n")
	})

	http.Handle("/", helpHandler)
	http.Handle(path, promhttp.Handler())

	log.Printf("Serving metrics on *:%d/%v", port, path)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
