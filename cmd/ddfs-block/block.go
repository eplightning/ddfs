package main

import (
	"math/rand"
	"time"

	"git.eplight.org/eplightning/ddfs/cmd/ddfs-block/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
