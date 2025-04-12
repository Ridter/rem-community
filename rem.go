package main

import (
	"github.com/chainreactors/rem/cmd/cmd"
)

//go:generate protoc --go_out=paths=source_relative:. .\message\msg.proto
func main() {
	cmd.RUN()
}
