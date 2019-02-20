package dic

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var DictionaryCommand *cobra.Command

var input *string
var minCount *uint64
var external map[string]struct{}

func init() {

	DictionaryCommand = &cobra.Command{
		Use:   "dic",
		Short: "make corpus dictionary",
		Long:  "collect corpus of space separated tokens from text file",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Dic:")
			log.Printf("\tInput: %s\n", *input)
			var e error
			if len(*input) > 0 {
				if external, e = loadExternal(*input); e!=nil{
					log.Fatal(e)
				}
			}

			if e := Dic(external); e != nil {
				log.Fatal(e)
			}
			//switch {
			//case len(*input) > 0:
			//	if e := Dic(); e != nil {
			//		log.Fatal(e)
			//	}
			//default:
			//	cmd.Usage()
			//}
		},
	}

	input = DictionaryCommand.Flags().StringP("input", "i", "", "external vocabulary")
	minCount = DictionaryCommand.Flags().Uint64P("minCount", "m", 0, "minimum word count")
	//DictionaryCommand.MarkFlagRequired("input")
}

type countable struct {
	item  string
	count uint64
}

func Dic(externalDic map[string]struct{}) error {
	reader := bufio.NewReader(os.Stdin)
	dic := make(map[string]uint64)
	for {
		if line, e := reader.ReadString('\n'); e != nil {
			break
		} else {
			if len(line) == 0 {
				continue
			}
			for _, token := range strings.Fields(line) {
				dic[token]++
			}
		}
	}
	// print dictionary
	tokens := make([]*countable, len(dic))
	i := 0
	for k, v := range dic {
		tokens[i] = &countable{k, v}
		i++
	}
	dic = nil
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].count > tokens[j].count
	})

	for _, token := range tokens {
		if *minCount>0 {
			if token.count >= *minCount {
				if externalDic == nil {
					os.Stdout.WriteString(fmt.Sprintf("%s %d\n", token.item, token.count))
				} else if _, ok := external[token.item]; ok {
					os.Stdout.WriteString(fmt.Sprintf("%s %d\n", token.item, token.count))
				}
			} else {
				break
			}
		} else {
			if externalDic == nil {
				os.Stdout.WriteString(fmt.Sprintf("%s %d\n", token.item, token.count))
			} else if _, ok := external[token.item]; ok {
				os.Stdout.WriteString(fmt.Sprintf("%s %d\n", token.item, token.count))
			}
		}
	}

	log.Printf("Collected %d tokens\n", len(tokens))
	return nil
}

type Word struct {
	ID    uint32
	Count uint32
	Word  string
}

type Dictionary struct {
	Words map[string]*Word
	Index []*Word
}

func (d *Dictionary) Len() int {
	return len(d.Words)
}

func LoadDictionary(path string) (d *Dictionary, e error) {
	var f *os.File
	if f, e = os.Open(path); e != nil {
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	d = new(Dictionary)
	d.Words = make(map[string]*Word)
	for {
		if line, e := reader.ReadString('\n'); e != nil {
			break
		} else {
			if len(line) == 0 {
				continue
			}
			parsedLine := strings.Fields(line)
			id := len(d.Words)
			word := parsedLine[0]
			var count int
			count, e = strconv.Atoi(parsedLine[1])
			w := Word{uint32(id), uint32(count), word}
			d.Words[w.Word] = &w
		}
	}
	d.Index = make([]*Word, d.Len())
	for _, w := range d.Words {
		d.Index[w.ID] = w
	}

	return
}


func loadExternal(path string) (d map[string]struct{}, e error) {
	var f *os.File
	if f, e = os.Open(path); e != nil {
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	d = make(map[string]struct{})
	for {
		if line, e := reader.ReadString('\n'); e != nil {
			break
		} else {
			if len(line) == 0 {
				continue
			}
			parsedLine := strings.Fields(line)
			word := parsedLine[0]
			d[word] = struct{}{}
		}
	}


	return
}
