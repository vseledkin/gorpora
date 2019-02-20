package interleave

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"bufio"
	"github.com/vseledkin/gorpora/tokenizer"
)

var Command *cobra.Command

var inputFilePath1, inputFilePath2, delimiter *string
var skipEquals, wordTokenizer *bool

func init() {

	Command = &cobra.Command{
		Use:   "interleave",
		Short: "mix content from two test files",
		Long:  "mix content from two test files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Interleave:")
			log.Printf("\tInput 1: %s\n", *inputFilePath1)
			log.Printf("\tInput 2: %s\n", *inputFilePath2)
			log.Printf("\tDelimiter: %s\n", *delimiter)
			if len(*inputFilePath2) == 0 && len(*delimiter) == 0 {
				log.Fatal("set input2 or delimiter for input1")
			}
			if len(*delimiter) == 0 {
				if e := Interleave(*inputFilePath1, *inputFilePath2); e != nil {
					log.Fatal(e)
				}
			} else {
				if e := InterleaveWithDelimiter(*inputFilePath1, *delimiter); e != nil {
					log.Fatal(e)
				}
			}
		},
	}

	inputFilePath1 = Command.Flags().StringP("input1", "1", "", "text file path 1, or stdin '-' if using interleave with delimiter")
	inputFilePath2 = Command.Flags().StringP("input2", "2", "", "text file path 2")
	delimiter = Command.Flags().StringP("delimiter", "d", "", "line delimiter if reading from single file in line1delimiterline2 format")
	skipEquals = Command.Flags().BoolP("skip.equals", "s", true, "skip pairs with equal lines")
	wordTokenizer = Command.Flags().BoolP("word.tokenizer", "t", false, "apply word tokenizer to lines")
	Command.MarkFlagRequired("input1")
}

func Interleave(f1, f2 string) (e error) {
	var file1, file2 *os.File
	if file1, e = os.Open(f1); e != nil {
		log.Fatal(e)
	}
	defer file1.Close()
	if file2, e = os.Open(f2); e != nil {
		log.Fatal(e)
	}
	defer file2.Close()

	reader1 := bufio.NewReader(file1)
	reader2 := bufio.NewReader(file2)
	parser := strings.Fields
	if *wordTokenizer {
		parser = tokenizer.Split2Tokens
	}
	for {
		if line1, e := reader1.ReadString('\n'); e != nil {
			break
		} else {
			if line2, e := reader2.ReadString('\n'); e != nil {
				break
			} else {
				line1 = strings.TrimSpace(line1)
				line2 = strings.TrimSpace(line2)
				if len(line1) > 0 && len(line2) > 0 {
					if *skipEquals {
						if line1 == line2 {
							continue
						} else {
							tokens1 := parser(line1)
							tokens2 := parser(line2)
							interleave(tokens1, tokens2)
						}
					} else {
						tokens1 := parser(line1)
						tokens2 := parser(line2)
						interleave(tokens1, tokens2)
					}
				}
			}
		}
	}
	return
}

func InterleaveWithDelimiter(f1 string, d string) (e error) {

	var reader1 *bufio.Reader
	if f1 == "-" {
		reader1 = bufio.NewReader(os.Stdin)
	} else {
		var file1 *os.File
		if file1, e = os.Open(f1); e != nil {
			return e
		}
		defer file1.Close()
		reader1 = bufio.NewReader(file1)
	}
	parser := strings.Fields
	if *wordTokenizer {
		parser = tokenizer.Split2Tokens
	}
	for {
		if line, e := reader1.ReadString('\n'); e != nil {
			break
		} else {
			parts := strings.Split(line, *delimiter)
			if len(parts) != 2 {
				//log.Printf("got %d lines after applying delimiter '%s' to line '%s'", len(parts), d, line)
				continue
			}
			line1, line2 := parts[0], parts[1]
			line1 = strings.TrimSpace(line1)
			line2 = strings.TrimSpace(line2)
			if len(line1) > 0 && len(line2) > 0 {
				if *skipEquals {
					if line1 == line2 {
						continue
					} else {
						tokens1 := parser(line1)
						tokens2 := parser(line2)
						interleave(tokens1, tokens2)
					}
				} else {
					tokens1 := parser(line1)
					tokens2 := parser(line2)
					interleave(tokens1, tokens2)
				}
			}
		}
	}
	return
}

func interleave(tokens1, tokens2 []string) {
	var i int
	prev := "∞"
	first := true
	for ; i < len(tokens1) && i < len(tokens2); i++ {
		if tokens1[i] != prev {
			if !first {
				os.Stdout.WriteString(" ")
			} else {
				first = false
			}
			os.Stdout.WriteString(tokens1[i])
			prev = tokens1[i]
		}
		if tokens2[i] != prev {
			if !first {
				os.Stdout.WriteString(" ")
			} else {
				first = false
			}
			os.Stdout.WriteString(tokens2[i])
			prev = tokens2[i]
		}
	}
	//write tails
	for ; i < len(tokens1); i++ {
		if tokens1[i] != prev {
			if !first {
				os.Stdout.WriteString(" ")
			} else {
				first = false
			}
			os.Stdout.WriteString(tokens1[i])
			prev = tokens1[i]
		}
	}
	//write tails
	for ; i < len(tokens2); i++ {
		if tokens2[i] != prev {
			if !first {
				os.Stdout.WriteString(" ")
			} else {
				first = false
			}
			os.Stdout.WriteString(tokens2[i])
			prev = tokens2[i]
		}
	}
	os.Stdout.WriteString("\n")
}
