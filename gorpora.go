package gorpora

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/vseledkin/gorpora/cld2"
)

func NormalizeHtmlEntities() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		os.Stdout.WriteString(html.UnescapeString(line))
	}
}

func Split() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		os.Stdout.WriteString(split2Tokens(line))
		os.Stdout.WriteString("\n")
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

func FilterLanguage(languages []string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		language := cld2.Detect(line)
		accept := false
		for _, lang := range languages {
			if language == lang {
				accept = true
				break
			}
		}
		if accept {
			os.Stdout.WriteString(line)
			//os.Stdout.WriteString("\n")
		}
	}
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
