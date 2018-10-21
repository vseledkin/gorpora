package collect

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

func priltLines(min, max int, r io.ReadCloser, rc *zip.ReadCloser) (collectedLineCount, filteredLineCount int) {
	defer func() {
		if e := r.Close(); e != nil {
			log.Print(e)
		}
		if rc != nil {
			if e := rc.Close(); e != nil {
				log.Print(e)
			}
		}
	}()

	reader := bufio.NewReader(r)
	var L int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		L = utf8.RuneCountInString(line)
		if max >= L && L >= min {
			os.Stdout.WriteString(line)
			os.Stdout.WriteString("\n")
			collectedLineCount++
		} else {
			filteredLineCount++
		}
	}
	return
}

func Collect(min, max int, input, extension string, level int) (collectedLineCount, filteredLineCount int, e error) {
	var collected, filtered int
	var fifos []os.FileInfo
	log.Printf("Reading dir %s", input)
	if fifos, e = ioutil.ReadDir(input); e != nil {
		return 0, 0, e
	} else {
		for _, fifo := range fifos {
			if fsPath := path.Join(input, fifo.Name()); fifo.IsDir() {
				if fifo.ModTime().UnixNano() >= startTime.UnixNano() {
					return 0, 0, fmt.Errorf("Directory %s is modifyed at %s after program start %s, cannot continue", fsPath, fifo.ModTime(), startTime)
				}

				if collected, filtered, e = Collect(min, max, fsPath, extension, level+1); e != nil {
					return 0, 0, e
				} else {
					collectedLineCount += collected
					filteredLineCount += filtered
				}
				if level == 0 {
					log.Printf("%cCollected: %d Filltered: %d", 13, collectedLineCount, filteredLineCount)
				}
			} else {
				if strings.HasSuffix(strings.ToLower(fsPath), "."+extension) {
					if fifo.ModTime().UnixNano() >= startTime.UnixNano() {
						return 0, 0, fmt.Errorf("File %s is modifyed at %s after program start %s, cannot continue", fsPath, fifo.ModTime(), startTime)
					}
					var f *os.File
					if f, e = os.OpenFile(fsPath, os.O_RDONLY|os.O_EXCL, 0); e != nil {
						return 0, 0, e
					} else {
						collected, filtered = priltLines(min, max, f, nil)
						collectedLineCount += collected
						filteredLineCount += filtered
						if level == 0 {
							log.Printf("%cCollected: %d Filltered: %d", 13, collectedLineCount, filteredLineCount)
						}
					}
				} else if strings.HasSuffix(strings.ToLower(fsPath), "."+extension+".zip") {
					if fifo.ModTime().UnixNano() >= startTime.UnixNano() {
						return 0, 0, fmt.Errorf("File %s is modifyed at %s after program start %s, cannot continue", fsPath, fifo.ModTime(), startTime)
					}
					var r *zip.ReadCloser
					if r, e = zip.OpenReader(fsPath); e != nil {
						return 0, 0, e
					} else {
						if len(r.File) == 1 {
							if f, e := r.File[0].Open(); e != nil {
								r.Close()
								return 0, 0, e
							} else {
								collected, filtered = priltLines(min, max, f, r)
								collectedLineCount += collected
								filteredLineCount += filtered
								if level == 0 {
									log.Printf("%cCollected: %d Filltered: %d", 13, collectedLineCount, filteredLineCount)
								}
							}
						} else {
							r.Close()
							log.Print(fmt.Errorf("expecting one file in archive %s got %d", fsPath, len(r.File)))
						}
					}
				}
			}
		}
	}
	return
}

var CollectCommand *cobra.Command

var startTime time.Time
var input, extension, language *string
var max, min *int

func init() {
	startTime = time.Now()
	CollectCommand = &cobra.Command{
		Use:   "collect",
		Short: "collect text from files",
		Long:  "collect lines from text files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print()
			log.Println("Collect:")
			log.Printf("\tInput: %s\n", *input)
			log.Printf("\tExtension: %s\n", *extension)
			log.Printf("\tMin line length: %d\n", *min)
			log.Printf("\tMax line length: %d\n", *max)

			switch {
			case *min <= *max:
				if _, _, e := Collect(*min, *max, *input, *extension, 0); e != nil {
					log.Fatal(e)
				}
			default:
				cmd.Usage()
			}
		},
	}

	input = CollectCommand.Flags().StringP("input", "i", "", "diirectry containing text files, processed recursively")
	extension = CollectCommand.Flags().StringP("extension", "e", "txt", "extension of files to process")
	min = CollectCommand.Flags().Int("min", 0, "minimun line length expressed in utf8 chars to be accepted for output, min <= max, and > 0 ")
	max = CollectCommand.Flags().Int("max", 1e6, "maximum line length expressed in utf8 chars to be accepted for output, min <= max, and > 0")
	CollectCommand.MarkFlagRequired("input")
}
