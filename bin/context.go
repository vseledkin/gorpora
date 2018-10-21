package gorpora

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// delimEnds maps each delim to a string of characters that terminate it.
var delimEnds = [...]string{
	delimDoubleQuote: `"`,
	delimSingleQuote: "'",
	// Determined empirically by running the below in various browsers.
	// var div = document.createElement("DIV");
	// for (var i = 0; i < 0x10000; ++i) {
	//   div.innerHTML = "<span title=x" + String.fromCharCode(i) + "-bar>";
	//   if (div.getElementsByTagName("SPAN")[0].title.indexOf("bar") < 0)
	//     document.write("<p>U+" + i.toString(16));
	// }
	delimSpaceOrTagEnd: " \t\n\f\r>",
}

// context describes the state an HTML parser must be in when it reaches the
// portion of HTML produced by evaluating a particular template node.
//
// The zero value of type context is the start context for a template that
// produces an HTML fragment as defined at
// http://www.w3.org/TR/html5/syntax.html#the-end
// where the context element is null.
type context struct {
	state   state
	delim   delim
	urlPart urlPart
	jsCtx   jsCtx
	attr    attr
	element element
	err     *error
}

func (c context) String() string {
	return fmt.Sprintf("{%v %v %v %v %v %v %v}", c.state, c.delim, c.urlPart, c.jsCtx, c.attr, c.element, c.err)
}

// eq reports whether two contexts are equal.
func (c context) eq(d context) bool {
	return c.state == d.state &&
		c.delim == d.delim &&
		c.urlPart == d.urlPart &&
		c.jsCtx == d.jsCtx &&
		c.attr == d.attr &&
		c.element == d.element &&
		c.err == d.err
}

// mangle produces an identifier that includes a suffix that distinguishes it
// from template names mangled with different contexts.
func (c context) mangle(templateName string) string {
	// The mangled name for the default context is the input templateName.
	if c.state == stateText {
		return templateName
	}
	s := templateName + "$htmltemplate_" + c.state.String()
	if c.delim != 0 {
		s += "_" + c.delim.String()
	}
	if c.urlPart != 0 {
		s += "_" + c.urlPart.String()
	}
	if c.jsCtx != 0 {
		s += "_" + c.jsCtx.String()
	}
	if c.attr != 0 {
		s += "_" + c.attr.String()
	}
	if c.element != 0 {
		s += "_" + c.element.String()
	}
	return s
}

// state describes a high-level HTML parser state.
//
// It bounds the top of the element stack, and by extension the HTML insertion
// mode, but also contains state that does not correspond to anything in the
// HTML5 parsing algorithm because a single token production in the HTML
// grammar may contain embedded actions in a template. For instance, the quoted
// HTML attribute produced by
//     <div title="Hello {{.World}}">
// is a single token in HTML's grammar but in a template spans several nodes.
type state uint8

const (
	// stateText is parsed character data. An HTML parser is in
	// this state when its parse position is outside an HTML tag,
	// directive, comment, and special element body.
	stateText state = iota
	// stateTag occurs before an HTML attribute or the end of a tag.
	stateTag
	// stateAttrName occurs inside an attribute name.
	// It occurs between the ^'s in ` ^name^ = value`.
	stateAttrName
	// stateAfterName occurs after an attr name has ended but before any
	// equals sign. It occurs between the ^'s in ` name^ ^= value`.
	stateAfterName
	// stateBeforeValue occurs after the equals sign but before the value.
	// It occurs between the ^'s in ` name =^ ^value`.
	stateBeforeValue
	// stateHTMLCmt occurs inside an <!-- HTML comment -->.
	stateHTMLCmt
	// stateRCDATA occurs inside an RCDATA element (<textarea> or <title>)
	// as described at http://www.w3.org/TR/html5/syntax.html#elements-0
	stateRCDATA
	// stateAttr occurs inside an HTML attribute whose content is text.
	stateAttr
	// stateURL occurs inside an HTML attribute whose content is a URL.
	stateURL
	// stateJS occurs inside an event handler or script element.
	stateJS
	// stateJSDqStr occurs inside a JavaScript double quoted string.
	stateJSDqStr
	// stateJSSqStr occurs inside a JavaScript single quoted string.
	stateJSSqStr
	// stateJSRegexp occurs inside a JavaScript regexp literal.
	stateJSRegexp
	// stateJSBlockCmt occurs inside a JavaScript /* block comment */.
	stateJSBlockCmt
	// stateJSLineCmt occurs inside a JavaScript // line comment.
	stateJSLineCmt
	// stateCSS occurs inside a <style> element or style attribute.
	stateCSS
	// stateCSSDqStr occurs inside a CSS double quoted string.
	stateCSSDqStr
	// stateCSSSqStr occurs inside a CSS single quoted string.
	stateCSSSqStr
	// stateCSSDqURL occurs inside a CSS double quoted url("...").
	stateCSSDqURL
	// stateCSSSqURL occurs inside a CSS single quoted url('...').
	stateCSSSqURL
	// stateCSSURL occurs inside a CSS unquoted url(...).
	stateCSSURL
	// stateCSSBlockCmt occurs inside a CSS /* block comment */.
	stateCSSBlockCmt
	// stateCSSLineCmt occurs inside a CSS // line comment.
	stateCSSLineCmt
	// stateError is an infectious error state outside any valid
	// HTML/CSS/JS construct.
	stateError
)

var stateNames = [...]string{
	stateText:        "stateText",
	stateTag:         "stateTag",
	stateAttrName:    "stateAttrName",
	stateAfterName:   "stateAfterName",
	stateBeforeValue: "stateBeforeValue",
	stateHTMLCmt:     "stateHTMLCmt",
	stateRCDATA:      "stateRCDATA",
	stateAttr:        "stateAttr",
	stateURL:         "stateURL",
	stateJS:          "stateJS",
	stateJSDqStr:     "stateJSDqStr",
	stateJSSqStr:     "stateJSSqStr",
	stateJSRegexp:    "stateJSRegexp",
	stateJSBlockCmt:  "stateJSBlockCmt",
	stateJSLineCmt:   "stateJSLineCmt",
	stateCSS:         "stateCSS",
	stateCSSDqStr:    "stateCSSDqStr",
	stateCSSSqStr:    "stateCSSSqStr",
	stateCSSDqURL:    "stateCSSDqURL",
	stateCSSSqURL:    "stateCSSSqURL",
	stateCSSURL:      "stateCSSURL",
	stateCSSBlockCmt: "stateCSSBlockCmt",
	stateCSSLineCmt:  "stateCSSLineCmt",
	stateError:       "stateError",
}

func (s state) String() string {
	if int(s) < len(stateNames) {
		return stateNames[s]
	}
	return fmt.Sprintf("illegal state %d", int(s))
}

// isComment is true for any state that contains content meant for template
// authors & maintainers, not for end-users or machines.
func isComment(s state) bool {
	switch s {
	case stateHTMLCmt, stateJSBlockCmt, stateJSLineCmt, stateCSSBlockCmt, stateCSSLineCmt:
		return true
	}
	return false
}

// isInTag return whether s occurs solely inside an HTML tag.
func isInTag(s state) bool {
	switch s {
	case stateTag, stateAttrName, stateAfterName, stateBeforeValue, stateAttr:
		return true
	}
	return false
}

// delim is the delimiter that will end the current HTML attribute.
type delim uint8

