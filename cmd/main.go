package main

import (
	"os"

	"github.com/kubex-ecosystem/getl/internal/module"
	"github.com/kubex-ecosystem/logz"
)

var l = logz.GetLoggerZ("Getl")

func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		l.Errorf("Error: %v", err)
		os.Exit(1)
	}
}
