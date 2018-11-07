package interleave

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"bufio"
)

var Command *cobra.Command

var inputFilePath1, inputFilePath2 *string

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
			if e := Interleave(*inputFilePath1, *inputFilePath2); e != nil {
				log.Fatal(e)
			}
		},
	}

	inputFilePath1 = Command.Flags().StringP("input1", "1", "", "text file path 1")
	inputFilePath2 = Command.Flags().StringP("input2", "2", "", "text file path 2")
	Command.MarkFlagRequired("input1")
	Command.MarkFlagRequired("input2")
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
	for {
		if line1, e := reader1.ReadString('\n'); e != nil {
			break
		} else {
			if line2, e := reader2.ReadString('\n'); e != nil {
				break
			} else {
				tokens1 := strings.Fields(line1)
				tokens2 := strings.Fields(line2)
				//log.Printf("%+v %+v", tokens1, tokens2)
				// write head
				var i int
				for ; i < len(tokens1) && i < len(tokens2); i++ {
					if i > 0 {
						os.Stdout.WriteString(" ")
					}
					os.Stdout.WriteString(tokens1[i])
					os.Stdout.WriteString(" ")
					os.Stdout.WriteString(tokens2[i])
				}
				//write tails
				for ; i < len(tokens1); i++ {
					os.Stdout.WriteString(" ")
					os.Stdout.WriteString(tokens1[i])
				}
				//write tails
				for ; i < len(tokens2); i++ {
					os.Stdout.WriteString(" ")
					os.Stdout.WriteString(tokens2[i])
				}
				os.Stdout.WriteString("\n")
			}
		}
	}
	return
}
