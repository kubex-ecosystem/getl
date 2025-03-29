package main

import (
	"fmt"
	"github.com/faelmori/logz"
	"os"
)

var l = logz.GetLogger("Getl")

func main() {
	if err := RegX().Execute(nil); err != nil {
		l.Error(fmt.Sprintf("Error: %v", err), map[string]interface{}{})
		os.Exit(1)
	}
}
