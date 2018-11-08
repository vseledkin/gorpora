package tokenizer

import (
	"github.com/spf13/cobra"
	"log"
	"bufio"
	"os"
	"strings"
	"unicode"
	"fmt"
	"io"
)

var WordTokenizerCommand *cobra.Command

var inputFilePath1, inputFilePath2 *string

func init() {

	WordTokenizerCommand = &cobra.Command{
		Use:   "word.tokenizer",
		Short: "split text by space and punctuation, punctuation is preserved",
		Long:  "split text by space and punctuation, punctuation is preserved ",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Word Tokenizer:")
			//log.Printf("\tInput 1: %s\n", *inputFilePath1)
			//log.Printf("\tInput 2: %s\n", *inputFilePath2)
			if e := SimpleTokenizer(); e != nil {
				log.Fatal(e)
			}
		},
	}

	//inputFilePath1 = Command.Flags().StringP("input1", "1", "", "text file path 1")
	//inputFilePath2 = Command.Flags().StringP("input2", "2", "", "text file path 2")
	//Command.MarkFlagRequired("input1")
	//Command.MarkFlagRequired("input2")
}

func SimpleTokenizer() (e error) {
	reader := bufio.NewReader(os.Stdin)
	var line string
	for {
		line, e = reader.ReadString('\n')
		if e != nil && e==io.EOF {
			e = nil
			break
		}
		tokens := split2Tokens(line)
		if len(tokens) == 0 {
			continue
		}
		os.Stdout.WriteString(tokens)
		os.Stdout.WriteString("\n")
	}
	return
}

func split2Tokens(s string) string {
	token := ""
	var split []string
	for _, r := range s {
		switch {
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			if len(token) > 0 {
				split = append(split, token)
				token = ""
			}
			split = append(split, string(r))
		case len(token) == 0 && unicode.IsSpace(r):
			continue // skip leading space
		case len(token) == 0 && !unicode.IsSpace(r):
			token = string(r)
		case len(token) > 0 && !unicode.IsSpace(r):
			token += string(r)
		case len(token) > 0 && unicode.IsSpace(r):
			split = append(split, token)
			token = ""
		default:
			panic(fmt.Errorf("unknown symbol %q", r))
		}
	}
	if len(token) > 0 {
		split = append(split, token)
	}
	return strings.Join(split, " ")
}
