package main

import (
	"log"

	"github.com/vseledkin/gorpora/cld2"
	"github.com/vseledkin/gorpora/collect"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "gorpora"}
	rootCmd.AddCommand(collect.CollectCommand)
	rootCmd.AddCommand(cld2.LanguageCommand)
	if e := rootCmd.Execute(); e != nil {
		log.Fatal(e)
	}
}
