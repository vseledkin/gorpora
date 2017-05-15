package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vseledkin/gorpora"
)

const (
	normalizeHtmlEntities = "normalize.html.etities"
)

func main() {
	normalizeHtmlEntitiesCommand := flag.NewFlagSet(normalizeHtmlEntities, flag.ExitOnError)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "gorpora <command> arguments\n")
		fmt.Fprintf(os.Stderr, "commands are:\n")
		fmt.Fprintf(os.Stderr, "%s\n", normalizeHtmlEntities)
		normalizeHtmlEntitiesCommand.PrintDefaults()
		flag.PrintDefaults()
	}
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch os.Args[1] {
	case normalizeHtmlEntities:
		normalizeHtmlEntitiesCommand.Parse(os.Args[:0])
	default:
		log.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(1)
	}

	// BUILD COMMAND ISSUED
	if normalizeHtmlEntitiesCommand.Parsed() {
		gorpora.NormalizeHtmlEntities()
		return
	}
}
