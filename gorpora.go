package main

import (
	"log"

	"github.com/vseledkin/gorpora/cld2"
	"github.com/vseledkin/gorpora/collect"

	"github.com/spf13/cobra"
	"github.com/vseledkin/gorpora/dic"
	"github.com/vseledkin/gorpora/embed"
	"github.com/vseledkin/gorpora/interleave"
)

func main() {
	var rootCmd = &cobra.Command{Use: "gorpora"}
	rootCmd.AddCommand(collect.CollectCommand)
	rootCmd.AddCommand(cld2.LanguageCommand)
	rootCmd.AddCommand(dic.DictionaryCommand)
	rootCmd.AddCommand(embed.EmbedCommand)
	rootCmd.AddCommand(interleave.Command)
	if e := rootCmd.Execute(); e != nil {
		log.Fatal(e)
	}
}
