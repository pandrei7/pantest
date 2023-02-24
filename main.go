package main

import (
	"github.com/alecthomas/kong"
)

func main() {
	var cli Cli
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
