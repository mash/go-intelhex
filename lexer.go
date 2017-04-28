package intelhex

/*

string -> lexer -> fields

*/

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type fieldType int

const (
	fieldError fieldType = iota
	fieldStartCode
	fieldByteCount
	fieldAddress
	fieldRecordType
	fieldData
	fieldChecksum
	fieldEOF // not record type EOF, but the end of hex file
)

// field represents a token returned from the lexer.
type field struct {
	typ fieldType
	val string
}

func (i field) String() string {
	switch i.typ {
	case fieldStartCode:
		return ":"
	case fieldByteCount:
		return "ByteCount  " + i.val
	case fieldAddress:
		return "Address    " + i.val
	case fieldRecordType:
		return "RecordType " + i.val
	case fieldData:
		return "Data       " + i.val
	case fieldChecksum:
		return "Checksum   " + i.val
	}
	return i.val
}

type lexer struct {
	input     string     // the string being scanned.
	start     int        // start position of this field.
	pos       int        // current position in the input.
	width     int        // width of last rune read from input.
	fields    chan field // channel of scanned fields.
	byteCount int64      // Intel Hex byte count in one line
}

type lexStateFunc func(*lexer) lexStateFunc

func lex(input string) (*lexer, <-chan field) {
	l := &lexer{
		input:  input,
		fields: make(chan field),
	}
	go l.run() // Concurrently run state machine.
	return l, (<-chan field)(l.fields)
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexStartCode; state != nil; {
		state = state(l)
	}
	close(l.fields) // No more tokens will be delivered.
}

// emit passes an field back to the client.
func (l *lexer) emit(t fieldType) {
	l.fields <- field{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const eof rune = -1

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptCount consumes a run of runes from the valid set.
func (l *lexer) acceptCount(valid string, count int) bool {
	var i int
	for i = 0; i < count; i++ {
		if strings.IndexRune(valid, l.next()) >= 0 {
			// ok
		} else {
			break
		}
	}
	if i == count {
		return true
	}
	for j := 0; j < i; j++ {
		l.backup()
	}
	return false
}

func (l *lexer) acceptCandidates(candidates []string) bool {
	for _, candidate := range candidates {
		if strings.HasPrefix(l.input[l.start:], candidate) {
			for i := 0; i < len(candidate); i++ {
				l.next()
			}
			return true
		}
	}
	return false
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) lexStateFunc {
	l.fields <- field{
		fieldError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func lexStartCode(l *lexer) lexStateFunc {
	if l.accept(":") {
		l.emit(fieldStartCode)
		return lexByteCount
	}
	return l.errorf("Expected \":\" but got: %q", l.peek())
}

func lexByteCount(l *lexer) lexStateFunc {
	if l.acceptCount("0123456789ABCDEF", 2) {
		byteCountString := l.input[l.start:l.pos]
		byteCount, err := strconv.ParseInt(byteCountString, 16, 16)
		if err != nil {
			return l.errorf("Failed to parse: %s", byteCountString)
		}
		l.byteCount = byteCount
		l.emit(fieldByteCount)
		return lexAddress
	}
	return l.errorf("Expected byteCount but got: %q", l.peek())
}

func lexAddress(l *lexer) lexStateFunc {
	if l.acceptCount("0123456789ABCDEF", 4) {
		l.emit(fieldAddress)
		return lexRecordType
	}
	return l.errorf("Expected address but got: %q", l.peek())
}

func lexRecordType(l *lexer) lexStateFunc {
	if l.acceptCandidates([]string{"00", "01", "02", "03", "04", "05"}) {
		l.emit(fieldRecordType)
		return lexData
	}
	return l.errorf("Expected record type but got: %q", l.peek())
}

func lexData(l *lexer) lexStateFunc {
	if l.byteCount == 0 {
		return lexChecksum
	}
	if l.acceptCount("0123456789ABCDEF", int(l.byteCount*2)) {
		l.emit(fieldData)
		return lexChecksum
	}
	return l.errorf("Expected %d bytes of data but failed", l.byteCount)
}

func lexChecksum(l *lexer) lexStateFunc {
	if l.acceptCount("0123456789ABCDEF", 2) {
		l.emit(fieldChecksum)
		return lexNewline
	}
	return l.errorf("Expected checksum but got: %q", l.peek())
}

func lexNewline(l *lexer) lexStateFunc {
	for l.accept("\r\n") {
		// loop all newlines
	}
	if l.next() == eof {
		l.emit(fieldEOF)
		return nil
	}
	l.backup()
	return lexStartCode
}
