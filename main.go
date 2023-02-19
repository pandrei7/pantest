package main

import (
	"github.com/alecthomas/kong"
)

func main() {
	var params Params
	kong.Parse(&params)
	runCli(params.ConfigFile)
}
