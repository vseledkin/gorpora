package fb2

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"encoding/json"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"archive/zip"
)

// Author struct
type Author struct {
	ID         string `json:",omitempty"`
	FirstName  string `json:",omitempty"`
	MiddleName string `json:",omitempty"`
	LastName   string `json:",omitempty"`
	Nickname   string `json:",omitempty"`
	HomePage   string `json:",omitempty"`
	Email      string `json:",omitempty"`
}

// P struct
type P struct {
	Text string `json:",omitempty"`
}

type HasSections interface {
	addSection(s *Section)
}

type HasP interface {
	addP(s *P)
}

type HasTitle interface {
	addTitle(t *P)
}

type HasEpigraph interface {
	addEpigraph(t *P)
}

type HasAnnotation interface {
	addAnnotation(t *Annotation)
}

// Section struct
type Section struct {
	Title      []*P        `json:",omitempty"`
	Epigraph   []*P        `json:",omitempty"`
	P          []*P        `json:",omitempty"`
	Annotation *Annotation `json:",omitempty"`
	Section    []*Section  `json:",omitempty"`
}

func (s *Section) addAnnotation(annotation *Annotation) {
	s.Annotation = annotation
}

func (s *Section) addEpigraph(epigraph *P) {
	s.Epigraph = append(s.Epigraph, epigraph)
}

func (s *Section) addSection(section *Section) {
	s.Section = append(s.Section, section)
}

func (s *Section) addP(paragraph *P) {
	s.P = append(s.P, paragraph)
}

func (s *Section) addTitle(title *P) {
	s.Title = append(s.Title, title)
}

// Body struct
type Body struct {
	Name     string `json:",omitempty"`
	Title    []*P   `json:",omitempty"`
	Epigraph []*P
	Section  []*Section `json:",omitempty"`
}

func (o *Body) addSection(section *Section) {
	o.Section = append(o.Section, section)
}

func (o *Body) addTitle(title *P) {
	o.Title = append(o.Title, title)
}

func (o *Body) addEpigraph(epigraph *P) {
	o.Epigraph = append(o.Epigraph, epigraph)
}

func fillSection(object HasSections, startToken xml.StartElement, d *xml.Decoder) error {
	section := new(Section)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "title":
					if e = fillTitle(section, t, d); e != nil {
						return e
					}
				case "subtitle", "emphasis", "strong": // turn subtitle into p
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					section.addP(paragraph)
				case "cite": // turn subtitle into p
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					section.addP(paragraph)
				case "poem", "text-author": // turn subtitle into p
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					section.addP(paragraph)
				case "epigraph":
					if e = fillEpigraph(section, t, d); e != nil {
						return e
					}
				case "section":
					if e = fillSection(section, t, d); e != nil {
						return e
					}
				case "p":
					if e = fillP(section, t, d); e != nil {
						return e
					}
				case "empty-line", "image", "table", "tr":
					if e = skip(t, d); e != nil {
						return e
					}
				case "annotation":
					if e = fillAnnotation(section, d); e != nil {
						return e
					}

				default:
					if e = skip(t, d); e != nil {
						return e
					}
					println("Skipped: ", t.Name.Local)
					//return fmt.Errorf("fillSection: %s not implemented", t.Name.Local)
				}
			case xml.EndElement:
				if t.Name.Local == "section" {
					object.addSection(section)
					return nil
				} else {
					return fmt.Errorf("expected </%s> but got </%s>", startToken.Name.Local, t.Name.Local)
				}
			}
		}
	}
}

type Empty struct{}
type Exit struct{}

var styledTextElements = map[string]Empty{
	"a":        Empty{},
	"strong":   Empty{},
	"emphasis": Empty{},
	"sup":      Empty{},
	"sub":      Empty{},
	"style":    Empty{},
}

func extendString(x, y string) string {
	x = strings.TrimSpace(x)
	y = strings.TrimSpace(y)
	if len(x) > 0 {
		if len(y) > 0 {
			return x + " " + y
		}
		return x
	}
	return y
}

