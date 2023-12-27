package main

import "github.com/analog-substance/arsenic/internal/app/cmd"

var version = "v0.0.0"
var commit = "replace"

func main() {
	cmd.SetVersionInfo(version, commit)
	cmd.Execute()
}
