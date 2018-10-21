// Package cld2 implements language detection using the
// Compact Language Detector.
//
// This package includes the relevant sources from the cld2
// project, so it doesn't require any external dependencies.
// For more information about CLD2, see https://code.google.com/p/cld2/.
package cld2

// #include <stdlib.h>
// #include "cld2.h"
import "C"
import (
	"bufio"
	"log"
	"os"
	"unsafe"

	"github.com/spf13/cobra"
)

var LanguageCommand *cobra.Command

// Detect returns the language code for detected language
// in the given text.
func Detect(text string) string {
	cs := C.CString(text)
	res := C.DetectLang(cs, -1)
	C.free(unsafe.Pointer(cs))
	var lang string
	if res != nil {
		lang = C.GoString(res)
	}
	return lang
}

var language *[]string

func init() {

	LanguageCommand = &cobra.Command{
		Use:   "language",
		Short: "collect text of some language",
		Long:  "collect text of some language",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Language:")
			log.Printf("\tLanguages: %s\n", language)

			filterLanguage(*(language))
		},
	}

	language = LanguageCommand.Flags().StringArrayP("language", "l", nil, "diirectry containing text files, processed recursively")
	LanguageCommand.MarkFlagRequired("language")
}

func filterLanguage(languages []string) {
	var collected, total int
	filtered := make(map[string]int)
	reader := bufio.NewReader(os.Stdin)
	var accepted bool
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		language := Detect(line)
		accepted = false
		total++
		for _, lang := range languages {
			if language == lang {
				accepted = true
				break
			}
		}
		if accepted {
			os.Stdout.WriteString(line)
			collected++
		} else {
			filtered[language]++
		}
		if total%100001 == 100000 {
			log.Printf("total: %d collected: %d", total, collected)
			if total%1000001 == 1000000 {
				for key, value := range filtered {
					log.Printf("\tlang: %s count: %d", key, value)
				}
			}
		}
	}
	os.Stdout.Sync()
}
