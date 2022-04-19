package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/cmd"
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
		var simErr *cmd.SimulationError
		if errors.As(r.(error), &simErr) {
			fmt.Fprintf(os.Stderr, "Simulation error: %v\n", r)
		} else {
			log.Errorf("unexpected %s: %q\n", reflect.TypeOf(r).Elem(), r)
		}
	}
}

func main() {
	defer panicExit()
	cmd.Execute()
	os.Exit(0)
}
