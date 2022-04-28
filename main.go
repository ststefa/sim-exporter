package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/cmd"
	simerrors "git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/pkg/errors"

	"github.com/sirupsen/logrus"
)

var (
	log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
)

func panicExit() {
	if r := recover(); r != nil {
		var simErr *simerrors.SimulationError
		if errors.As(r.(error), &simErr) {
			fmt.Fprintf(os.Stderr, "Simulation error: %v\n", r)
		} else {
			fmt.Fprintf(os.Stderr, "unexpected %s: %q\n", reflect.TypeOf(r).Elem(), r)
		}
		os.Exit(1)
	}
}

func main() {
	defer panicExit()
	cmd.Execute()
	os.Exit(0)
}
