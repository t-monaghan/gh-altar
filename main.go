// Package main only contains the required logic for cobra.
package main

import (
	"github.com/t-monaghan/gh-altar/cmd"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.Execute(version, commit)
}