func fillP(object HasP, startToken xml.StartElement, d *xml.Decoder) error {
	paragraph := new(P)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.CharData:
				paragraph.Text = extendString(paragraph.Text, string(t))
			case xml.StartElement:
				if _, ok := styledTextElements[t.Name.Local]; ok {
					var text string
					if text, e = getText(t, d); e != nil {
						return e
					}
					paragraph.Text = extendString(paragraph.Text, text)
				} else {
					switch t.Name.Local {
					case "image":
						if e = skip(t, d); e != nil {
							return e
						}
					default:
						var text string
						if text, e = getText(t, d); e != nil {
							return e
						}
						paragraph.Text = extendString(paragraph.Text, text)
						//return fmt.Errorf("fillP: %s not implemented", t.Name.Local)
					}
				}
			case xml.EndElement:
				if t.Name.Local == startToken.Name.Local {
					object.addP(paragraph)
					return nil
				} else {
					return fmt.Errorf("expected </%s> but got </%s>", startToken.Name.Local, t.Name.Local)
				}
			}
		}
	}
}

func fillTitle(object HasTitle, startToken xml.StartElement, d *xml.Decoder) error {
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "p":
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					object.addTitle(paragraph)
				case "poem": // add poem to title p collection
					if e = fillTitle(object, t, d); e != nil {
						return e
					}
				case "empty-line": // add poem to title p collection
					if e = skip(t, d); e != nil {
						return e
					}
				default:
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					object.addTitle(paragraph)
					//return fmt.Errorf("fillTitle: %s not implemented", t.Name.Local)
				}
			case xml.EndElement:
				if t.Name.Local == startToken.Name.Local {
					return nil
				} else {
					return fmt.Errorf("expected </%s> but got </%s>", startToken.Name.Local, t.Name.Local)
				}
			}
		}
	}
}

func fillEpigraph(object HasEpigraph, startToken xml.StartElement, d *xml.Decoder) error {
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "p":
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					object.addEpigraph(paragraph)
				case "poem": // add poem to title p collection
					if e = fillEpigraph(object, t, d); e != nil {
						return e
					}
				case "stanza": // add poem to title p collection
					if e = fillEpigraph(object, t, d); e != nil {
						return e
					}
				case "v": // add poem to title p collection
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					object.addEpigraph(paragraph)
				case "text-author": // add poem to title p collection
					if e = fillEpigraph(object, t, d); e != nil {
						return e
					}
				case "empty-line": // add poem to title p collection
					if e = skip(t, d); e != nil {
						return e
					}
				default:
					//return fmt.Errorf("fillEpigraph: %s not implemented", t.Name.Local)
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					object.addEpigraph(paragraph)
				}
			case xml.EndElement:
				if t.Name.Local == startToken.Name.Local {
					return nil
				} else {
					return fmt.Errorf("expected </%s> but got </%s>", startToken.Name.Local, t.Name.Local)
				}
			}
		}
	}
}

// Annotation struct
type Annotation struct {
	ID string `json:",omitempty"`
	P  []*P   `json:",omitempty"`
}

func (a *Annotation) addP(paragraph *P) {
	a.P = append(a.P, paragraph)
}

func (a *FB2) addAnnotation(annotation *Annotation) {
	a.Annotation = annotation
}

// FB2 document
type FB2 struct {
	Title       string      `json:",omitempty"`
	Language    string      `json:",omitempty"`
	Genre       []string    `json:",omitempty"`
	Author      []*Author   `json:",omitempty"`
	Translator  []*Author   `json:",omitempty"`
	Annotation  *Annotation `json:",omitempty"`
	Keywords    string      `json:",omitempty"`
	DateAttr    string      `json:",omitempty"`
	Date        string      `json:",omitempty"`
	Sequence    string      `json:",omitempty"`
	SrcLanguage string      `json:",omitempty"`
	Body        []*Body     `json:",omitempty"`
	File        string      `json:",omitempty"`
}

func DumpP(f io.Writer, text []*P) error {
	if text == nil {
		return nil
	}
	for _, p := range text {
		if p == nil {
			continue
		}
		if _, e := f.Write([]byte(p.Text + endl)); e != nil {
			return e
		}
	}
	return nil
}

func DumpSection(f io.Writer, s *Section) (e error) {

	if e = DumpP(f, s.Title); e != nil {
		return e
	}

	if s.Annotation != nil {
		if e = DumpP(f, s.Annotation.P); e != nil {
			return e
		}
	}

	if e = DumpP(f, s.Epigraph); e != nil {
		return e
	}

	if e = DumpP(f, s.Epigraph); e != nil {
		return e
	}

	if e = DumpP(f, s.P); e != nil {
		return e
	}

	//if s.Section != nil && s.P != nil {
	//	return fmt.Errorf("section has both P s and sections, order is not preserved")
	//} else {
	if s.Section != nil {
		for _, section := range s.Section {
			if e = DumpSection(f, section); e != nil {
				return e
			}
		}
	} //else {
	//if e = DumpP(f, s.P); e != nil {
	//	return e
	//}
	//}
	//}
	return nil
}

