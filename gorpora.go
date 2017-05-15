package gorpora

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"strings"
	"unicode"
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