const (
	// delimNone occurs outside any attribute.
	delimNone delim = iota
	// delimDoubleQuote occurs when a double quote (") closes the attribute.
	delimDoubleQuote
	// delimSingleQuote occurs when a single quote (') closes the attribute.
	delimSingleQuote
	// delimSpaceOrTagEnd occurs when a space or right angle bracket (>)
	// closes the attribute.
	delimSpaceOrTagEnd
)

var delimNames = [...]string{
	delimNone:          "delimNone",
	delimDoubleQuote:   "delimDoubleQuote",
	delimSingleQuote:   "delimSingleQuote",
	delimSpaceOrTagEnd: "delimSpaceOrTagEnd",
}

func (d delim) String() string {
	if int(d) < len(delimNames) {
		return delimNames[d]
	}
	return fmt.Sprintf("illegal delim %d", int(d))
}

// urlPart identifies a part in an RFC 3986 hierarchical URL to allow different
// encoding strategies.
type urlPart uint8

const (
	// urlPartNone occurs when not in a URL, or possibly at the start:
	// ^ in "^http://auth/path?k=v#frag".
	urlPartNone urlPart = iota
	// urlPartPreQuery occurs in the scheme, authority, or path; between the
	// ^s in "h^ttp://auth/path^?k=v#frag".
	urlPartPreQuery
	// urlPartQueryOrFrag occurs in the query portion between the ^s in
	// "http://auth/path?^k=v#frag^".
	urlPartQueryOrFrag
	// urlPartUnknown occurs due to joining of contexts both before and
	// after the query separator.
	urlPartUnknown
)

var urlPartNames = [...]string{
	urlPartNone:        "urlPartNone",
	urlPartPreQuery:    "urlPartPreQuery",
	urlPartQueryOrFrag: "urlPartQueryOrFrag",
	urlPartUnknown:     "urlPartUnknown",
}

func (u urlPart) String() string {
	if int(u) < len(urlPartNames) {
		return urlPartNames[u]
	}
	return fmt.Sprintf("illegal urlPart %d", int(u))
}

// jsCtx determines whether a '/' starts a regular expression literal or a
// division operator.
type jsCtx uint8

const (
	// jsCtxRegexp occurs where a '/' would start a regexp literal.
	jsCtxRegexp jsCtx = iota
	// jsCtxDivOp occurs where a '/' would start a division operator.
	jsCtxDivOp
	// jsCtxUnknown occurs where a '/' is ambiguous due to context joining.
	jsCtxUnknown
)

func (c jsCtx) String() string {
	switch c {
	case jsCtxRegexp:
		return "jsCtxRegexp"
	case jsCtxDivOp:
		return "jsCtxDivOp"
	case jsCtxUnknown:
		return "jsCtxUnknown"
	}
	return fmt.Sprintf("illegal jsCtx %d", int(c))
}

// element identifies the HTML element when inside a start tag or special body.
// Certain HTML element (for example <script> and <style>) have bodies that are
// treated differently from stateText so the element type is necessary to
// transition into the correct context at the end of a tag and to identify the
// end delimiter for the body.
type element uint8

const (
	// elementNone occurs outside a special tag or special element body.
	elementNone element = iota
	// elementScript corresponds to the raw text <script> element
	// with JS MIME type or no type attribute.
	elementScript
	// elementStyle corresponds to the raw text <style> element.
	elementStyle
	// elementTextarea corresponds to the RCDATA <textarea> element.
	elementTextarea
	// elementTitle corresponds to the RCDATA <title> element.
	elementTitle
)

var elementNames = [...]string{
	elementNone:     "elementNone",
	elementScript:   "elementScript",
	elementStyle:    "elementStyle",
	elementTextarea: "elementTextarea",
	elementTitle:    "elementTitle",
}

func (e element) String() string {
	if int(e) < len(elementNames) {
		return elementNames[e]
	}
	return fmt.Sprintf("illegal element %d", int(e))
}

// attr identifies the current HTML attribute when inside the attribute,
// that is, starting from stateAttrName until stateTag/stateText (exclusive).
type attr uint8

const (
	// attrNone corresponds to a normal attribute or no attribute.
	attrNone attr = iota
	// attrScript corresponds to an event handler attribute.
	attrScript
	// attrScriptType corresponds to the type attribute in script HTML element
	attrScriptType
	// attrStyle corresponds to the style attribute whose value is CSS.
	attrStyle
	// attrURL corresponds to an attribute whose value is a URL.
	attrURL
)

var attrNames = [...]string{
	attrNone:       "attrNone",
	attrScript:     "attrScript",
	attrScriptType: "attrScriptType",
	attrStyle:      "attrStyle",
	attrURL:        "attrURL",
}

func (a attr) String() string {
	if int(a) < len(attrNames) {
		return attrNames[a]
	}
	return fmt.Sprintf("illegal attr %d", int(a))
}

// transitionFunc is the array of context transition functions for text nodes.
// A transition function takes a context and template text input, and returns
// the updated context and the number of bytes consumed from the front of the
// input.
var transitionFunc = [...]func(context, []byte) (context, int){
	stateText:        tText,
	stateTag:         tTag,
	stateAttrName:    tAttrName,
	stateAfterName:   tAfterName,
	stateBeforeValue: tBeforeValue,
	stateHTMLCmt:     tHTMLCmt,
	stateRCDATA:      tSpecialTagEnd,
	stateAttr:        tAttr,
	stateURL:         tURL,
	stateJS:          tJS,
	stateJSDqStr:     tJSDelimited,
	stateJSSqStr:     tJSDelimited,
	stateJSRegexp:    tJSDelimited,
	stateJSBlockCmt:  tBlockCmt,
	stateJSLineCmt:   tLineCmt,
	stateCSS:         tCSS,
	stateCSSDqStr:    tCSSStr,
	stateCSSSqStr:    tCSSStr,
	stateCSSDqURL:    tCSSStr,
	stateCSSSqURL:    tCSSStr,
	stateCSSURL:      tCSSStr,
	stateCSSBlockCmt: tBlockCmt,
	stateCSSLineCmt:  tLineCmt,
	stateError:       tError,
}

var commentStart = []byte("<!--")
var commentEnd = []byte("-->")

// tText is the context transition function for the text state.
func tText(c context, s []byte) (context, int) {
	k := 0
	for {
		i := k + bytes.IndexByte(s[k:], '<')
		if i < k || i+1 == len(s) {
			return c, len(s)
		} else if i+4 <= len(s) && bytes.Equal(commentStart, s[i:i+4]) {
			return context{state: stateHTMLCmt}, i + 4
		}
		i++
		end := false
		if s[i] == '/' {
			if i+1 == len(s) {
				return c, len(s)
			}
			end, i = true, i+1
		}
		j, e := eatTagName(s, i)
		if j != i {
			if end {
				e = elementNone
			}
			// We've found an HTML tag.
			return context{state: stateTag, element: e}, j
		}
		k = j
	}
}

var elementContentType = [...]state{
	elementNone:     stateText,
	elementScript:   stateJS,
	elementStyle:    stateCSS,
	elementTextarea: stateRCDATA,
	elementTitle:    stateRCDATA,
}