func (b *FB2) Dump() error {
	if file, e := os.Create(b.File); e != nil {
		return e
	} else {
		defer file.Close()
		// write title
		file.WriteString(b.Title)
		file.WriteString(endl)

		// write title
		if b.Annotation != nil {
			for _, a := range b.Annotation.P {
				file.WriteString(a.Text)
				file.WriteString(endl)
			}
		}
		// write body
		for _, body := range b.Body {
			for _, s := range body.Section {
				if e = DumpSection(file, s); e != nil {
					return e
				}

			}
		}

	}
	return nil
}

func (b *FB2) String() string {
	if bits, e := json.MarshalIndent(b, "", " "); e != nil {
		return e.Error()
	} else {
		return string(bits)
	}

}

func getText(parent xml.StartElement, d *xml.Decoder) (text string, e error) {
	for {
		var token xml.Token
		if token, e = d.Token(); e != nil {
			return "", e
		}
		switch t := token.(type) {
		case xml.CharData:
			text = extendString(text, string(t))
		case xml.StartElement:

			var innerText string
			if innerText, e = getText(t, d); e != nil {
				return "", e
			}
			text = extendString(text, innerText)

		case xml.EndElement:
			if t.Name.Local == parent.Name.Local {
				return
			}
			return "", fmt.Errorf("expected </%s> tag but got %+v", parent.Name.Local, t.Name.Local)
		default:
			return "", fmt.Errorf("unexpected token %s when parsing <...>text</...>; chardata or ending tag expected", token)
		}
	}
	return
}

func skip(skipToken xml.StartElement, d *xml.Decoder) (e error) {
	for {
		var token xml.Token
		if token, e = d.Token(); e != nil {
			return
		}
		switch t := token.(type) {
		case xml.EndElement:
			if t.Name.Local == skipToken.Name.Local {
				return
			}
		}
	}
	return fmt.Errorf("cannot find ending tag for %s", skipToken.Name.Local)
}

func (b *FB2) fillAuthor(d *xml.Decoder) error {
	author := new(Author)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {

			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "first-name":
					if author.FirstName, e = getText(t, d); e != nil {
						return e
					}
				case "middle-name":
					if author.MiddleName, e = getText(t, d); e != nil {
						return e
					}
				case "last-name":
					if author.LastName, e = getText(t, d); e != nil {
						return e
					}
				case "nickname":
					if author.Nickname, e = getText(t, d); e != nil {
						return e
					}
				case "id":
					if author.ID, e = getText(t, d); e != nil {
						return e
					}
				case "home-page":
					if author.HomePage, e = getText(t, d); e != nil {
						return e
					}
				case "email":
					if author.Email, e = getText(t, d); e != nil {
						return e
					}
				default:
					return fmt.Errorf("fillAuthor: %s in FB2 not implemented", token)
				}
			case xml.EndElement:
				if t.Name.Local == "author" {
					b.Author = append(b.Author, author)
					return nil
				} else {
					return fmt.Errorf("expected </author> but got </%s>", t.Name.Local)
				}
			}
		}
	}
}

func (b *FB2) fillTranslator(d *xml.Decoder) error {
	author := new(Author)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {

			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "first-name":
					if author.FirstName, e = getText(t, d); e != nil {
						return e
					}
				case "middle-name":
					if author.MiddleName, e = getText(t, d); e != nil {
						return e
					}
				case "last-name":
					if author.LastName, e = getText(t, d); e != nil {
						return e
					}
				case "nickname":
					if author.Nickname, e = getText(t, d); e != nil {
						return e
					}
				case "id":
					if author.ID, e = getText(t, d); e != nil {
						return e
					}
				case "home-page":
					if author.HomePage, e = getText(t, d); e != nil {
						return e
					}
				case "email":
					if author.Email, e = getText(t, d); e != nil {
						return e
					}
				default:
					return fmt.Errorf("fillTranslator: %s in FB2 not implemented", token)
				}
			case xml.EndElement:
				if t.Name.Local == "translator" {
					b.Translator = append(b.Translator, author)
					return nil
				} else {
					return fmt.Errorf("expected </translator> but got </%s>", t.Name.Local)
				}
			}
		}
	}
}

