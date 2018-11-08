package embed

import (
	"bufio"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vseledkin/gorpora/asm"
	"github.com/vseledkin/gorpora/dic"
	"encoding/json"
	"github.com/pkg/errors"
	"fmt"
)

var EmbedCommand *cobra.Command
var NNCommand *cobra.Command

var inputFilePath, dictionaryFilePath, method, model *string
var batchSize, window *uint32
var epochs *uint64
var threshold *float32

func init() {

	EmbedCommand = &cobra.Command{
		Use:   "embed",
		Short: "make word embeddings",
		Long:  "make word embeddings from space tokenized text file",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Embed:")
			log.Printf("\tMethod: %s\n", *method)
			log.Printf("\tDictionary: %s\n", *dictionaryFilePath)
			log.Printf("\tInput: %s\n", *inputFilePath)
			log.Printf("\tBatch size: %d\n", *batchSize)
			log.Printf("\tEpochs: %d\n", *epochs)
			log.Printf("\tWindow: %d\n", *window)

			if d, e := dic.LoadDictionary(*dictionaryFilePath); e != nil {
				log.Fatal(e)
			} else {
				log.Printf("Loaded dictionary of %d words", d.Len())

				rand.Seed(time.Now().UnixNano())
				var m Word2VecModel
				switch *method {
				case "skipgram+hs":
					m = Word2VecModel{Size: 128, Skipgram: true, Hs: true, Window: *window, Dictionary: d}
				case "cbow+hs":
					m = Word2VecModel{Size: 128, Cbow: true, Hs: true, Window: *window, Dictionary: d}
				default:
					log.Fatal(errors.New("unsupported method " + *method))
				}
				m.Train(*inputFilePath, d, *epochs, 4, 0.025, 1e-5)
				if output, e := os.Create("model.json"); e != nil {
					log.Fatal(e)
				} else {
					if e = json.NewEncoder(output).Encode(m); e != nil {
						log.Fatal(e)
					}
				}

			}

		},
	}

	NNCommand = &cobra.Command{
		Use:   "nn",
		Short: "explore nearest words",
		Long:  "explore nearest words",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("NN:")
			log.Printf("\tModel : %s\n", *model)
			log.Printf("\tThreshold : %f\n", *threshold)
			if f, e := os.Open(*model); e != nil {
				log.Fatal(e)
			} else {
				defer f.Close()
				var m Word2VecModel
				if e = json.NewDecoder(f).Decode(&m); e != nil {
					log.Fatal(e)
				} else {
					m.precompute()

					fi := bufio.NewReader(os.Stdin)
					for {
						fmt.Printf("query: ")
						if query, ok := readline(fi); ok {
							if len(query) > 0 {
								m.search(query)
							}
						} else {
							break
						}
					}

				}
			}
		},
	}

	model = NNCommand.Flags().StringP("model", "m", "", "model file path")
	threshold = NNCommand.Flags().Float32P("threshold", "t", 0.6, "min cosine distance")
	NNCommand.MarkFlagRequired("model")

	inputFilePath = EmbedCommand.Flags().StringP("input", "i", "", "text file path")
	dictionaryFilePath = EmbedCommand.Flags().StringP("dic", "d", "", "dictionary file path")
	batchSize = EmbedCommand.Flags().Uint32P("batch", "b", 128, "batch size")
	epochs = EmbedCommand.Flags().Uint64P("epochs", "e", 5, "number of epochs")
	window = EmbedCommand.Flags().Uint32P("window", "w", 5, "context window")
	method = EmbedCommand.Flags().StringP("method", "m", "skipgram+hs", "model type")
	EmbedCommand.MarkFlagRequired("input")
	EmbedCommand.MarkFlagRequired("dic")
	EmbedCommand.MarkFlagRequired("method")
}

const (
	//MINALPHA minimum learning speed allowed
	minalpha         = 0.000
	unigramTableSize = 100000000
	unigramPower     = 0.75
)

//var exptable = MakeW2VFunction(1000, 6)

/*
Word2VecModel word2vec model
*/
type Word2VecModel struct {
	Words, Size, Negative, Window uint32
	Hs, Cbow, Skipgram            bool
	Syn0, Syn1                    []float32
	Dictionary                    *dic.Dictionary
}

func (m *Word2VecModel) precompute() {
	for _, w := range m.Dictionary.Index {
		v := m.Syn0[w.ID*m.Size : (w.ID+1)*m.Size]
		asm.Sscale(1/asm.Snrm2(v), v)
	}
}

func (m *Word2VecModel) search(query string) {
	//s := int(m.Size)
	if w, ok := m.Dictionary.Words[query]; ok {
		qv := m.Syn0[w.ID*m.Size : (w.ID+1)*m.Size]
		for _, w := range m.Dictionary.Index {
			v := m.Syn0[w.ID*m.Size : (w.ID+1)*m.Size]
			dot := asm.Sdot(v, qv)
			if dot > 0.7 {
				log.Printf("%f %s", dot, w.Word)
			}
		}
	}
}