// tTag is the context transition function for the tag state.
func tTag(c context, s []byte) (context, int) {
	// Find the attribute name.
	i := eatWhiteSpace(s, 0)
	if i == len(s) {
		return c, len(s)
	}
	if s[i] == '>' {
		return context{
			state:   elementContentType[c.element],
			element: c.element,
		}, i + 1
	}
	j, err := eatAttrName(s, i)
	if err != nil {
		return context{state: stateError, err: err}, len(s)
	}
	state, attr := stateTag, attrNone
	if i == j {
		e := fmt.Errorf("expected space, attr name, or end of tag, but got %q", s[i:])
		return context{
			state: stateError,
			err:   &e,
		}, len(s)
	}

	attrName := string(s[i:j])
	if c.element == elementScript && attrName == "type" {
		attr = attrScriptType
	} else {
		switch attrType(attrName) {
		case contentTypeURL:
			attr = attrURL
		case contentTypeCSS:
			attr = attrStyle
		case contentTypeJS:
			attr = attrScript
		}
	}

	if j == len(s) {
		state = stateAttrName
	} else {
		state = stateAfterName
	}
	return context{state: state, element: c.element, attr: attr}, j
}

// tAttrName is the context transition function for stateAttrName.
func tAttrName(c context, s []byte) (context, int) {
	i, err := eatAttrName(s, 0)
	if err != nil {
		return context{state: stateError, err: err}, len(s)
	} else if i != len(s) {
		c.state = stateAfterName
	}
	return c, i
}

// tAfterName is the context transition function for stateAfterName.
func tAfterName(c context, s []byte) (context, int) {
	// Look for the start of the value.
	i := eatWhiteSpace(s, 0)
	if i == len(s) {
		return c, len(s)
	} else if s[i] != '=' {
		// Occurs due to tag ending '>', and valueless attribute.
		c.state = stateTag
		return c, i
	}
	c.state = stateBeforeValue
	// Consume the "=".
	return c, i + 1
}

var attrStartStates = [...]state{
	attrNone:       stateAttr,
	attrScript:     stateJS,
	attrScriptType: stateAttr,
	attrStyle:      stateCSS,
	attrURL:        stateURL,
}

// tBeforeValue is the context transition function for stateBeforeValue.
func tBeforeValue(c context, s []byte) (context, int) {
	i := eatWhiteSpace(s, 0)
	if i == len(s) {
		return c, len(s)
	}
	// Find the attribute delimiter.
	delim := delimSpaceOrTagEnd
	switch s[i] {
	case '\'':
		delim, i = delimSingleQuote, i+1
	case '"':
		delim, i = delimDoubleQuote, i+1
	}
	c.state, c.delim = attrStartStates[c.attr], delim
	return c, i
}

// tHTMLCmt is the context transition function for stateHTMLCmt.
func tHTMLCmt(c context, s []byte) (context, int) {
	if i := bytes.Index(s, commentEnd); i != -1 {
		return context{}, i + 3
	}
	return c, len(s)
}

// specialTagEndMarkers maps element types to the character sequence that
// case-insensitively signals the end of the special tag body.
var specialTagEndMarkers = [...][]byte{
	elementScript:   []byte("script"),
	elementStyle:    []byte("style"),
	elementTextarea: []byte("textarea"),
	elementTitle:    []byte("title"),
}

var (
	specialTagEndPrefix = []byte("</")
	tagEndSeparators    = []byte("> \t\n\f/")
)

// tSpecialTagEnd is the context transition function for raw text and RCDATA
// element states.
func tSpecialTagEnd(c context, s []byte) (context, int) {
	if c.element != elementNone {
		if i := indexTagEnd(s, specialTagEndMarkers[c.element]); i != -1 {
			return context{}, i
		}
	}
	return c, len(s)
}

// indexTagEnd finds the index of a special tag end in a case insensitive way, or returns -1
func indexTagEnd(s []byte, tag []byte) int {
	res := 0
	plen := len(specialTagEndPrefix)
	for len(s) > 0 {
		// Try to find the tag end prefix first
		i := bytes.Index(s, specialTagEndPrefix)
		if i == -1 {
			return i
		}
		s = s[i+plen:]
		// Try to match the actual tag if there is still space for it
		if len(tag) <= len(s) && bytes.EqualFold(tag, s[:len(tag)]) {
			s = s[len(tag):]
			// Check the tag is followed by a proper separator
			if len(s) > 0 && bytes.IndexByte(tagEndSeparators, s[0]) != -1 {
				return res + i
			}
			res += len(tag)
		}
		res += i + plen
	}
	return -1
}

// tAttr is the context transition function for the attribute state.
func tAttr(c context, s []byte) (context, int) {
	return c, len(s)
}

// tURL is the context transition function for the URL state.
func tURL(c context, s []byte) (context, int) {
	if bytes.IndexAny(s, "#?") >= 0 {
		c.urlPart = urlPartQueryOrFrag
	} else if len(s) != eatWhiteSpace(s, 0) && c.urlPart == urlPartNone {
		// HTML5 uses "Valid URL potentially surrounded by spaces" for
		// attrs: http://www.w3.org/TR/html5/index.html#attributes-1
		c.urlPart = urlPartPreQuery
	}
	return c, len(s)
}

// tJS is the context transition function for the JS state.
func tJS(c context, s []byte) (context, int) {
	i := bytes.IndexAny(s, `"'/`)
	if i == -1 {
		// Entire input is non string, comment, regexp tokens.
		c.jsCtx = nextJSCtx(s, c.jsCtx)
		return c, len(s)
	}
	c.jsCtx = nextJSCtx(s[:i], c.jsCtx)
	switch s[i] {
	case '"':
		c.state, c.jsCtx = stateJSDqStr, jsCtxRegexp
	case '\'':
		c.state, c.jsCtx = stateJSSqStr, jsCtxRegexp
	case '/':
		switch {
		case i+1 < len(s) && s[i+1] == '/':
			c.state, i = stateJSLineCmt, i+1
		case i+1 < len(s) && s[i+1] == '*':
			c.state, i = stateJSBlockCmt, i+1
		case c.jsCtx == jsCtxRegexp:
			c.state = stateJSRegexp
		case c.jsCtx == jsCtxDivOp:
			c.jsCtx = jsCtxRegexp
		default:
			e := fmt.Errorf("'/' could start a division or regexp: %.32q", s[i:])
			return context{
				state: stateError,
				err:   &e,
			}, len(s)
		}
	default:
		panic("unreachable")
	}
	return c, i + 1
}

// tJSDelimited is the context transition function for the JS string and regexp
// states.
func tJSDelimited(c context, s []byte) (context, int) {
	specials := `\"`
	switch c.state {
	case stateJSSqStr:
		specials = `\'`
	case stateJSRegexp:
		specials = `\/[]`
	}

	k, inCharset := 0, false
	for {
		i := k + bytes.IndexAny(s[k:], specials)
		if i < k {
			break
		}
		switch s[i] {
		case '\\':
			i++
			if i == len(s) {
				e := fmt.Errorf("unfinished escape sequence in JS string: %q", s)
				return context{
					state: stateError,
					err:   &e,
				}, len(s)
			}
		case '[':
			inCharset = true
		case ']':
			inCharset = false
		default:
			// end delimiter
			if !inCharset {
				c.state, c.jsCtx = stateJS, jsCtxDivOp
				return c, i + 1
			}
		}
		k = i + 1
	}

	if inCharset {
		// This can be fixed by making context richer if interpolation
		// into charsets is desired.
		e := fmt.Errorf("unfinished JS regexp charset: %q", s)
		return context{
			state: stateError,
			err:   &e,
		}, len(s)
	}

	return c, len(s)
}

var blockCommentEnd = []byte("*/")

// tBlockCmt is the context transition function for /*comment*/ states.
func tBlockCmt(c context, s []byte) (context, int) {
	i := bytes.Index(s, blockCommentEnd)
	if i == -1 {
		return c, len(s)
	}
	switch c.state {
	case stateJSBlockCmt:
		c.state = stateJS
	case stateCSSBlockCmt:
		c.state = stateCSS
	default:
		panic(c.state.String())
	}
	return c, i + 2
}

