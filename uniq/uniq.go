package uniq

import (
	"github.com/spf13/cobra"
	"log"
	"crypto/md5"
	"encoding/hex"
	"bufio"
	"os"
)

var UniqCommand *cobra.Command

var input *string

func init() {

	UniqCommand = &cobra.Command{
		Use:   "uniq",
		Short: "reads lines from stdin, outputs unique lines to stdout",
		Long:  "reads lines from stdin, outputs unique lines to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Uniq:")
			log.Printf("\tInput: %s\n", "stdin")
			log.Printf("\tOutput: %s\n", "stdout")
			Unique(false)

		},
	}
}

func GetMD5Hash(bytes []byte) string {
	hasher := md5.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil))
}

func Unique(DEBUG bool) {
	reader := bufio.NewReader(os.Stdin)
	dic := make(map[string]int)
	lineCount := 0
	uniqueCount := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lineCount++
		hash := GetMD5Hash([]byte(line))

		if _, ok := dic[hash]; ok {
			if DEBUG {
				os.Stdout.WriteString("DUBLICATE: " + line)
			}
			dic[hash] += 1
		} else {
			dic[hash] = 1
			if !DEBUG {
				os.Stdout.WriteString(line)
				uniqueCount++
			}
		}
		if lineCount%10e6 == 0 {
			log.Printf("clean: dic size %d %d total", len(dic), lineCount)
			for k, v := range dic {
				if v < 2 {
					delete(dic, k)
				} else {
					dic[k]--
				}
			}
			log.Printf("dic size %d %d total", len(dic), lineCount)
		}
	}

	log.Println(lineCount, "lines total")
	log.Println(uniqueCount, "unique lines")
	log.Println(lineCount-uniqueCount, "non unique lines")
}