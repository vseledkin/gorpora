package small_embed

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
	//	"sort"
	"sort"
)

var EmbedCommand *cobra.Command
var NNCommand *cobra.Command
var inputFilePath, nnInputPath *[]string
var smallDictionaryFilePath, bigDictionaryFilePath, method, model *string
var batchSize, window *uint32

var threshold *float32
var alpha *float32

var epochs, threads *uint64

func init() {

	EmbedCommand = &cobra.Command{
		Use:   "small_embed",
		Short: "make word embeddings",
		Long:  "make word embeddings from space tokenized text file",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Small Embed:")
			log.Printf("\tMethod: %s\n", *method)
			log.Printf("\tBig dictionary: %s\n", *bigDictionaryFilePath)
			log.Printf("\tSmall dictionary: %s\n", *smallDictionaryFilePath)
			for _, path := range *inputFilePath {
				log.Printf("\tInput: %s\n", path)
			}
			log.Printf("\tBatch size: %d\n", *batchSize)
			log.Printf("\tEpochs: %d\n", *epochs)
			log.Printf("\tWindow: %d\n", *window)
			log.Printf("\tThreads: %d\n", *threads)
			log.Printf("\tAlpha: %f\n", *alpha)

			var e error
			var bigDictionary, smallDictionary *dic.Dictionary
			if bigDictionary, e = dic.LoadDictionary(*bigDictionaryFilePath); e != nil {
				log.Fatal(e)
			}

			log.Printf("Loaded dictionary of %d words", bigDictionary.Len())

			if smallDictionary, e = dic.LoadDictionary(*smallDictionaryFilePath); e != nil {
				log.Fatal(e)
			}

			log.Printf("Loaded dictionary of %d words", smallDictionary.Len())

			rand.Seed(time.Now().UnixNano())
			m := Word2VecModel{Size: 128, PosSize: 16, Skipgram: true, Hs: true, Window: *window, BigDictionary: bigDictionary, SmallDictionary: smallDictionary}
			switch *method {
			case "skipgram+hs":
				m.Skipgram = true
				m.Cbow = false
			case "cbow+hs":
				m.Skipgram = false
				m.Cbow = true
			default:
				log.Fatal(errors.New("unsupported method " + *method))
			}
			m.Train(*inputFilePath, *epochs, *threads, *alpha, 1e-5)
			if output, e := os.Create("model.json"); e != nil {
				log.Fatal(e)
			} else {
				if e = json.NewEncoder(output).Encode(m); e != nil {
					log.Fatal(e)
				}
			}

		},
	}

	NNCommand = &cobra.Command{
		Use:   "small_nn",
		Short: "explore nearest words",
		Long:  "explore nearest words",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Small NN:")
			log.Printf("\tModel : %s\n", *model)
			log.Printf("\tInput : %s\n", *nnInputPath)
			log.Printf("\tThreshold : %f\n", *threshold)

			if f, e := os.Open(*model); e != nil {
				log.Fatal(e)
			} else {
				defer f.Close()
				var m Word2VecModel
				if e = json.NewDecoder(f).Decode(&m); e != nil {
					log.Fatal(e)
				} else {
					m.precompute(*nnInputPath)

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
	nnInputPath = NNCommand.Flags().StringArrayP("input", "i", []string{}, "text file paths, supports reading from multiple files")
	NNCommand.MarkFlagRequired("model")
	NNCommand.MarkFlagRequired("input")

	inputFilePath = EmbedCommand.Flags().StringArrayP("input", "i", []string{}, "text file paths, supports reading from multiple files")
	bigDictionaryFilePath = EmbedCommand.Flags().StringP("big_dic", "d", "", "big dictionary file path")
	smallDictionaryFilePath = EmbedCommand.Flags().StringP("small_dic", "s", "", "small dictionary file path")
	batchSize = EmbedCommand.Flags().Uint32P("batch", "b", 128, "batch size")
	epochs = EmbedCommand.Flags().Uint64P("epochs", "e", 5, "number of epochs")
	threads = EmbedCommand.Flags().Uint64P("threads", "t", 2, "number of threads")
	window = EmbedCommand.Flags().Uint32P("window", "w", 5, "context window")
	method = EmbedCommand.Flags().StringP("method", "m", "cbow+hs", "model type")
	alpha = EmbedCommand.Flags().Float32P("alpha", "a", 0.025, "model type")
	EmbedCommand.MarkFlagRequired("input")
	EmbedCommand.MarkFlagRequired("big_dic")
	EmbedCommand.MarkFlagRequired("small_dic")
}

const (
	//MINALPHA minimum learning speed allowed
	minalpha     = 0.000
	unigramPower = 0.75
)

//var exptable = MakeW2VFunction(1000, 6)

/*
Word2VecModel word2vec model
*/
type Word2VecModel struct {
	Size, PosSize, Negative, Window uint32
	Hs, Cbow, Skipgram              bool
	Syn0, Syn1, Pos                 []float32
	BigDictionary                   *dic.Dictionary
	SmallDictionary                 *dic.Dictionary
}

func (m *Word2VecModel) precompute(paths []string) {
	pipe := make(chan [][]uint32, 1024)
	for _, path := range paths {
		go ReadCorpus(path, m.BigDictionary, m.SmallDictionary, pipe)
	}
	nilcount := 0
	s := m.Size
	dic := make(map[uint32]struct{}, 1024)
	// clean all ouptut vectors
	asm.Sclean(m.Syn1)

	for document := range pipe {
		if document == nil {
			nilcount++
			if nilcount == len(paths) {
				break
			}
		}

		for _, word := range document {
			if _, ok := dic[word[0]]; !ok {
				var p uint32
				for _, subword := range word[1:] {
					asm.Sxmulelyplusz(m.Syn0[s*subword:s*subword+s], m.Pos[s*p:s*p+s], m.Syn1[s*word[0]:s*word[0]+s])
					p++
				}
				dic[word[0]] = struct{}{}
			}
		}
	}
	// normalize vectors
	for _, w := range m.BigDictionary.Index {
		v := m.Syn1[w.ID*m.Size : (w.ID+1)*m.Size]
		asm.Sscale(1/asm.Snrm2(v), v)
	}
}

type weightable struct {
	item   string
	weight float32
}

func (m *Word2VecModel) search(query string) {
	var result []*weightable
	if w, ok := m.BigDictionary.Words[query]; ok {
		qv := m.Syn1[w.ID*m.Size : (w.ID+1)*m.Size]
		for _, w := range m.BigDictionary.Index {
			v := m.Syn1[w.ID*m.Size : (w.ID+1)*m.Size]
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

func restore(s string) string {
	ret := make([]rune, len(s))
	p := 0
	for _, l := range s {
		if l != 9601 && l != 32 {
			ret[p] = l
			p++
		}
	}
	return string(ret[:p])
}

func ReadCorpus(path string, big_dic, small_dic *dic.Dictionary, pipe chan [][]uint32) (e error) {
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
			// split to words
			line = strings.TrimSpace(line)
			parsedLine := strings.Split(line, "¡")
			document := make([][]uint32, len(parsedLine))
			i := 0
			for _, token := range parsedLine {
				subtokens := strings.Fields(token)
				if len(subtokens) > 16 {
					log.Printf("Token too long [%s] %+v", token, subtokens)
					continue
				}
				token = restore(token)
				if token == "..." {
					continue
				}
				// split token
				document[i] = make([]uint32, len(subtokens)+1)
				// put big word id as first token
				if t, ok := big_dic.Words[token]; ok {
					document[i][0] = t.ID
				} else {
					log.Printf("Token [%s] not found in big dictionary", token)
					continue
				}
				// put subtokens
				j := 0
				for _, subtoken := range subtokens {
					if st, ok := small_dic.Words[subtoken]; ok {
						document[i][1+j] = st.ID
						j++
					} else {
						log.Printf("Token [%s] not found in small dictionary", subtoken)
					}
				}
				document[i] = document[i][:1+j]
				i++
			}
			pipe <- document[:i]
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
			totalWords += uint64(len(strings.Split(line, "¡")))
		}
	}
	return
}
type WC struct {
	c float64
	e float64
	r *rand.Rand
}
/*
Train  word 2 vec model
*/
func (m *Word2VecModel) Train(paths []string, iterations uint64, threads uint64, alpha, subsample float32) (e error) {
	switch {

	case m.Cbow && m.Hs:
		log.Println("CBOW + HS")
	case m.Skipgram && m.Hs:
		log.Println("SKIPGRAM + HS")
	default:
		log.Println("Inconsistent parameters")
		panic("Inconsistent parameters")
	}

	log.Printf("Input-hidden: %d x %d\n", m.SmallDictionary.Len(), m.Size)
	log.Printf("Hidden-output: %d x %d\n", m.Size, m.BigDictionary.Len())

	// allocate memory
	m.Syn0 = MakeRandomVector(uint32(m.SmallDictionary.Len()) * m.Size)
	m.Syn1 = MakeRandomVector(m.Size * uint32(m.BigDictionary.Len()))
	m.Pos = MakeRandomVector(m.PosSize * m.Size)

	//var unigramTable []uint32
	var Codes [][]byte
	var Points [][]uint32

	rand.Seed(time.Now().UnixNano())
	if m.Hs {
		Codes, Points, _ = dic.BuildHuffmanTreeFromDictionary(m.BigDictionary)
	}

	/*
		if m.Negative > 0 {
			unigramTable = precomputeUnigramTable(d)
		}*/

	//var frequencyTable []float64

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

	start := time.Now()
	var sentenceCount uint32
	for iteration := range make([]struct{}, iterations) {
		//ialpha = alpha / (1.0 + float32(iteration)/float32(iterations))
		pipe := make(chan [][]uint32, 1024)
		for _, path := range paths {
			go ReadCorpus(path, m.BigDictionary, m.SmallDictionary, pipe)
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
			wordsReady += wc.c
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

func (m *Word2VecModel) updateCbowHs(sentence [][]uint32, alpha float32, workChannel chan *WC, wc *WC, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)

	s := m.Size
	window := int(m.Window)
	hidden, hiddenError := make([]float32, s), make([]float32, s)
	hidden0, posError, wordError := make([]float32, m.PosSize*s), make([]float32, m.PosSize*s), make([]float32, m.PosSize*s)
	var reducedWindow, j, k, d, a, b int
	var code uint8
	var g, f float32
	var l2 []float32
	var p, subword uint32
	var loss float64
	var lossCount float64
	for i, current := range sentence {
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
			//word = sentence[j][0]
			p = 0
			for _, subword := range sentence[j][1:] {
				// forward
				asm.Sxmulelyplusz(m.Syn0[s*subword:s*subword+s], m.Pos[s*p:s*p+s], hidden0[s*p:s*p+s])
				asm.Sxpy(hidden0[s*p:s*p+s], hidden)
				p++
			}
			//asm.Sscale(1/float32(len(sentence[j])-1), hidden)
		}
		// HIERARCHICAL SOFTMAX
		for d, code = range Codes[current[0]] {
			// Propagate hidden -> output
			l2 = m.Syn1[s*Points[current[0]][d]:][:s]

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
			lossCount++
			g *= alpha
			// 'g' is the gradient multiplied by the learning rate
			// Propagate errors output -> hidden
			asm.Saxpy(g, l2, hiddenError)
			// Learn output weights
			asm.Saxpy(g, hidden, l2)
			asm.Sscale(1/asm.Snrm2(l2), l2)

		}
		// propagate error from penaltimate layer
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			p = 0
			for _, subword = range sentence[j][1:] {
				// get positional error
				asm.Sxmulelyplusz(hiddenError, m.Syn0[s*subword:s*subword+s], posError[s*p:s*p+s])
				// get word error
				asm.Sxmulelyplusz(hiddenError, m.Pos[s*p:s*p+s], wordError[s*p:s*p+s])
				p++
			}
			p = 0
			// learn input layer
			for _, subword = range sentence[j][1:] {
				// learn positions
				asm.Sxpy(posError[s*p:s*p+s], m.Pos[s*p:s*p+s])
				// learn words
				asm.Sxpy(wordError[s*p:s*p+s], m.Syn0[s*subword:s*subword+s])
				p++
			}
		}
		asm.Sclean(hidden)
		asm.Sclean(hiddenError)
		asm.Sclean(hidden0)
		asm.Sclean(posError)
		asm.Sclean(wordError)
	}
	wc.c = float64(sentenceLength)
	wc.e = loss / lossCount
	workChannel <- wc
}

func (m *Word2VecModel) updateSkipGramHs(sentence [][]uint32, alpha float32, workChannel chan *WC, wc *WC, Codes [][]byte, Points [][]uint32) {
	sentenceLength := len(sentence)
	window := int(m.Window)
	neu1e := make([]float32, m.Size)
	var loss float64
	var lossCount float64
	var reducedWindow, j, k, d, a, b int
	var code uint8
	s := m.Size
	var g, f float32
	var l1, l2 []float32
	var word uint32
	for i, current := range sentence {
		reducedWindow = wc.r.Intn(window)
		a = c.MaxInt(0, i-window+reducedWindow)
		b = c.MinInt(sentenceLength, i+window+1-reducedWindow)
		// train skip gram model
		for j, k = a, b; j < k; j++ {
			if j == i {
				continue
			}
			word = sentence[j][0] // broken skipgram

			// HIERARCHICAL SOFTMAX
			l1 = m.Syn0[s*word:][:s]

			for d, code = range Codes[current[0]] {
				// Propagate hidden -> output
				l2 = m.Syn1[s*Points[current[0]][d]:][:s]
				f = asm.Sdot(l1, l2)
				loss += float64(f)

				if f > 6.0 {
					g = float32(1.0-code) - 1
				} else if f < -6.0 {
					g = float32(1.0-code) - 0
				} else {
					f = float32(math.Exp(float64(f)))
					g = float32(1.0-code) - f/(f+1.0)
				}
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
	wc.c = float64(sentenceLength)
	wc.e = loss / lossCount
	workChannel <- wc
}