// tLineCmt is the context transition function for //comment states.
func tLineCmt(c context, s []byte) (context, int) {
	var lineTerminators string
	var endState state
	switch c.state {
	case stateJSLineCmt:
		lineTerminators, endState = "\n\r\u2028\u2029", stateJS
	case stateCSSLineCmt:
		lineTerminators, endState = "\n\f\r", stateCSS
		// Line comments are not part of any published CSS standard but
		// are supported by the 4 major browsers.
		// This defines line comments as
		//     LINECOMMENT ::= "//" [^\n\f\d]*
		// since http://www.w3.org/TR/css3-syntax/#SUBTOK-nl defines
		// newlines:
		//     nl ::= #xA | #xD #xA | #xD | #xC
	default:
		panic(c.state.String())
	}

	i := bytes.IndexAny(s, lineTerminators)
	if i == -1 {
		return c, len(s)
	}
	c.state = endState
	// Per section 7.4 of EcmaScript 5 : http://es5.github.com/#x7.4
	// "However, the LineTerminator at the end of the line is not
	// considered to be part of the single-line comment; it is
	// recognized separately by the lexical grammar and becomes part
	// of the stream of input elements for the syntactic grammar."
	return c, i
}

// tCSS is the context transition function for the CSS state.
func tCSS(c context, s []byte) (context, int) {
	// CSS quoted strings are almost never used except for:
	// (1) URLs as in background: "/foo.png"
	// (2) Multiword font-names as in font-family: "Times New Roman"
	// (3) List separators in content values as in inline-lists:
	//    <style>
	//    ul.inlineList { list-style: none; padding:0 }
	//    ul.inlineList > li { display: inline }
	//    ul.inlineList > li:before { content: ", " }
	//    ul.inlineList > li:first-child:before { content: "" }
	//    </style>
	//    <ul class=inlineList><li>One<li>Two<li>Three</ul>
	// (4) Attribute value selectors as in a[href="http://example.com/"]
	//
	// We conservatively treat all strings as URLs, but make some
	// allowances to avoid confusion.
	//
	// In (1), our conservative assumption is justified.
	// In (2), valid font names do not contain ':', '?', or '#', so our
	// conservative assumption is fine since we will never transition past
	// urlPartPreQuery.
	// In (3), our protocol heuristic should not be tripped, and there
	// should not be non-space content after a '?' or '#', so as long as
	// we only %-encode RFC 3986 reserved characters we are ok.
	// In (4), we should URL escape for URL attributes, and for others we
	// have the attribute name available if our conservative assumption
	// proves problematic for real code.

	k := 0
	for {
		i := k + bytes.IndexAny(s[k:], `("'/`)
		if i < k {
			return c, len(s)
		}
		switch s[i] {
		case '(':
			// Look for url to the left.
			p := bytes.TrimRight(s[:i], "\t\n\f\r ")
			if endsWithCSSKeyword(p, "url") {
				j := len(s) - len(bytes.TrimLeft(s[i+1:], "\t\n\f\r "))
				switch {
				case j != len(s) && s[j] == '"':
					c.state, j = stateCSSDqURL, j+1
				case j != len(s) && s[j] == '\'':
					c.state, j = stateCSSSqURL, j+1
				default:
					c.state = stateCSSURL
				}
				return c, j
			}
		case '/':
			if i+1 < len(s) {
				switch s[i+1] {
				case '/':
					c.state = stateCSSLineCmt
					return c, i + 2
				case '*':
					c.state = stateCSSBlockCmt
					return c, i + 2
				}
			}
		case '"':
			c.state = stateCSSDqStr
			return c, i + 1
		case '\'':
			c.state = stateCSSSqStr
			return c, i + 1
		}
		k = i + 1
	}
}

// tCSSStr is the context transition function for the CSS string and URL states.
func tCSSStr(c context, s []byte) (context, int) {
	var endAndEsc string
	switch c.state {
	case stateCSSDqStr, stateCSSDqURL:
		endAndEsc = `\"`
	case stateCSSSqStr, stateCSSSqURL:
		endAndEsc = `\'`
	case stateCSSURL:
		// Unquoted URLs end with a newline or close parenthesis.
		// The below includes the wc (whitespace character) and nl.
		endAndEsc = "\\\t\n\f\r )"
	default:
		panic(c.state.String())
	}

	k := 0
	for {
		i := k + bytes.IndexAny(s[k:], endAndEsc)
		if i < k {
			c, nread := tURL(c, decodeCSS(s[k:]))
			return c, k + nread
		}
		if s[i] == '\\' {
			i++
			if i == len(s) {
				e := fmt.Errorf("unfinished escape sequence in CSS string: %q", s)
				return context{
					state: stateError,
					err:   &e,
				}, len(s)
			}
		} else {
			c.state = stateCSS
			return c, i + 1
		}
		c, _ = tURL(c, decodeCSS(s[:i+1]))
		k = i + 1
	}
}

// tError is the context transition function for the error state.
func tError(c context, s []byte) (context, int) {
	return c, len(s)
}

// eatAttrName returns the largest j such that s[i:j] is an attribute name.
// It returns an error if s[i:] does not look like it begins with an
// attribute name, such as encountering a quote mark without a preceding
// equals sign.
func eatAttrName(s []byte, i int) (int, *error) {
	for j := i; j < len(s); j++ {
		switch s[j] {
		case ' ', '\t', '\n', '\f', '\r', '=', '>':
			return j, nil
		case '\'', '"', '<':
			e := fmt.Errorf("%q in attribute name: %.32q", s[j:j+1], s)
			// These result in a parse warning in HTML5 and are
			// indicative of serious problems if seen in an attr
			// name in a template.
			return -1, &e
		default:
			// No-op.
		}
	}
	return len(s), nil
}

var elementNameMap = map[string]element{
	"script":   elementScript,
	"style":    elementStyle,
	"textarea": elementTextarea,
	"title":    elementTitle,
}

// asciiAlpha reports whether c is an ASCII letter.
func asciiAlpha(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
}

// asciiAlphaNum reports whether c is an ASCII letter or digit.
func asciiAlphaNum(c byte) bool {
	return asciiAlpha(c) || '0' <= c && c <= '9'
}

// eatTagName returns the largest j such that s[i:j] is a tag name and the tag type.
func eatTagName(s []byte, i int) (int, element) {
	if i == len(s) || !asciiAlpha(s[i]) {
		return i, elementNone
	}
	j := i + 1
	for j < len(s) {
		x := s[j]
		if asciiAlphaNum(x) {
			j++
			continue
		}
		// Allow "x-y" or "x:y" but not "x-", "-y", or "x--y".
		if (x == ':' || x == '-') && j+1 < len(s) && asciiAlphaNum(s[j+1]) {
			j += 2
			continue
		}
		break
	}
	return j, elementNameMap[strings.ToLower(string(s[i:j]))]
}

// eatWhiteSpace returns the largest j such that s[i:j] is white space.
func eatWhiteSpace(s []byte, i int) int {
	for j := i; j < len(s); j++ {
		switch s[j] {
		case ' ', '\t', '\n', '\f', '\r':
			// No-op.
		default:
			return j
		}
	}
	return len(s)
}

