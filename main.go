package main

import (
	"example.com/sim-exporter/cmd"
	"fmt"
	"reflect"
)

func panicExit() {
	if r := recover(); r != nil {
		fmt.Printf("unexpected %s: '%v'\n", reflect.TypeOf(r).Elem(), r)
	}
}

func main() {
	defer panicExit()
	cmd.Execute()
}
