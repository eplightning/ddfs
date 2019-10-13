package main

import (
	"math/rand"
	"time"

	"github.com/eplightning/ddfs/cmd/ddfs-cli/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