// Strings of content from a trusted source.
type (
	// CSS encapsulates known safe content that matches any of:
	//   1. The CSS3 stylesheet production, such as `p { color: purple }`.
	//   2. The CSS3 rule production, such as `a[href=~"https:"].foo#bar`.
	//   3. CSS3 declaration productions, such as `color: red; margin: 2px`.
	//   4. The CSS3 value production, such as `rgba(0, 0, 255, 127)`.
	// See http://www.w3.org/TR/css3-syntax/#parsing and
	// https://web.archive.org/web/20090211114933/http://w3.org/TR/css3-syntax#style
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	CSS string

	// HTML encapsulates a known safe HTML document fragment.
	// It should not be used for HTML from a third-party, or HTML with
	// unclosed tags or comments. The outputs of a sound HTML sanitizer
	// and a template escaped by this package are fine for use with HTML.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	HTML string

	// HTMLAttr encapsulates an HTML attribute from a trusted source,
	// for example, ` dir="ltr"`.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	HTMLAttr string

	// JS encapsulates a known safe EcmaScript5 Expression, for example,
	// `(x + y * z())`.
	// Template authors are responsible for ensuring that typed expressions
	// do not break the intended precedence and that there is no
	// statement/expression ambiguity as when passing an expression like
	// "{ foo: bar() }\n['foo']()", which is both a valid Expression and a
	// valid Program with a very different meaning.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	//
	// Using JS to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JS string

	// JSStr encapsulates a sequence of characters meant to be embedded
	// between quotes in a JavaScript expression.
	// The string must match a series of StringCharacters:
	//   StringCharacter :: SourceCharacter but not `\` or LineTerminator
	//                    | EscapeSequence
	// Note that LineContinuations are not allowed.
	// JSStr("foo\\nbar") is fine, but JSStr("foo\\\nbar") is not.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	JSStr string

	// URL encapsulates a known safe URL or URL substring (see RFC 3986).
	// A URL like `javascript:checkThatFormNotEditedBeforeLeavingPage()`
	// from a trusted source should go in the page, but by default dynamic
	// `javascript:` URLs are filtered out since they are a frequently
	// exploited injection vector.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	URL string
)

type contentType uint8

const (
	contentTypePlain contentType = iota
	contentTypeCSS
	contentTypeHTML
	contentTypeHTMLAttr
	contentTypeJS
	contentTypeJSStr
	contentTypeURL
	// contentTypeUnsafe is used in attr.go for values that affect how
	// embedded content and network messages are formed, vetted,
	// or interpreted; or which credentials network messages carry.
	contentTypeUnsafe
)

// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

var (
	errorType       = reflect.TypeOf((*error)(nil)).Elem()
	fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
)

// indirectToStringerOrError returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of fmt.Stringer
// or error,
func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

// stringify converts its arguments to a string and the type of the content.
// All pointers are dereferenced, as in the text/template package.
func stringify(args ...interface{}) (string, contentType) {
	if len(args) == 1 {
		switch s := indirect(args[0]).(type) {
		case string:
			return s, contentTypePlain
		case CSS:
			return string(s), contentTypeCSS
		case HTML:
			return string(s), contentTypeHTML
		case HTMLAttr:
			return string(s), contentTypeHTMLAttr
		case JS:
			return string(s), contentTypeJS
		case JSStr:
			return string(s), contentTypeJSStr
		case URL:
			return string(s), contentTypeURL
		}
	}
	for i, arg := range args {
		args[i] = indirectToStringerOrError(arg)
	}
	return fmt.Sprint(args...), contentTypePlain
}

// attrTypeMap[n] describes the value of the given attribute.
// If an attribute affects (or can mask) the encoding or interpretation of
// other content, or affects the contents, idempotency, or credentials of a
// network message, then the value in this map is contentTypeUnsafe.
// This map is derived from HTML5, specifically
// http://www.w3.org/TR/html5/Overview.html#attributes-1
// as well as "%URI"-typed attributes from
// http://www.w3.org/TR/html4/index/attributes.html
var attrTypeMap = map[string]contentType{
	"accept":          contentTypePlain,
	"accept-charset":  contentTypeUnsafe,
	"action":          contentTypeURL,
	"alt":             contentTypePlain,
	"archive":         contentTypeURL,
	"async":           contentTypeUnsafe,
	"autocomplete":    contentTypePlain,
	"autofocus":       contentTypePlain,
	"autoplay":        contentTypePlain,
	"background":      contentTypeURL,
	"border":          contentTypePlain,
	"checked":         contentTypePlain,
	"cite":            contentTypeURL,
	"challenge":       contentTypeUnsafe,
	"charset":         contentTypeUnsafe,
	"class":           contentTypePlain,
	"classid":         contentTypeURL,
	"codebase":        contentTypeURL,
	"cols":            contentTypePlain,
	"colspan":         contentTypePlain,
	"content":         contentTypeUnsafe,
	"contenteditable": contentTypePlain,
	"contextmenu":     contentTypePlain,
	"controls":        contentTypePlain,
	"coords":          contentTypePlain,
	"crossorigin":     contentTypeUnsafe,
	"data":            contentTypeURL,
	"datetime":        contentTypePlain,
	"default":         contentTypePlain,
	"defer":           contentTypeUnsafe,
	"dir":             contentTypePlain,
	"dirname":         contentTypePlain,
	"disabled":        contentTypePlain,
	"draggable":       contentTypePlain,
	"dropzone":        contentTypePlain,
	"enctype":         contentTypeUnsafe,
	"for":             contentTypePlain,
	"form":            contentTypeUnsafe,
	"formaction":      contentTypeURL,
	"formenctype":     contentTypeUnsafe,
	"formmethod":      contentTypeUnsafe,
	"formnovalidate":  contentTypeUnsafe,
	"formtarget":      contentTypePlain,
	"headers":         contentTypePlain,
	"height":          contentTypePlain,
	"hidden":          contentTypePlain,
	"high":            contentTypePlain,
	"href":            contentTypeURL,
	"hreflang":        contentTypePlain,
	"http-equiv":      contentTypeUnsafe,
	"icon":            contentTypeURL,
	"id":              contentTypePlain,
	"ismap":           contentTypePlain,
	"keytype":         contentTypeUnsafe,
	"kind":            contentTypePlain,
	"label":           contentTypePlain,
	"lang":            contentTypePlain,
	"language":        contentTypeUnsafe,
	"list":            contentTypePlain,
	"longdesc":        contentTypeURL,
	"loop":            contentTypePlain,
	"low":             contentTypePlain,
	"manifest":        contentTypeURL,
	"max":             contentTypePlain,
	"maxlength":       contentTypePlain,
	"media":           contentTypePlain,
	"mediagroup":      contentTypePlain,
	"method":          contentTypeUnsafe,
	"min":             contentTypePlain,
	"multiple":        contentTypePlain,
	"name":            contentTypePlain,
	"novalidate":      contentTypeUnsafe,
	// Skip handler names from
	// http://www.w3.org/TR/html5/webappapis.html#event-handlers-on-elements,-document-objects,-and-window-objects
	// since we have special handling in attrType.
	"open":        contentTypePlain,
	"optimum":     contentTypePlain,
	"pattern":     contentTypeUnsafe,
	"placeholder": contentTypePlain,
	"poster":      contentTypeURL,
	"profile":     contentTypeURL,
	"preload":     contentTypePlain,
	"pubdate":     contentTypePlain,
	"radiogroup":  contentTypePlain,
	"readonly":    contentTypePlain,
	"rel":         contentTypeUnsafe,
	"required":    contentTypePlain,
	"reversed":    contentTypePlain,
	"rows":        contentTypePlain,
	"rowspan":     contentTypePlain,
	"sandbox":     contentTypeUnsafe,
	"spellcheck":  contentTypePlain,
	"scope":       contentTypePlain,
	"scoped":      contentTypePlain,
	"seamless":    contentTypePlain,
	"selected":    contentTypePlain,
	"shape":       contentTypePlain,
	"size":        contentTypePlain,
	"sizes":       contentTypePlain,
	"span":        contentTypePlain,
	"src":         contentTypeURL,
	"srcdoc":      contentTypeHTML,
	"srclang":     contentTypePlain,
	"start":       contentTypePlain,
	"step":        contentTypePlain,
	"style":       contentTypeCSS,
	"tabindex":    contentTypePlain,
	"target":      contentTypePlain,
	"title":       contentTypePlain,
	"type":        contentTypeUnsafe,
	"usemap":      contentTypeURL,
	"value":       contentTypeUnsafe,
	"width":       contentTypePlain,
	"wrap":        contentTypePlain,
	"xmlns":       contentTypeURL,
}

