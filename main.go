package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/setsal/go-AMF/commands"
)

// Name :
const Name string = "go-AMF"

var version = "0.0.0"

func main() {
	var v = flag.Bool("v", false, "Show version")

	if envArgs := os.Getenv("GOVAL_DICTIONARY_ARGS"); 0 < len(envArgs) {
		flag.CommandLine.Parse(strings.Fields(envArgs))
	} else {
		flag.Parse()
	}

	if *v {
		fmt.Printf("go-AMF %s \n", version)
		os.Exit(0)
	}

	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