// precompute and cache unigram frequencies
func precomputeUnigramTable(dic *dic.Dictionary) (table []uint32) {
	table = make([]uint32, unigramTableSize)
	L := uint32(len(dic.Words))
	var trainWordsPow float64

	for _, w := range dic.Words {
		trainWordsPow += math.Pow(float64(w.Count), unigramPower)
	}

	var i uint32
	w := math.Pow(float64(dic.Index[i].Count), unigramPower) / trainWordsPow

	for a := range table {
		table[a] = i
		if float64(a)/float64(unigramTableSize) > w {
			i++
			w += math.Pow(float64(dic.Index[i].Count), unigramPower) / trainWordsPow
		}
		if i >= L {
			i = L - 1
		}
	}
	return
}

func MakeRandomVector(size uint32) (v []float32) {
	v = make([]float32, size)
	for j := range v {
		v[j] = (rand.Float32() - 0.5) / float32(size)
	}
	return
}
func readline(fi *bufio.Reader) (string, bool) {
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s[:len(s)-1], true
}

/*MaxInt max int*/
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/*MinInt min int*/
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

/*MaxFloat32 max float32*/
func MaxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func ReadCorpus(path string, d *dic.Dictionary, pipe chan []uint32) (e error) {
	var f *os.File
	if f, e = os.Open(path); e != nil {
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	for {
		if line, e := reader.ReadString('\n'); e != nil {
			break
		} else {
			if len(line) == 0 {
				continue
			}
			parsedLine := strings.Fields(line)
			document := make([]uint32, len(parsedLine))
			for i, token := range strings.Fields(line) {
				document[i] = d.Words[token].ID
			}
			pipe <- document
		}
	}
	pipe <- nil
	return
}

func fileStats(path string) (documents, totalWords uint64, e error) {
	var f *os.File
	if f, e = os.Open(path); e != nil {
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	for {
		if line, e := reader.ReadString('\n'); e != nil {
			break
		} else {
			if len(line) == 0 {
				continue
			}
			documents++
			totalWords += uint64(len(strings.Fields(line)))
		}
	}
	return
}

/*
Train train word 2 vec model
*/
func (m *Word2VecModel) Train(path string, d *dic.Dictionary, iterations, threads uint64, alpha, subsample float32) (e error) {
	switch {

	case m.Cbow && m.Hs:
		log.Println("CBOW + HS")
	case m.Skipgram && m.Hs:
		log.Println("SKIPGRAM + HS")
	default:
		log.Println("Inconsistent parameters")
		panic("Inconsistent parameters")
	}

	L := uint32(d.Len())
	log.Printf("Input-hidden: %d x %d\n", L, m.Size)
	log.Printf("Hidden-output: %d x %d\n", m.Size, L)
	m.Words = L
	// allocate memory
	m.Syn0 = MakeRandomVector(m.Words * m.Size)
	m.Syn1 = make([]float32, m.Size*m.Words)

	//var unigramTable []uint32
	var Codes [][]byte
	var Points [][]uint32

	rand.Seed(time.Now().UnixNano())
	if m.Hs {
		Codes, Points, _ = dic.BuildHuffmanTreeFromDictionary(d)
	}

	/*
		if m.Negative > 0 {
			unigramTable = precomputeUnigramTable(d)
		}*/

	var frequencyTable []float64

	var wordsReady float64
	var errorsReady float64
	var errorsCount float64
	var documents, totalWords uint64
	if documents, totalWords, e = fileStats(path); e != nil {
		return e
	}
	log.Printf("Input has %d documents %d words", documents, totalWords)

	start := time.Now()
	var sentenceCount uint32
	for iteration := range make([]struct{}, iterations) {
		//ialpha = alpha / (1.0 + float32(iteration)/float32(iterations))
		pipe := make(chan []uint32, 1024)
		go ReadCorpus(path, d, pipe)

		workChannel := make(chan struct {
			c float64
			e float64
		}, threads)
		var currentAlpha float32
		// setup reporting
		ticker := time.NewTicker(time.Second)
		go func() {
			prev := wordsReady
			var progress float64
			for range ticker.C {
				if prev != wordsReady {

					progress = 100.0 * float64(sentenceCount) / float64(documents*iterations)

					log.Printf("%cIt: %d Proc: %d Alpha: %f Progress: %.2f%% Loss: %f Words/thread/sec: %fk All word/sec: %fk", 13, iteration, runtime.NumGoroutine(),
						currentAlpha, progress, errorsReady/errorsCount, (wordsReady-prev)/float64(threads)/1000.0,
						(wordsReady-prev)/1000.0)
					prev = wordsReady
					errorsReady = 0
					errorsCount = 0
				}
			}
		}()

		// read text corpus ventilating Work throught workChannel
		for range make([]struct{}, threads) {
			workChannel <- struct {
				c float64
				e float64
			}{0, 100}
		}

		for document := range pipe {
			if document == nil {
				break
			}
			wc := <-workChannel
			wordsReady += wc.c
			errorsReady += wc.e
			errorsCount++

			currentAlpha = float32(MaxFloat32(minalpha, alpha*(1.0-float32(sentenceCount)/float32(documents*iterations))))

			if subsample > 0 {
				if len(frequencyTable) == 0 {
					frequencyTable = make([]float64, d.Len())
					var portion float64
					for _, w := range d.Words {
						portion = float64(w.Count) / (float64(totalWords) * float64(subsample))
						frequencyTable[w.ID] = (math.Sqrt(portion) + 1) / portion
					}
				}
			}
			switch {
			case m.Cbow && m.Hs:
				go m.updateCbowHs(document, currentAlpha, workChannel, Codes, Points)
			case m.Skipgram && m.Hs:
				go m.updateSkipGramHs(document, currentAlpha, workChannel, Codes, Points)
			default:
				panic("Inconsistent parameters")
			}
			sentenceCount++
		}
		ticker.Stop()

		for range make([]struct{}, threads) {
			wc := <-workChannel
			wordsReady += wc.c
			errorsReady += wc.e
			errorsCount++
		}

	}
	log.Printf("\ntotal_words: %d\ntotal_sentences: %d\ntraining time: %v\n", totalWords,
		sentenceCount, time.Now().Sub(start))
	return
}

func (m *Word2VecModel) updateCbowHs(sentence []uint32, alpha float32, workChannel chan struct {
	c float64
	e float64
}, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)

	window := int(m.Window)
	hidden, hiddenError := make([]float32, m.Size), make([]float32, m.Size)
	var reducedWindow, j, k, i, d, a, b int
	var code uint8
	s := m.Size
	var g, f float32
	var l2 []float32
	var word, current uint32
	var loss float64
	var lossCount float64
	for i, current = range sentence {
		reducedWindow = rand.Int() % window
		a = MaxInt(0, i-window+reducedWindow)
		b = MinInt(sentenceLength, i+window+1-reducedWindow)
		/*
		   train bag of words model, context predicts word in -> hidden
		*/
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			word = sentence[j]
			asm.Sxpy(m.Syn0[s*word:][:s], hidden)
		}
		// HIERARCHICAL SOFTMAX
		for d, code = range Codes[current] {
			// Propagate hidden -> output
			l2 = m.Syn1[s*Points[current][d]:][:s]

			f = asm.Sdot(hidden, l2)
			if f > 6.0 {
				f = 1
				g = (float32(1.0-code) - 1)
			} else if f < -6.0 {
				f = 0
				g = (float32(1.0-code) - 0)
			} else {
				f = float32(math.Exp(float64(f)))
				g = (float32(1.0-code) - f/(f+1.0))
			}
			loss += float64(g)
			lossCount++
			g *= alpha
			// 'g' is the gradient multiplied by the learning rate
			// Propagate errors output -> hidden
			asm.Saxpy(g, l2, hiddenError)
			// Learn weights hidden -> output
			asm.Saxpy(g, hidden, l2)
		}
		// hidden -> in
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			word = sentence[j]
			asm.Sxpy(hiddenError, m.Syn0[s*word:][:s])
		}
		asm.Sclean(hidden)
		asm.Sclean(hiddenError)
	}

	workChannel <- struct {
		c float64
		e float64
	}{float64(sentenceLength), loss / lossCount}
}

