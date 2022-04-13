package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"example.com/sim-exporter/cmd"
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
			log.Errorf("unexpected %s: '%v'\n", reflect.TypeOf(r).Elem(), r)
		}
	}
}

func main() {
	defer panicExit()
	cmd.Execute()
	os.Exit(0)
}
