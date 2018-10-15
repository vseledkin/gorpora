package udpipe

import (
	"os/exec"

	"log"

	"bufio"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type Sentence struct {
	ID     int
	Body   string
	Tokens []*Token
}

func (s *Sentence) MakeDependencies() {
	for _, token := range s.Tokens {

		if token.Dependency > 0 {
			token.DependencyToken = s.Tokens[token.Dependency-1]
		}
	}
}

type Token struct {
	ID              int
	Dependency      int
	Word            string
	Lemma           string
	Pos             string
	FinePos         string
	Features        map[string]string
	Function        string
	DependencyToken *Token
}

type Parser struct {
	stdout  io.ReadCloser
	stdin   io.WriteCloser
	stderr  io.ReadCloser
	cmd     *exec.Cmd
	licence chan struct{}
}

func (p *Parser) Close() {
	p.stdin.Close()
	p.stdout.Close()
	p.stderr.Close()
}

func (p *Parser) Start() (err error) {
	if p.licence == nil {
		p.licence = make(chan struct{}, 1)
		p.licence <- struct{}{}
	}
	os := runtime.GOOS
	arch := runtime.GOARCH
	ppath := fmt.Sprintf("./udpipe/udpipe_%s_%s", os, arch)
	log.Printf("Starting %s parser\n", ppath)
	//p.cmd = exec.Command(ppath, "--tokenize", "--tag", "--parse", "--immediate", "./udpipe/english-ud-2.0-conll17-170315.udpipe")
	p.cmd = exec.Command(ppath, "--tokenize", "--tag", "--parse", "--immediate", "./udpipe/russian-ud-2.0-170801.udpipe")
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return err
	}

	p.stdout, err = p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	p.stderr, err = p.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the process
	if err = p.cmd.Start(); err != nil {
		return err
	}
	var w sync.WaitGroup
	w.Add(1)
	go func() {
		scanner := bufio.NewScanner(p.stderr)
		for scanner.Scan() {
			log.Printf("parser stderr-> %s\n", scanner.Text())
			w.Done()
		}
	}()
	w.Wait()
	return nil
}
func LenWithoutSpaces(str string) int {
	return len(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str))
}
func (p *Parser) Parse(text string) (sentences []*Sentence, err error) {
	<-p.licence
	defer func() {
		p.licence <- struct{}{}
	}()
	L := LenWithoutSpaces(text)
	LL := 0
	if _, err = p.stdin.Write([]byte(text + "\n\n")); err != nil {
		println(err)
	}

	//go func() {
	scanner := bufio.NewScanner(p.stdout)
	sentenceID := -1
	var sentence *Sentence = nil
Loop:
	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf("o-> [%s]\n", line)
		switch {
		case len(line) == 0:
			if sentence == nil {
				panic(fmt.Errorf("Sentence cannot be nil"))
			}
			sentences = append(sentences, sentence)
			sentence = nil
			//log.Printf("o-eos> L=%d LL=%d\n", L, LL)
			if L == LL {
				break Loop // end of sentence
			}
		case line == "# newdoc":
		case line == "# newpar":
		case strings.HasPrefix(line, "# text = "):
			sentence.Body = line[9:]
			LL += LenWithoutSpaces(sentence.Body)
		case sentence != nil: // parse token line
			lineParts := strings.Fields(line)
			token := &Token{}
			token.ID, err = strconv.Atoi(lineParts[0])
			if err != nil {
				panic(err) // serious error parser will not be usable after this
			}
			token.Word = lineParts[1]
			token.Lemma = strings.ToLower(lineParts[2])
			token.Pos = lineParts[3]
			token.FinePos = lineParts[4]
			for _, feature := range strings.Split(lineParts[5], "|") {
				if token.Features == nil {
					token.Features = make(map[string]string)
				}
				featureParts := strings.Split(feature, "=")
				if len(featureParts) == 2 {
					token.Features[featureParts[0]] = featureParts[1]
				}
			}
			token.Dependency, err = strconv.Atoi(lineParts[6])
			if err != nil {
				panic(err) // serious error parser will not be usable after this
			}
			token.Function = lineParts[7]
			sentence.Tokens = append(sentence.Tokens, token)
		case strings.HasPrefix(line, "# sent_id ="):
			sentenceID, err = strconv.Atoi(strings.Split(line, "= ")[1])
			if err != nil {
				panic(err) // serious error parser will not be usable after this
			}
			sentence = &Sentence{ID: sentenceID}
		}
	}
	// check integrity
	if L != LL {
		panic(fmt.Errorf("Input length %d does not match output length %d ", L, LL))
	}
	return
}