func (m *Word2VecModel) updateSkipGramHs(sentence []uint32, alpha float32, workChannel chan struct {
	c float64
	e float64
}, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)
	window := int(m.Window)
	neu1e := make([]float32, m.Size)
	var loss float64
	var lossCount float64
	var reducedWindow, j, k, i, d, a, b int
	var code uint8
	s := m.Size
	var g, f float32
	var l1, l2 []float32
	var word, current uint32
	for i, current = range sentence {
		reducedWindow = rand.Int() % window
		a = MaxInt(0, i-window+reducedWindow)
		b = MinInt(sentenceLength, i+window+1-reducedWindow)
		// train skip gram model
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			word = sentence[j]

			// HIERARCHICAL SOFTMAX
			l1 = m.Syn0[s*word:][:s]

			for d, code = range Codes[current] {
				// Propagate hidden -> output
				l2 = m.Syn1[s*Points[current][d]:][:s]
				f = asm.Sdot(l1, l2)
				if f > 6.0 {
					g = float32(1.0-code) - 1
				} else if f < -6.0 {
					g = float32(1.0-code) - 0
				} else {
					f = float32(math.Exp(float64(f)))
					g = float32(1.0-code) - f/(f+1.0)
				}
				loss += float64(g)
				lossCount++
				g *= alpha
				// Propagate errors output -> hidden
				asm.Saxpy(g, l2, neu1e)
				// Learn weights hidden -> output
				asm.Saxpy(g, l1, l2)
			}
			// Learn weights input -> hidden
			asm.Sxpy(neu1e, l1)
			asm.Sclean(neu1e)
		}

	}
	workChannel <- struct {
		c float64
		e float64
	}{float64(sentenceLength), loss / lossCount}
}