func (b *FB2) fillBody(d *xml.Decoder) error {
	body := new(Body)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {

			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "title":
					if e = fillTitle(body, t, d); e != nil {
						return e
					}
				case "section":
					if e = fillSection(body, t, d); e != nil {
						return e
					}
				case "image", "empty-line":
					if e = skip(t, d); e != nil {
						return e
					}
				case "epigraph":
					if e = fillEpigraph(body, t, d); e != nil {
						return e
					}
				default:
					//return fmt.Errorf("fillBody: %s in FB2 not implemented", t.Name.Local)
					if e = skip(t, d); e != nil {
						return e
					}

				}
			case xml.EndElement:
				if t.Name.Local == "body" {
					b.Body = append(b.Body, body)
					return nil
				} else {
					return fmt.Errorf("fillBody: expected </body> but got </%s>", t.Name.Local)
				}
			}
		}
	}
}

func fillAnnotation(object HasAnnotation, d *xml.Decoder) error {
	annotation := new(Annotation)
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				if _, ok := styledTextElements[t.Name.Local]; ok {
					paragraph := new(P)
					if paragraph.Text, e = getText(t, d); e != nil {
						return e
					}
					annotation.P = append(annotation.P, paragraph)
				} else {
					switch t.Name.Local {
					case "p":
						if e = fillP(annotation, t, d); e != nil {
							return e
						}
					case "empty-line":
						if e = skip(t, d); e != nil {
							return e
						}
					default:
						//return fmt.Errorf("fillP: %s not implemented", t.Name.Local)
						paragraph := new(P)
						if paragraph.Text, e = getText(t, d); e != nil {
							return e
						}
						annotation.P = append(annotation.P, paragraph)
					}
				}
			case xml.EndElement:
				if t.Name.Local == "annotation" {
					object.addAnnotation(annotation)
					return nil
				} else {
					return fmt.Errorf("expected </annotation> but got </%s>", t.Name.Local)
				}
			}
		}
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	if charset == "windows-1251" {
		return transform.NewReader(input, charmap.Windows1251.NewDecoder()), nil
	} else if charset == "iso-8859-1" {
		return transform.NewReader(input, charmap.Windows1252.NewDecoder()), nil
	}
	return nil, fmt.Errorf("unsupported charset: %q", charset)
}

func (b *FB2) fillTitleInfo(d *xml.Decoder) error {
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "genre":
					if genre, e := getText(t, d); e != nil {
						return e
					} else {
						b.Genre = append(b.Genre, genre)
					}
				case "author":
					if e := b.fillAuthor(d); e != nil {
						return e
					}
				case "translator":
					if e := b.fillTranslator(d); e != nil {
						return e
					}
				case "book-title":
					if b.Title, e = getText(t, d); e != nil {
						return e
					}
				case "lang":
					if b.Language, e = getText(t, d); e != nil {
						return e
					}
				case "src-lang":
					if b.Language, e = getText(t, d); e != nil {
						return e
					}
				case "image":
					skip(t, d)
				case "coverpage":
					skip(t, d)
				case "annotation":
					if e := fillAnnotation(b, d); e != nil {
						return e
					}
				case "keywords":
					if b.Keywords, e = getText(t, d); e != nil {
						return e
					}
				case "sequence":
					if len(t.Attr) == 1 {
						if t.Attr[0].Name.Local == "name" {
							b.Sequence = t.Attr[0].Value
						} else {
							return fmt.Errorf("unexpected attribute %s", t.Attr[0].Name)
						}
					}
					// withdraw sequence
					if e = skip(t, d); e != nil {
						return e
					}
					//var sequenceStr string
					//if sequenceStr, e = getText(t, d); e != nil {
					//	return e
					//}
					//log.Printf("Sequence: a{%s}-t{%s}", b.Sequence, sequenceStr)
				case "date":
					if len(t.Attr) == 1 {
						if t.Attr[0].Name.Local == "value" {
							b.DateAttr = t.Attr[0].Value

						} else {
							return fmt.Errorf("unexpected attribute %s", t.Attr[0].Name)
						}
					}
					if b.Date, e = getText(t, d); e != nil {
						return e
					}
				}
			case xml.EndElement:
				if t.Name.Local == "title-info" {
					return nil
				} else {
					return fmt.Errorf("expected </title-info> but got </%s>", t.Name.Local)
				}
			}
		}
	}

	return nil
}

