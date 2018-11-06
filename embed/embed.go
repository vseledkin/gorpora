package embed

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/vseledkin/gorpora/dic"
	"math"
	"math/rand"
	"time"
	"github.com/vseledkin/gorpora/asm"
	"os"
	"bufio"
	"strings"
	"runtime"
)

var EmbedCommand *cobra.Command

var inputFilePath, dictionaryFilePath *string
var batchSize *uint32

func init() {

	EmbedCommand = &cobra.Command{
		Use:   "embed",
		Short: "make word embeddings",
		Long:  "make word embeddings from space tokenized text file",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Embed:")
			log.Printf("\tDictionary: %s\n", *dictionaryFilePath)
			log.Printf("\tInput: %s\n", *inputFilePath)
			log.Printf("\tBatch size: %d\n", *batchSize)

			if d, e := dic.LoadDictionary(*dictionaryFilePath); e != nil {
				log.Fatal(e)
			} else {
				log.Printf("Loaded dictionary of %d words", d.Len())

				rand.Seed(time.Now().UnixNano())
				m := Word2VecModel{Cbow: true, Hs: true, Window: 5}
				m.Train(*inputFilePath, d, 1, 4, 0.025, 1e-5)

				/*
								for i, v := range d.Index {
									log.Printf("%d %s %+v\n", i, v.Word, Codes[i])
									if i > 10 {
										break
									}
								}*/
			}

		},
	}

	inputFilePath = EmbedCommand.Flags().StringP("input", "i", "", "text file path")
	dictionaryFilePath = EmbedCommand.Flags().StringP("dic", "d", "", "dictionary file path")
	batchSize = EmbedCommand.Flags().Uint32P("batch", "b", 128, "batch size")
	EmbedCommand.MarkFlagRequired("input")
	EmbedCommand.MarkFlagRequired("dic")
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

			//atomic.AddUint64(&documents, 1)
		}
	}
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
	default:
		log.Println("Inconsistent parameters")
		panic("Inconsistent parameters")
	}

	m.Size = m.Size

	L := uint32(d.Len())
	log.Printf("Input-hidden: %dx%d\n", L, m.Size)
	log.Printf("Hidden-output: %dx%d\n", m.Size, L)
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

		workChannel := make(chan uint64, threads)
		var currentAlpha float32
		// setup reporting
		ticker := time.NewTicker(time.Second)
		go func() {
			prev := wordsReady
			var progress float64
			for range ticker.C {
				if prev != wordsReady {

					progress = 100.0 * float64(sentenceCount) / float64(documents*iterations)

					log.Printf("%cIt: %d Proc: %d Alpha: %f Progress: %.2f%% Words/thread/sec: %fk All word/sec: %fk", 13, iteration, runtime.NumGoroutine(),
						currentAlpha, progress, (wordsReady-prev)/float64(threads)/1000.0,
						(wordsReady-prev)/1000.0)
					prev = wordsReady
				}
			}
		}()

		// read text corpus ventilating Work throught workChannel
		for _ = range make([]struct{}, threads) {
			workChannel <- 0
		}

		for document := range pipe {
			wordsReady += float64(<-workChannel)

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
				go m.update(document, currentAlpha, workChannel, Codes, Points)
				//case m.Skipgram && m.Hs:
				//	go m.learnSentenceSkipGramHs(uint32(len(document.Text)), document.Text, currentAlpha, workChannel, Codes, Points)
			default:
				panic("Inconsistent parameters")
			}
			sentenceCount++
		}
		ticker.Stop()

		for _ = range make([]struct{}, threads) {
			wordsReady += float64(<-workChannel)
		}

	}
	log.Printf("\ntotal_words: %d\ntotal_sentences: %d\ntraining time: %v\n", totalWords,
		sentenceCount, time.Now().Sub(start))
	return
}

func (m *Word2VecModel) update(sentence []uint32, alpha float32, workChannel chan uint64, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)

	window := int(m.Window)
	hidden, hiddenError := make([]float32, m.Size), make([]float32, m.Size)
	var reducedWindow, j, k, i, d, a, b int
	var code uint8
	s := m.Size
	var g, f float32
	var l2 []float32
	var word, current uint32
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
			//if int(s*word+s) > len(m.Syn0) {
			//	panic("Problem!!!!!")
				//fmt.Println("Problem!!!!!")
				//fmt.Println("Word", word)
				//fmt.Println("Sentence", sentence)
				//fmt.Println("Len", len(m.Syn0))
				//fmt.Println("Index", s*word)
				//fmt.Println("Index+s", s*word+s)
			//	workChannel <- uint64(sentenceLength)
			//	return
			//}
			asm.Sxpy(m.Syn0[s*word:][:s], hidden)
		}
		// HIERARCHICAL SOFTMAX

		//fmt.Println(current, m.Codes[current])
		for d, code = range Codes[current] {
			// Propagate hidden -> output
			l2 = m.Syn1[s*Points[current][d]:][:s]

			f = asm.Sdot(hidden, l2)

			if f > 6.0 {
				g = (float32(1.0-code) - 1) * alpha
			} else if f < -6.0 {
				g = (float32(1.0-code) - 0) * alpha
			} else {
				f = float32(math.Exp(float64(f)))
				g = (float32(1.0-code) - f/(f+1.0)) * alpha
			}

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

	workChannel <- uint64(sentenceLength)
}
