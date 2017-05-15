package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vseledkin/gorpora"
)

const (
	normalizeHtmlEntities = "normalize.html.entities"
	split                 = "split"
	filterLanguage        = "filter.language"
)

var MAX int

type arrayFlags []string

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func (i *arrayFlags) String() string {
	return "my string representation"
}

var languages arrayFlags

func main() {
	normalizeHtmlEntitiesCommand := flag.NewFlagSet(normalizeHtmlEntities, flag.ExitOnError)
	normalizeHtmlEntitiesCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")

	splitCommand := flag.NewFlagSet(split, flag.ExitOnError)
	splitCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")

	filterLanguageCommand := flag.NewFlagSet(filterLanguage, flag.ExitOnError)
	filterLanguageCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")
	filterLanguageCommand.Var(&languages, "lang", "set of accepted languages")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "gorpora <command> arguments\n")
		fmt.Fprintf(os.Stderr, "commands are:\n")

		fmt.Fprintf(os.Stderr, "%s\n", normalizeHtmlEntities)
		normalizeHtmlEntitiesCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", split)
		splitCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", filterLanguage)
		filterLanguageCommand.PrintDefaults()

		flag.PrintDefaults()
	}
	flag.Parse()
	log.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case normalizeHtmlEntities:
		normalizeHtmlEntitiesCommand.Parse(os.Args[2:])

	case split:
		splitCommand.Parse(os.Args[2:])

	case filterLanguage:
		filterLanguageCommand.Parse(os.Args[2:])

	default:
		log.Printf("%q is not valid command.\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}

	// NORMALIZE ENTITIES COMMAND ISSUED
	if normalizeHtmlEntitiesCommand.Parsed() {
		gorpora.NormalizeHtmlEntities()
		return
	}

	// SPLIT COMMAND ISSUED
	if splitCommand.Parsed() {
		gorpora.Split()
		return
	}

	// FILTER LANGUAGES COMMAND ISSUED
	if filterLanguageCommand.Parsed() {
		if len(languages) == 0 {
			log.Printf("no -lang parameters given\n")
			flag.Usage()
		}
		gorpora.FilterLanguage(languages)
		return
	}
}
