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
	c "github.com/vseledkin/gorpora/common"
	"encoding/json"
	"github.com/pkg/errors"
	"fmt"
	"sort"
	"github.com/vseledkin/gorpora/small_embed"
)

var EmbedCommand *cobra.Command
var NNCommand *cobra.Command
var inputFilePath *[]string
var dictionaryFilePath, method, model *string
var batchSize, window *uint32
var alpha *float32

var threshold *float32

var epochs, threads *uint64

var frequencyTable []float64

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
			for _, path := range *inputFilePath {
				log.Printf("\tInput: %s\n", path)
			}
			log.Printf("\tBatch size: %d\n", *batchSize)
			log.Printf("\tEpochs: %d\n", *epochs)
			log.Printf("\tWindow: %d\n", *window)
			log.Printf("\tThreads: %d\n", *threads)

			if d, e := dic.LoadDictionary(*dictionaryFilePath); e != nil {
				log.Fatal(e)
			} else {
				log.Printf("Loaded dictionary of %d words", d.Len())

				rand.Seed(time.Now().UnixNano())

				var m Word2VecModel
				switch *method {
				case "skipgram+neg":
					m = Word2VecModel{Size: 128, Skipgram: true, Hs: true, Window: *window, Dictionary: d, Negative: 10}
				case "skipgram+hs":
					m = Word2VecModel{Size: 128, Skipgram: true, Hs: true, Window: *window, Dictionary: d}
				case "cbow+hs":
					m = Word2VecModel{Size: 128, Cbow: true, Hs: true, Window: *window, Dictionary: d}
				default:
					log.Fatal(errors.New("unsupported method " + *method))
				}
				m.Train(*inputFilePath, d, *epochs, *threads, *alpha, 1e-5)
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

	inputFilePath = EmbedCommand.Flags().StringArrayP("input", "i", []string{}, "text file paths, supports reading from multiple files")
	dictionaryFilePath = EmbedCommand.Flags().StringP("dic", "d", "", "dictionary file path")
	batchSize = EmbedCommand.Flags().Uint32P("batch", "b", 128, "batch size")
	epochs = EmbedCommand.Flags().Uint64P("epochs", "e", 5, "number of epochs")
	threads = EmbedCommand.Flags().Uint64P("threads", "t", 2, "number of threads")
	window = EmbedCommand.Flags().Uint32P("window", "w", 5, "context window")
	method = EmbedCommand.Flags().StringP("method", "m", "skipgram+hs", "model type")
	alpha = EmbedCommand.Flags().Float32P("alpha", "a", 0.025, "start learning rate")
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

type weightable struct {
	item   string
	weight float32
}

func (m *Word2VecModel) search(query string) {
	//s := int(m.Size)
	var result []*weightable
	if w, ok := m.Dictionary.Words[query]; ok {
		qv := m.Syn0[w.ID*m.Size : (w.ID+1)*m.Size]
		for _, w := range m.Dictionary.Index {
			v := m.Syn0[w.ID*m.Size : (w.ID+1)*m.Size]
			dot := asm.Sdot(v, qv)
			if dot > *threshold {
				if w.Count > 1 {
					result = append(result, &weightable{w.Word, dot})
				}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].weight > result[j].weight
	})
	if len(result) > 35 {
		result = result[:35]
	}
	for _, r := range result {
		log.Printf("%f %s", r.weight, r.item)
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

func ReadCorpus(path string, d *dic.Dictionary, pipe chan []uint32, ft []float64, wordsReady *float64) (e error) {
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
			*wordsReady += float64(len(parsedLine))
			document := make([]uint32, len(parsedLine))
			var i int
			for _, token := range strings.Fields(line) {
				if w, ok := d.Words[token]; ok {
					//if ft[w.ID] >= rand.Float64() { turn sampling off
					document[i] = w.ID
					i++
					//}
				}
			}
			document = document[:i]
			if len(document) > 1 {
				pipe <- document
			}
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

type WC struct {
	c float64
	e float64
	r *rand.Rand
	//	errors []float32
}

/*
Train  word 2 vec model
*/
func (m *Word2VecModel) Train(paths []string, d *dic.Dictionary, iterations, threads uint64, alpha, subsample float32) (e error) {
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

	var unigramTable []uint32
	if m.Negative > 0 {
		unigramTable = precomputeUnigramTable(d)
	}

	var wordsReady float64
	var errorsReady float64
	var errorsCount float64
	var documents, totalWords uint64
	for _, path := range paths {
		log.Printf("Counting words in %s", path)
		var d, t uint64
		if d, t, e = fileStats(path); e != nil {
			return e
		}
		documents += d
		totalWords += t
	}
	log.Printf("Input has %d documents %d words", documents, totalWords)

	frequencyTable := make([]float64, d.Len())
	var portion float64
	for _, w := range d.Words {
		portion = float64(w.Count) / (float64(totalWords) * float64(subsample))
		frequencyTable[w.ID] = (math.Sqrt(portion) + 1) / portion
	}

	start := time.Now()
	var sentenceCount uint32
	ma := small_embed.NewMovingAverage(1024)
	for iteration := range make([]struct{}, iterations) {
		//ialpha = alpha / (1.0 + float32(iteration)/float32(iterations))
		pipe := make(chan []uint32, 1024)
		for _, path := range paths {
			go ReadCorpus(path, d, pipe, frequencyTable, &wordsReady)
		}

		workChannel := make(chan *WC, threads)
		var currentAlpha float32
		// setup reporting
		ticker := time.NewTicker(time.Second)
		go func() {
			prev := wordsReady
			var progress float64
			for range ticker.C {
				if prev != wordsReady {
					ma.Add(float32(errorsReady / errorsCount))
					progress = 100.0 * float64(sentenceCount) / float64(documents*iterations)

					log.Printf("%cIt: %d Proc: %d Alpha: %f Progress: %.2f%% Loss: %f Words/thread/sec: %fk All word/sec: %fk", 13, iteration, runtime.NumGoroutine(),
						currentAlpha, progress, ma.Avg(), (wordsReady-prev)/float64(threads)/1000.0,
						(wordsReady-prev)/1000.0)
					prev = wordsReady
					errorsReady = 0
					errorsCount = 0
				}
			}
		}()

		// read text corpus ventilating Work throught workChannel
		for t := range make([]struct{}, threads) {
			workChannel <- &WC{0, 100, rand.New(rand.NewSource(int64(t)))}
		}
		var nilCount int
		for document := range pipe {
			if document == nil {
				nilCount++
				if nilCount == len(paths) {
					break
				}
			}
			wc := <-workChannel
			//wordsReady += wc.c
			errorsReady += wc.e
			errorsCount++

			currentAlpha = float32(c.MaxFloat32(minalpha, alpha*(1.0-float32(sentenceCount)/float32(documents*iterations))))

			/*if subsample > 0 {
				if len(frequencyTable) == 0 {
					frequencyTable = make([]float64, d.Len())
					var portion float64
					for _, w := range d.Words {
						portion = float64(w.Count) / (float64(totalWords) * float64(subsample))
						frequencyTable[w.ID] = (math.Sqrt(portion) + 1) / portion
					}
				}
			}*/
			switch {
			case m.Cbow && m.Hs:
				go m.updateCbowHs(document, currentAlpha, workChannel, wc, Codes, Points)
			case m.Skipgram && m.Hs:
				go m.updateSkipGramHs(document, currentAlpha, workChannel, wc, Codes, Points)
			case m.Skipgram && m.Negative > 0:
				go m.updateSkipGramNeg(document, currentAlpha, workChannel, wc, unigramTable)
			default:
				panic("Inconsistent parameters")
			}
			sentenceCount++
		}
		ticker.Stop()

		for range make([]struct{}, threads) {
			wc := <-workChannel
			//wordsReady += wc.c
			errorsReady += wc.e
			errorsCount++
		}

		if output, e := os.Create(fmt.Sprintf("model.%d.json", iteration)); e != nil {
			log.Fatal(e)
		} else {
			if e = json.NewEncoder(output).Encode(m); e != nil {
				log.Fatal(e)
			}

		}
		log.Printf("Model saved at iteration %d to model.%d.json", iteration, iteration)
	}
	log.Printf("\ntotal_words: %d\ntotal_sentences: %d\ntraining time: %v\n", totalWords,
		sentenceCount, time.Now().Sub(start))
	return
}

func (m *Word2VecModel) updateCbowHs(sentence []uint32, alpha float32, workChannel chan *WC, wc *WC, Codes [][]byte, Points [][]uint32) {
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
	//var lossCount float64
	for i, current = range sentence {
		reducedWindow = wc.r.Intn(window)
		a = c.MaxInt(0, i-window+reducedWindow)
		b = c.MinInt(sentenceLength, i+window+1-reducedWindow)
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
				g = float32(1.0-code) - 1
			} else if f < -6.0 {
				f = 0
				g = float32(1.0-code) - 0
			} else {
				f = float32(math.Exp(float64(f)))
				g = float32(1.0-code) - f/(f+1.0)
			}
			loss += float64(g)
			//lossCount++
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

	wc.c = float64(sentenceLength)
	wc.e = loss /// lossCount
	workChannel <- wc
}

func (m *Word2VecModel) updateSkipGramHs(sentence []uint32, alpha float32, workChannel chan *WC, wc *WC, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)
	window := int(m.Window)
	errors := make([]float32, m.Size)
	var loss float64
	//var lossCount float64
	var reducedWindow, j, k, i, d, a, b int
	var code uint8
	s := m.Size
	var g, f float32
	var l1, l2 []float32
	var word, current uint32
	for i, current = range sentence {
		reducedWindow = rand.Int() % window
		a = c.MaxInt(0, i-window+reducedWindow)
		b = c.MinInt(sentenceLength, i+window+1-reducedWindow)

		// train skip gram model
		for j, k = a, b; j < k; j++ {
			asm.Sclean(errors)
			if j == i {
				continue
			}
			word = sentence[j]

			// HIERARCHICAL SOFTMAX
			l1 = m.Syn0[s*word:][:s]
			points := Points[current]
			for d, code = range Codes[current] {
				// Propagate hidden -> output
				l2 = m.Syn1[s*points[d] : s*points[d]+s]
				f = asm.Sdot(l1, l2)

				if f > 6.0 {
					g = float32(1.0-code) - 1
				} else if f < -6.0 {
					g = float32(1.0-code) - 0
				} else {
					f = float32(math.Exp(float64(f)))
					g = float32(1.0-code) - f/(f+1.0)
				}
				loss -= float64(g)
				//lossCount++
				g *= alpha
				// Propagate errors output -> hidden
				asm.Saxpy(g, l2, errors)
				// Learn weights hidden -> output
				asm.Saxpy(g, l1, l2)
			}
			// Learn weights input -> hidden
			asm.Sxpy(errors, l1)
		}

	}
	wc.c = float64(sentenceLength)
	wc.e = loss /// lossCount
	workChannel <- wc
}

func (m *Word2VecModel) updateSkipGramNeg(sentence []uint32, alpha float32, workChannel chan *WC, wc *WC, unigramTable []uint32) {
	sentenceLength := len(sentence)
	window := int(m.Window)
	errors := make([]float32, m.Size)
	var loss float64
	//var lossCount float64
	var reducedWindow, j, k, i, d, a, b int

	s := m.Size
	var g, f float32
	var l1, l2 []float32
	var word, current uint32
	for i, current = range sentence {
		reducedWindow = rand.Int() % window
		a = c.MaxInt(0, i-window+reducedWindow)
		b = c.MinInt(sentenceLength, i+window+1-reducedWindow)

		// train skip gram model
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			word = sentence[j]

			// NEGATIVE SAMPLING
			l1 = m.Syn0[s*word:][:s]
			var target uint32
			var label float32
			for d = 0; d < int(m.Negative)+1; d++ {
				if d == 0 {
					target = current
					label = 1
				} else {
					target = unigramTable[rand.Intn(unigramTableSize)]
					//fmt.Println(target)
					if target == current {
						continue
					}
					label = 0
				}
				l2 = m.Syn1[s*target:][:s]

				f = asm.Sdot(l1, l2)

				if f > 6.0 {
					g = label - 1
				} else if f < -6.0 {
					g = label - 0
				} else {
					f = float32(math.Exp(float64(f)))
					g = label - f/(f+1.0)
				}
				loss -= float64(g)
				g *= alpha
				asm.Saxpy(g, l2, errors)
				asm.Saxpy(g, l1, l2)
			}

			// Learn weights input -> hidden
			asm.Sxpy(errors, l1)
			asm.Sclean(errors)
		}

	}
	wc.c = float64(sentenceLength)
	wc.e = loss /// lossCount
	workChannel <- wc
}
