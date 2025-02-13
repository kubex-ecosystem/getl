package main

import (
	"github.com/faelmori/getl/cmd"
)

func main() {
	rootCmd := cmd.RegX()
	_ = rootCmd.Execute(nil)
}
