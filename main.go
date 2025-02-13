package main

import (
	"fmt"
	"github.com/faelmori/getl/cli"
)

func main() {
	rootCmd := cli.RegX()
	execErr := rootCmd.Execute()
	if execErr != nil {
		panic(fmt.Sprintf("Error executing command: %s", execErr.Error()))
		return
	}
}
