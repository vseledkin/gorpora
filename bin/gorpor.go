package gorpora

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"html"
	"log"
	"os"
	"strings"
	"time"

	"bytes"

	"unicode/utf8"

	"fmt"
	"unicode"

	"github.com/vseledkin/gorpora/cld2"
	"github.com/vseledkin/gorpora/udpipe"
)

var startTime time.Time

func init() {
	startTime = time.Now().Add(-time.Second)
}

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

// stripTags takes a snippet of HTML and returns only the text content.
// For example, `<b>&iexcl;Hi!</b> <script>...</script>` -> `&iexcl;Hi! `.
func stripTags(html string) string {
	var b bytes.Buffer
	s, c, i, allText := []byte(html), context{}, 0, true
	// Using the transition funcs helps us avoid mangling
	// `<div title="1>2">` or `I <3 Ponies!`.
	for i != len(s) {
		if c.delim == delimNone {
			st := c.state
			// Use RCDATA instead of parsing into JS or CSS styles.
			if c.element != elementNone && !isInTag(st) {
				st = stateRCDATA
			}
			d, nread := transitionFunc[st](c, s[i:])
			i1 := i + nread
			if c.state == stateText || c.state == stateRCDATA {
				// Emit text up to the start of the tag or comment.
				j := i1
				if d.state != c.state {
					for j1 := j - 1; j1 >= i; j1-- {
						if s[j1] == '<' {
							j = j1
							break
						}
					}
				}
				b.Write(s[i:j])
			} else {
				allText = false
			}
			c, i = d, i1
			continue
		}
		i1 := i + bytes.IndexAny(s[i:], delimEnds[c.delim])
		if i1 < i {
			break
		}
		if c.delim != delimSpaceOrTagEnd {
			// Consume any quote.
			i1++
		}
		c, i = context{state: stateTag, element: c.element}, i1
	}
	if allText {
		return html
	} else if c.state == stateText || c.state == stateRCDATA {
		b.Write(s[i:])
	}
	return b.String()
}

func StripHtml() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = stripTags(line)
		os.Stdout.WriteString(html.UnescapeString(line))
	}
}

func Split(use_udpipe, output_lemmas bool) {
	if use_udpipe {
		PARSER = new(udpipe.Parser)
		PARSER.Start()
		defer PARSER.Close()
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if use_udpipe {
			sentences, err := PARSER.Parse(line)
			if err != nil {
				log.Println(err)
				return
			}
			var tokens []string
			for _, sentence := range sentences {
				for _, token := range sentence.Tokens {
					if output_lemmas {
						tokens = append(tokens, token.Lemma)
					} else {
						tokens = append(tokens, token.Word)
					}
				}
			}
			os.Stdout.WriteString(strings.Join(tokens, " "))
		} else {
			os.Stdout.WriteString(split2Tokens(line))
		}
		os.Stdout.WriteString("\n")
	}
}

var PARSER *udpipe.Parser

func Sentesize(min, max int) {
	PARSER = new(udpipe.Parser)
	PARSER.Start()
	defer PARSER.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		sentences, err := PARSER.Parse(line)
		if err != nil {
			log.Println(err)
			return
		}
		for _, sentence := range sentences {
			L := utf8.RuneCountInString(sentence.Body)
			if L >= min && L <= max {
				os.Stdout.WriteString(sentence.Body)
				os.Stdout.WriteString("\n")
			}
		}
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