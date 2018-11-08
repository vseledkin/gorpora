package main

import (
	"log"

	"github.com/vseledkin/gorpora/cld2"
	"github.com/vseledkin/gorpora/collect"

	"github.com/spf13/cobra"
	"github.com/vseledkin/gorpora/dic"
	"github.com/vseledkin/gorpora/embed"
	"github.com/vseledkin/gorpora/interleave"
	"github.com/vseledkin/gorpora/tokenizer"
)

func main() {
	var rootCmd = &cobra.Command{Use: "gorpora"}
	rootCmd.AddCommand(collect.CollectCommand)
	rootCmd.AddCommand(cld2.LanguageCommand)
	rootCmd.AddCommand(dic.DictionaryCommand)
	rootCmd.AddCommand(embed.EmbedCommand)
	rootCmd.AddCommand(embed.NNCommand)
	rootCmd.AddCommand(interleave.Command)
	rootCmd.AddCommand(tokenizer.WordTokenizerCommand)
	if e := rootCmd.Execute(); e != nil {
		log.Fatal(e)
	}
}