func (b *FB2) fillDescription(d *xml.Decoder) error {
	for {
		if token, e := d.Token(); e != nil {
			return e
		} else {
			switch t := token.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "title-info":
					if e = b.fillTitleInfo(d); e != nil {
						return e
					}
				case "src-title-info":
					if e = skip(t, d); e != nil {
						return e
					}
				case "document-info":
					if e = skip(t, d); e != nil {
						return e
					}
				case "publish-info":
					if e = skip(t, d); e != nil {
						return e
					}
				case "custom-info":
					if e = skip(t, d); e != nil {
						return e
					}
				default:
					if e = skip(t, d); e != nil {
						return e
					}
				}
			case xml.EndElement:
				if t.Name.Local == "description" {
					return nil
				} else {
					return fmt.Errorf("expected </description> but got </%s>", t.Name.Local)
				}
			}
		}
	}
	return nil
}

func parseFB2(r io.ReadCloser) (*FB2, error) {
	decoder := xml.NewDecoder(r)

	decoder.CharsetReader = charsetReader

	book := new(FB2)
	// read
	for {
		if token, e := decoder.Token(); e != nil {
			if e == io.EOF {
				break
			}
			return nil, e
		} else {
			if t, ok := token.(xml.StartElement); ok {
				switch t.Name.Local {
				case "description":
					if e = book.fillDescription(decoder); e != nil {
						return nil, e
					}
				case "body":
					if e = book.fillBody(decoder); e != nil {
						return nil, e
					}
				}
			}
		}
	}
	return book, nil
}

func work(fsPath string, output chan interface{}, license chan Empty, r io.ReadCloser, rc *zip.ReadCloser) {
	defer func() {
		if e := r.Close(); e != nil {
			output <- e
		}
		if rc != nil {
			if e := rc.Close(); e != nil {
				output <- e
			}
		}
	}()

	defer func(license chan Empty) {
		license <- Empty{}
	}(license)

	if book, e := parseFB2(r); e != nil {
		output <- e
	} else {
		book.File = fsPath + ".txt"
		output <- book
	}
}

func parseDir(input string, output chan interface{}, licence chan Empty) {

	if fifos, e := ioutil.ReadDir(input); e != nil {
		output <- e
	} else {
		for _, fifo := range fifos {
			if fsPath := path.Join(input, fifo.Name()); fifo.IsDir() {
				parseDir(fsPath, output, licence)
			} else {
				if strings.HasSuffix(strings.ToLower(fifo.Name()), ".fb2") {
					if f, e := os.Open(fsPath); e != nil {
						output <- e
					} else {
						log.Printf("-> %s\n", fsPath)
						<-licence
						go work(fsPath, output, licence, f, nil)
					}
				} else if strings.HasSuffix(strings.ToLower(fifo.Name()), ".fb2.zip") {
					if r, e := zip.OpenReader(fsPath); e != nil {
						output <- e
					} else {
						if len(r.File) == 1 {
							if f, e := r.File[0].Open(); e != nil {
								r.Close()
								output <- e
							} else {
								log.Printf("%s\n", fsPath)
								<-licence
								go work(fsPath, output, licence, f, r)
							}
						} else {
							r.Close()
							output <- fmt.Errorf("expecting one file in archive %s got %d", fsPath, len(r.File))
						}
					}
				}
			}
		}
	}
}

var endl string;

func ConvertFB2text(inDir string, paragraphEnding int, threads int) {

	if paragraphEnding < 1 {
		paragraphEnding = 1
	}
	if threads < 1 {
		threads = 1
	}
	endl = strings.Repeat("\n", paragraphEnding)

	license := make(chan Empty, threads)
	for i := 0; i < threads; i++ {
		license <- Empty{}
	}

	output := make(chan interface{}, threads)
	go func(output chan interface{}, licence chan Empty) {
		parseDir(inDir, output, licence)
		output <- Exit{}

	}(output, license)

	for result := range output {
		switch t := result.(type) {
		case error:
			log.Print(t)
		case *FB2:
			t.Dump()
		case Exit:
			for i := 0; i < threads; i++ {
				<-license
			}
			close(output)
		}
	}
}
