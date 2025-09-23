package main

import (
	"fmt"
	"os"

	"github.com/kubex-ecosystem/getl/internal/module"
	"github.com/kubex-ecosystem/logz"
)

var l = logz.GetLogger("Getl")

func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		l.ErrorCtx(fmt.Sprintf("Error: %v", err), map[string]interface{}{})
		os.Exit(1)
	}
}
