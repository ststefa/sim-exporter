package main

import (
	"fmt"
	"os"
	"reflect"

	"example.com/sim-exporter/cmd"
)

func panicExit() {
	if r := recover(); r != nil {
		fmt.Printf("unexpected %s: '%v'\n", reflect.TypeOf(r).Elem(), r)
	}
}

func main() {
	defer panicExit()
	cmd.Execute()
	os.Exit(0)
}
