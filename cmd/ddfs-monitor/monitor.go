package main

import (
	"math/rand"
	"time"

	"github.com/eplightning/ddfs/cmd/ddfs-monitor/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