// attrType returns a conservative (upper-bound on authority) guess at the
// type of the named attribute.
func attrType(name string) contentType {
	name = strings.ToLower(name)
	if strings.HasPrefix(name, "data-") {
		// Strip data- so that custom attribute heuristics below are
		// widely applied.
		// Treat data-action as URL below.
		name = name[5:]
	} else if colon := strings.IndexRune(name, ':'); colon != -1 {
		if name[:colon] == "xmlns" {
			return contentTypeURL
		}
		// Treat svg:href and xlink:href as href below.
		name = name[colon+1:]
	}
	if t, ok := attrTypeMap[name]; ok {
		return t
	}
	// Treat partial event handler names as script.
	if strings.HasPrefix(name, "on") {
		return contentTypeJS
	}

	// Heuristics to prevent "javascript:..." injection in custom
	// data attributes and custom attributes like g:tweetUrl.
	// http://www.w3.org/TR/html5/dom.html#embedding-custom-non-visible-data-with-the-data-*-attributes
	// "Custom data attributes are intended to store custom data
	//  private to the page or application, for which there are no
	//  more appropriate attributes or elements."
	// Developers seem to store URL content in data URLs that start
	// or end with "URI" or "URL".
	if strings.Contains(name, "src") ||
		strings.Contains(name, "uri") ||
		strings.Contains(name, "url") {
		return contentTypeURL
	}
	return contentTypePlain
}

// nextJSCtx returns the context that determines whether a slash after the
// given run of tokens starts a regular expression instead of a division
// operator: / or /=.
//
// This assumes that the token run does not include any string tokens, comment
// tokens, regular expression literal tokens, or division operators.
//
// This fails on some valid but nonsensical JavaScript programs like
// "x = ++/foo/i" which is quite different than "x++/foo/i", but is not known to
// fail on any known useful programs. It is based on the draft
// JavaScript 2.0 lexical grammar and requires one token of lookbehind:
// http://www.mozilla.org/js/language/js20-2000-07/rationale/syntax.html
func nextJSCtx(s []byte, preceding jsCtx) jsCtx {
	s = bytes.TrimRight(s, "\t\n\f\r \u2028\u2029")
	if len(s) == 0 {
		return preceding
	}

	// All cases below are in the single-byte UTF-8 group.
	switch c, n := s[len(s)-1], len(s); c {
	case '+', '-':
		// ++ and -- are not regexp preceders, but + and - are whether
		// they are used as infix or prefix operators.
		start := n - 1
		// Count the number of adjacent dashes or pluses.
		for start > 0 && s[start-1] == c {
			start--
		}
		if (n-start)&1 == 1 {
			// Reached for trailing minus signs since "---" is the
			// same as "-- -".
			return jsCtxRegexp
		}
		return jsCtxDivOp
	case '.':
		// Handle "42."
		if n != 1 && '0' <= s[n-2] && s[n-2] <= '9' {
			return jsCtxDivOp
		}
		return jsCtxRegexp
		// Suffixes for all punctuators from section 7.7 of the language spec
		// that only end binary operators not handled above.
	case ',', '<', '>', '=', '*', '%', '&', '|', '^', '?':
		return jsCtxRegexp
		// Suffixes for all punctuators from section 7.7 of the language spec
		// that are prefix operators not handled above.
	case '!', '~':
		return jsCtxRegexp
		// Matches all the punctuators from section 7.7 of the language spec
		// that are open brackets not handled above.
	case '(', '[':
		return jsCtxRegexp
		// Matches all the punctuators from section 7.7 of the language spec
		// that precede expression starts.
	case ':', ';', '{':
		return jsCtxRegexp
		// CAVEAT: the close punctuators ('}', ']', ')') precede div ops and
		// are handled in the default except for '}' which can precede a
		// division op as in
		//    ({ valueOf: function () { return 42 } } / 2
		// which is valid, but, in practice, developers don't divide object
		// literals, so our heuristic works well for code like
		//    function () { ... }  /foo/.test(x) && sideEffect();
		// The ')' punctuator can precede a regular expression as in
		//     if (b) /foo/.test(x) && ...
		// but this is much less likely than
		//     (a + b) / c
	case '}':
		return jsCtxRegexp
	default:
		// Look for an IdentifierName and see if it is a keyword that
		// can precede a regular expression.
		j := n
		for j > 0 && isJSIdentPart(rune(s[j-1])) {
			j--
		}
		if regexpPrecederKeywords[string(s[j:])] {
			return jsCtxRegexp
		}
	}
	// Otherwise is a punctuator not listed above, or
	// a string which precedes a div op, or an identifier
	// which precedes a div op.
	return jsCtxDivOp
}

// regexpPrecederKeywords is a set of reserved JS keywords that can precede a
// regular expression in JS source.
var regexpPrecederKeywords = map[string]bool{
	"break":      true,
	"case":       true,
	"continue":   true,
	"delete":     true,
	"do":         true,
	"else":       true,
	"finally":    true,
	"in":         true,
	"instanceof": true,
	"return":     true,
	"throw":      true,
	"try":        true,
	"typeof":     true,
	"void":       true,
}

var jsonMarshalType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()

