package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vseledkin/gorpora"
	"github.com/vseledkin/gorpora/fb2"
)

const (
	stripHtml             = "strip.html"
	normalizeHtmlEntities = "normalize.html.entities"
	tokenize              = "word.tokenizer"
	unique                = "unique"
	filterLanguage        = "filter.language"
	sentences             = "sentence.tokenizer"
	fb2text               = "fb2text"
	collect               = "collect"
)

type arrayFlags []string

var (
	MAX_LEN            int
	MIN_LEN            int
	MAX_COLLECT_LEN    int
	MIN_COLLECT_LEN    int
	DEBUG              bool
	languages          arrayFlags
	LEMMAS             bool
	UDPIPE             bool
	COLLECT_INPUT      string
	INPUT              string
	THREADS            int
	OUTPUT_LINE_ENDING int
	EXTENSION          string
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

	collectCommand := flag.NewFlagSet(collect, flag.ExitOnError)
	collectCommand.StringVar(&COLLECT_INPUT, "i", ".", "directory with files, will be processed recursively")
	collectCommand.StringVar(&EXTENSION, "e", "txt", "extension of accepted files")
	collectCommand.IntVar(&MIN_COLLECT_LEN, "min", 1, "minimun line length expressed in utf8 chars to be accepted for output")
	collectCommand.IntVar(&MAX_COLLECT_LEN, "max", 1000000, "maximum line length expressed in utf8 chars to be accepted for output")

	fb2textCommand := flag.NewFlagSet(fb2text, flag.ExitOnError)
	fb2textCommand.StringVar(&INPUT, "i", "", "directory with fb2 files, will be processed recursively")
	fb2textCommand.IntVar(&THREADS, "t", 1, "number of threads for parallel processing of conversion jobs")
	fb2textCommand.IntVar(&OUTPUT_LINE_ENDING, "l", 1, "number of \n added after each text output block (paragraph)")

	normalizeHtmlEntitiesCommand := flag.NewFlagSet(normalizeHtmlEntities, flag.ExitOnError)
	normalizeHtmlEntitiesCommand.IntVar(&MAX_LEN, "max", 0, "maximum number of lines to process")
	normalizeHtmlEntitiesCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

	stripHtmlCommand := flag.NewFlagSet(stripHtml, flag.ExitOnError)

	tokenizeCommand := flag.NewFlagSet(tokenize, flag.ExitOnError)
	tokenizeCommand.BoolVar(&UDPIPE, "udpipe", false, "use Udpipe as tokenizer")
	tokenizeCommand.BoolVar(&LEMMAS, "lemma", false, "output lemmas instead of words")
	tokenizeCommand.BoolVar(&DEBUG, "debug", false, "do nothing only print use cases")

	sentenceCommand := flag.NewFlagSet(sentences, flag.ExitOnError)
	sentenceCommand.IntVar(&MAX_LEN, "max", 1000000, "maximum sentence length in chars")
	sentenceCommand.IntVar(&MIN_LEN, "min", 10, "minimun sentence length in chars")
	sentenceCommand.BoolVar(&DEBUG, "debug", false, "do nothing only print use cases")

	uniqueCommand := flag.NewFlagSet(unique, flag.ExitOnError)
	uniqueCommand.BoolVar(&DEBUG, "debug", false, "do nothing only print use cases")

	filterLanguageCommand := flag.NewFlagSet(filterLanguage, flag.ExitOnError)
	filterLanguageCommand.Var(&languages, "lang", "set of accepted languages")
	filterLanguageCommand.BoolVar(&DEBUG, "debug", false, "do othing only print use cases")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "gorpora <command> arguments\n")
		fmt.Fprintf(os.Stderr, "commands are:\n")

		fmt.Fprintf(os.Stderr, "%s\n", fb2text)
		fb2textCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", collect)
		collectCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", normalizeHtmlEntities)
		normalizeHtmlEntitiesCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", stripHtml)
		stripHtmlCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", tokenize)
		tokenizeCommand.PrintDefaults()

		fmt.Fprintf(os.Stderr, "%s\n", sentences)
		sentenceCommand.PrintDefaults()

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
	case collect:
		collectCommand.Parse(os.Args[2:])

	case fb2text:
		fb2textCommand.Parse(os.Args[2:])

	case normalizeHtmlEntities:
		normalizeHtmlEntitiesCommand.Parse(os.Args[2:])

	case stripHtml:
		stripHtmlCommand.Parse(os.Args[2:])

	case tokenize:
		tokenizeCommand.Parse(os.Args[2:])

	case sentences:
		sentenceCommand.Parse(os.Args[2:])

	case filterLanguage:
		filterLanguageCommand.Parse(os.Args[2:])

	case unique:
		uniqueCommand.Parse(os.Args[2:])

	default:
		log.Printf("%q is not valid command.\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}

	// collect tool
	if collectCommand.Parsed() {
		gorpora.Collect(MIN_COLLECT_LEN, MAX_COLLECT_LEN, COLLECT_INPUT, EXTENSION, 0)
		return
	}

	// fb2 convert tot text
	if fb2textCommand.Parsed() {
		fb2.ConvertFB2text(INPUT, OUTPUT_LINE_ENDING, THREADS)
		return
	}

	// NORMALIZE ENTITIES COMMAND ISSUED
	if normalizeHtmlEntitiesCommand.Parsed() {
		gorpora.NormalizeHtmlEntities()
		return
	}
	// STRIP HTML ENTITIES COMMAND ISSUED
	if stripHtmlCommand.Parsed() {
		gorpora.StripHtml()
		return
	}

	// SPLIT COMMAND ISSUED
	if tokenizeCommand.Parsed() {
		gorpora.Split(UDPIPE, LEMMAS)
		return
	}

	// SENTENCE COMMAND ISSUED
	if sentenceCommand.Parsed() {
		gorpora.Sentesize(MIN_LEN, MAX_LEN)
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
