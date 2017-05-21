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
	unique                = "unique"
	filterLanguage        = "filter.language"
)

type arrayFlags []string

var (
	MAX       int
	DEBUG     bool
	languages arrayFlags
)

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func (i *arrayFlags) String() string {
	return "my string representation"
}

func main() {
	log.SetOutput(os.Stderr)
	normalizeHtmlEntitiesCommand := flag.NewFlagSet(normalizeHtmlEntities, flag.ExitOnError)
	normalizeHtmlEntitiesCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")
	normalizeHtmlEntitiesCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

	splitCommand := flag.NewFlagSet(split, flag.ExitOnError)
	splitCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")
	splitCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

	uniqueCommand := flag.NewFlagSet(unique, flag.ExitOnError)
	uniqueCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")
	uniqueCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

	filterLanguageCommand := flag.NewFlagSet(filterLanguage, flag.ExitOnError)
	filterLanguageCommand.IntVar(&MAX, "max", 0, "maximum number of lines to process")
	filterLanguageCommand.Var(&languages, "lang", "set of accepted languages")
	filterLanguageCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

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

		fmt.Fprintf(os.Stderr, "%s\n", unique)
		uniqueCommand.PrintDefaults()

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

	case unique:
		uniqueCommand.Parse(os.Args[2:])

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

	// UNIQUE COMMAND ISSUED
	if uniqueCommand.Parsed() {
		gorpora.Unique(DEBUG)
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