// indirectToJSONMarshaler returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of json.Marshal.
func indirectToJSONMarshaler(a interface{}) interface{} {
	v := reflect.ValueOf(a)
	for !v.Type().Implements(jsonMarshalType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

// jsValEscaper escapes its inputs to a JS Expression (section 11.14) that has
// neither side-effects nor free variables outside (NaN, Infinity).
func jsValEscaper(args ...interface{}) string {
	var a interface{}
	if len(args) == 1 {
		a = indirectToJSONMarshaler(args[0])
		switch t := a.(type) {
		case JS:
			return string(t)
		case JSStr:
			// TODO: normalize quotes.
			return `"` + string(t) + `"`
		case json.Marshaler:
			// Do not treat as a Stringer.
		case fmt.Stringer:
			a = t.String()
		}
	} else {
		for i, arg := range args {
			args[i] = indirectToJSONMarshaler(arg)
		}
		a = fmt.Sprint(args...)
	}
	// TODO: detect cycles before calling Marshal which loops infinitely on
	// cyclic data. This may be an unacceptable DoS risk.

	b, err := json.Marshal(a)
	if err != nil {
		// Put a space before comment so that if it is flush against
		// a division operator it is not turned into a line comment:
		//     x/{{y}}
		// turning into
		//     x//* error marshaling y:
		//          second line of error message */null
		return fmt.Sprintf(" /* %s */null ", strings.Replace(err.Error(), "*/", "* /", -1))
	}

	// TODO: maybe post-process output to prevent it from containing
	// "<!--", "-->", "<![CDATA[", "]]>", or "</script"
	// in case custom marshalers produce output containing those.

	// TODO: Maybe abbreviate \u00ab to \xab to produce more compact output.
	if len(b) == 0 {
		// In, `x=y/{{.}}*z` a json.Marshaler that produces "" should
		// not cause the output `x=y/*z`.
		return " null "
	}
	first, _ := utf8.DecodeRune(b)
	last, _ := utf8.DecodeLastRune(b)
	var buf bytes.Buffer
	// Prevent IdentifierNames and NumericLiterals from running into
	// keywords: in, instanceof, typeof, void
	pad := isJSIdentPart(first) || isJSIdentPart(last)
	if pad {
		buf.WriteByte(' ')
	}
	written := 0
	// Make sure that json.Marshal escapes codepoints U+2028 & U+2029
	// so it falls within the subset of JSON which is valid JS.
	for i := 0; i < len(b); {
		rune, n := utf8.DecodeRune(b[i:])
		repl := ""
		if rune == 0x2028 {
			repl = `\u2028`
		} else if rune == 0x2029 {
			repl = `\u2029`
		}
		if repl != "" {
			buf.Write(b[written:i])
			buf.WriteString(repl)
			written = i + n
		}
		i += n
	}
	if buf.Len() != 0 {
		buf.Write(b[written:])
		if pad {
			buf.WriteByte(' ')
		}
		b = buf.Bytes()
	}
	return string(b)
}

// jsStrEscaper produces a string that can be included between quotes in
// JavaScript source, in JavaScript embedded in an HTML5 <script> element,
// or in an HTML5 event handler attribute such as onclick.
func jsStrEscaper(args ...interface{}) string {
	s, t := stringify(args...)
	if t == contentTypeJSStr {
		return replace(s, jsStrNormReplacementTable)
	}
	return replace(s, jsStrReplacementTable)
}

// jsRegexpEscaper behaves like jsStrEscaper but escapes regular expression
// specials so the result is treated literally when included in a regular
// expression literal. /foo{{.X}}bar/ matches the string "foo" followed by
// the literal text of {{.X}} followed by the string "bar".
func jsRegexpEscaper(args ...interface{}) string {
	s, _ := stringify(args...)
	s = replace(s, jsRegexpReplacementTable)
	if s == "" {
		// /{{.X}}/ should not produce a line comment when .X == "".
		return "(?:)"
	}
	return s
}

// replace replaces each rune r of s with replacementTable[r], provided that
// r < len(replacementTable). If replacementTable[r] is the empty string then
// no replacement is made.
// It also replaces runes U+2028 and U+2029 with the raw strings `\u2028` and
// `\u2029`.
func replace(s string, replacementTable []string) string {
	var b bytes.Buffer
	r, w, written := rune(0), 0, 0
	for i := 0; i < len(s); i += w {
		// See comment in htmlEscaper.
		r, w = utf8.DecodeRuneInString(s[i:])
		var repl string
		switch {
		case int(r) < len(replacementTable) && replacementTable[r] != "":
			repl = replacementTable[r]
		case r == '\u2028':
			repl = `\u2028`
		case r == '\u2029':
			repl = `\u2029`
		default:
			continue
		}
		b.WriteString(s[written:i])
		b.WriteString(repl)
		written = i + w
	}
	if written == 0 {
		return s
	}
	b.WriteString(s[written:])
	return b.String()
}

var jsStrReplacementTable = []string{
	0:    `\0`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\x0b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\x22`,
	'&':  `\x26`,
	'\'': `\x27`,
	'+':  `\x2b`,
	'/':  `\/`,
	'<':  `\x3c`,
	'>':  `\x3e`,
	'\\': `\\`,
}

// jsStrNormReplacementTable is like jsStrReplacementTable but does not
// overencode existing escapes since this table has no entry for `\`.
var jsStrNormReplacementTable = []string{
	0:    `\0`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\x0b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\x22`,
	'&':  `\x26`,
	'\'': `\x27`,
	'+':  `\x2b`,
	'/':  `\/`,
	'<':  `\x3c`,
	'>':  `\x3e`,
}

var jsRegexpReplacementTable = []string{
	0:    `\0`,
	'\t': `\t`,
	'\n': `\n`,
	'\v': `\x0b`, // "\v" == "v" on IE 6.
	'\f': `\f`,
	'\r': `\r`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\x22`,
	'$':  `\$`,
	'&':  `\x26`,
	'\'': `\x27`,
	'(':  `\(`,
	')':  `\)`,
	'*':  `\*`,
	'+':  `\x2b`,
	'-':  `\-`,
	'.':  `\.`,
	'/':  `\/`,
	'<':  `\x3c`,
	'>':  `\x3e`,
	'?':  `\?`,
	'[':  `\[`,
	'\\': `\\`,
	']':  `\]`,
	'^':  `\^`,
	'{':  `\{`,
	'|':  `\|`,
	'}':  `\}`,
}

// isJSIdentPart reports whether the given rune is a JS identifier part.
// It does not handle all the non-Latin letters, joiners, and combining marks,
// but it does handle every codepoint that can occur in a numeric literal or
// a keyword.
func isJSIdentPart(r rune) bool {
	switch {
	case r == '$':
		return true
	case '0' <= r && r <= '9':
		return true
	case 'A' <= r && r <= 'Z':
		return true
	case r == '_':
		return true
	case 'a' <= r && r <= 'z':
		return true
	}
	return false
}

// isJSType returns true if the given MIME type should be considered JavaScript.
//
// It is used to determine whether a script tag with a type attribute is a javascript container.
func isJSType(mimeType string) bool {
	// per
	//   https://www.w3.org/TR/html5/scripting-1.html#attr-script-type
	//   https://tools.ietf.org/html/rfc7231#section-3.1.1
	//   https://tools.ietf.org/html/rfc4329#section-3
	//   https://www.ietf.org/rfc/rfc4627.txt

	// discard parameters
	if i := strings.Index(mimeType, ";"); i >= 0 {
		mimeType = mimeType[:i]
	}
	mimeType = strings.TrimSpace(mimeType)
	switch mimeType {
	case
		"application/ecmascript",
		"application/javascript",
		"application/json",
		"application/x-ecmascript",
		"application/x-javascript",
		"text/ecmascript",
		"text/javascript",
		"text/javascript1.0",
		"text/javascript1.1",
		"text/javascript1.2",
		"text/javascript1.3",
		"text/javascript1.4",
		"text/javascript1.5",
		"text/jscript",
		"text/livescript",
		"text/x-ecmascript",
		"text/x-javascript":
		return true
	default:
		return false
	}
}

// endsWithCSSKeyword reports whether b ends with an ident that
// case-insensitively matches the lower-case kw.
func endsWithCSSKeyword(b []byte, kw string) bool {
	i := len(b) - len(kw)
	if i < 0 {
		// Too short.
		return false
	}
	if i != 0 {
		r, _ := utf8.DecodeLastRune(b[:i])
		if isCSSNmchar(r) {
			// Too long.
			return false
		}
	}
	// Many CSS keywords, such as "!important" can have characters encoded,
	// but the URI production does not allow that according to
	// http://www.w3.org/TR/css3-syntax/#TOK-URI
	// This does not attempt to recognize encoded keywords. For example,
	// given "\75\72\6c" and "url" this return false.
	return string(bytes.ToLower(b[i:])) == kw
}

// isCSSNmchar reports whether rune is allowed anywhere in a CSS identifier.
func isCSSNmchar(r rune) bool {
	// Based on the CSS3 nmchar production but ignores multi-rune escape
	// sequences.
	// http://www.w3.org/TR/css3-syntax/#SUBTOK-nmchar
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		'0' <= r && r <= '9' ||
		r == '-' ||
		r == '_' ||
		// Non-ASCII cases below.
		0x80 <= r && r <= 0xd7ff ||
		0xe000 <= r && r <= 0xfffd ||
		0x10000 <= r && r <= 0x10ffff
}

// decodeCSS decodes CSS3 escapes given a sequence of stringchars.
// If there is no change, it returns the input, otherwise it returns a slice
// backed by a new array.
// http://www.w3.org/TR/css3-syntax/#SUBTOK-stringchar defines stringchar.
func decodeCSS(s []byte) []byte {
	i := bytes.IndexByte(s, '\\')
	if i == -1 {
		return s
	}
	// The UTF-8 sequence for a codepoint is never longer than 1 + the
	// number hex digits need to represent that codepoint, so len(s) is an
	// upper bound on the output length.
	b := make([]byte, 0, len(s))
	for len(s) != 0 {
		i := bytes.IndexByte(s, '\\')
		if i == -1 {
			i = len(s)
		}
		b, s = append(b, s[:i]...), s[i:]
		if len(s) < 2 {
			break
		}
		// http://www.w3.org/TR/css3-syntax/#SUBTOK-escape
		// escape ::= unicode | '\' [#x20-#x7E#x80-#xD7FF#xE000-#xFFFD#x10000-#x10FFFF]
		if isHex(s[1]) {
			// http://www.w3.org/TR/css3-syntax/#SUBTOK-unicode
			//   unicode ::= '\' [0-9a-fA-F]{1,6} wc?
			j := 2
			for j < len(s) && j < 7 && isHex(s[j]) {
				j++
			}
			r := hexDecode(s[1:j])
			if r > unicode.MaxRune {
				r, j = r/16, j-1
			}
			n := utf8.EncodeRune(b[len(b):cap(b)], r)
			// The optional space at the end allows a hex
			// sequence to be followed by a literal hex.
			// string(decodeCSS([]byte(`\A B`))) == "\nB"
			b, s = b[:len(b)+n], skipCSSSpace(s[j:])
		} else {
			// `\\` decodes to `\` and `\"` to `"`.
			_, n := utf8.DecodeRune(s[1:])
			b, s = append(b, s[1:1+n]...), s[1+n:]
		}
	}
	return b
}

// isHex reports whether the given character is a hex digit.
func isHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

// hexDecode decodes a short hex digit sequence: "10" -> 16.
func hexDecode(s []byte) rune {
	n := '\x00'
	for _, c := range s {
		n <<= 4
		switch {
		case '0' <= c && c <= '9':
			n |= rune(c - '0')
		case 'a' <= c && c <= 'f':
			n |= rune(c-'a') + 10
		case 'A' <= c && c <= 'F':
			n |= rune(c-'A') + 10
		default:
			panic(fmt.Sprintf("Bad hex digit in %q", s))
		}
	}
	return n
}

// skipCSSSpace returns a suffix of c, skipping over a single space.
func skipCSSSpace(c []byte) []byte {
	if len(c) == 0 {
		return c
	}
	// wc ::= #x9 | #xA | #xC | #xD | #x20
	switch c[0] {
	case '\t', '\n', '\f', ' ':
		return c[1:]
	case '\r':
		// This differs from CSS3's wc production because it contains a
		// probable spec error whereby wc contains all the single byte
		// sequences in nl (newline) but not CRLF.
		if len(c) >= 2 && c[1] == '\n' {
			return c[2:]
		}
		return c[1:]
	}
	return c
}

// isCSSSpace reports whether b is a CSS space char as defined in wc.
func isCSSSpace(b byte) bool {
	switch b {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	}
	return false
}

// cssEscaper escapes HTML and CSS special characters using \<hex>+ escapes.
func cssEscaper(args ...interface{}) string {
	s, _ := stringify(args...)
	var b bytes.Buffer
	r, w, written := rune(0), 0, 0
	for i := 0; i < len(s); i += w {
		// See comment in htmlEscaper.
		r, w = utf8.DecodeRuneInString(s[i:])
		var repl string
		switch {
		case int(r) < len(cssReplacementTable) && cssReplacementTable[r] != "":
			repl = cssReplacementTable[r]
		default:
			continue
		}
		b.WriteString(s[written:i])
		b.WriteString(repl)
		written = i + w
		if repl != `\\` && (written == len(s) || isHex(s[written]) || isCSSSpace(s[written])) {
			b.WriteByte(' ')
		}
	}
	if written == 0 {
		return s
	}
	b.WriteString(s[written:])
	return b.String()
}

var cssReplacementTable = []string{
	0:    `\0`,
	'\t': `\9`,
	'\n': `\a`,
	'\f': `\c`,
	'\r': `\d`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\22`,
	'&':  `\26`,
	'\'': `\27`,
	'(':  `\28`,
	')':  `\29`,
	'+':  `\2b`,
	'/':  `\2f`,
	':':  `\3a`,
	';':  `\3b`,
	'<':  `\3c`,
	'>':  `\3e`,
	'\\': `\\`,
	'{':  `\7b`,
	'}':  `\7d`,
}

var expressionBytes = []byte("expression")
var mozBindingBytes = []byte("mozbinding")

// filterFailsafe is an innocuous word that is emitted in place of unsafe values
// by sanitizer functions. It is not a keyword in any programming language,
// contains no special characters, is not empty, and when it appears in output
// it is distinct enough that a developer can find the source of the problem
// via a search engine.
const filterFailsafe = "ZgotmplZ"

// cssValueFilter allows innocuous CSS values in the output including CSS
// quantities (10px or 25%), ID or class literals (#foo, .bar), keyword values
// (inherit, blue), and colors (#888).
// It filters out unsafe values, such as those that affect token boundaries,
// and anything that might execute scripts.
func cssValueFilter(args ...interface{}) string {
	s, t := stringify(args...)
	if t == contentTypeCSS {
		return s
	}
	b, id := decodeCSS([]byte(s)), make([]byte, 0, 64)

	// CSS3 error handling is specified as honoring string boundaries per
	// http://www.w3.org/TR/css3-syntax/#error-handling :
	//     Malformed declarations. User agents must handle unexpected
	//     tokens encountered while parsing a declaration by reading until
	//     the end of the declaration, while observing the rules for
	//     matching pairs of (), [], {}, "", and '', and correctly handling
	//     escapes. For example, a malformed declaration may be missing a
	//     property, colon (:) or value.
	// So we need to make sure that values do not have mismatched bracket
	// or quote characters to prevent the browser from restarting parsing
	// inside a string that might embed JavaScript source.
	for i, c := range b {
		switch c {
		case 0, '"', '\'', '(', ')', '/', ';', '@', '[', '\\', ']', '`', '{', '}':
			return filterFailsafe
		case '-':
			// Disallow <!-- or -->.
			// -- should not appear in valid identifiers.
			if i != 0 && b[i-1] == '-' {
				return filterFailsafe
			}
		default:
			if c < utf8.RuneSelf && isCSSNmchar(rune(c)) {
				id = append(id, c)
			}
		}
	}
	id = bytes.ToLower(id)
	if bytes.Contains(id, expressionBytes) || bytes.Contains(id, mozBindingBytes) {
		return filterFailsafe
	}
	return string(b)
}
