package intelhex

/*

fields -> parser -> records

*/

import (
	"bytes"
	"fmt"
	"strconv"
)

type RecordType int

const (
	RecordTypeData RecordType = iota
	RecordTypeEOF
	RecordTypeExtendedSegmentAddress // TODO: support these (currently unsupported)
	RecordTypeStartSegmentAddress
	RecordTypeExtendedLinearAddress
	RecordTypeStartLinearAddress
	RecordTypeError
)

type Record struct {
	ByteCount int64
	Address   int64
	Type      RecordType
	Data      string
}

// maxByteCount is usually 16 or 32, max 0xFF.
func (r *Record) Format(maxByteCount uint8) []byte {
	var b bytes.Buffer

	switch r.Type {
	case RecordTypeData:
		remainingByteCount := r.ByteCount
		var offset int64 = r.Address
		for remainingByteCount > 0 {
			var nextByteCount int64 = 0
			if remainingByteCount > int64(maxByteCount) {
				nextByteCount = int64(maxByteCount)
			} else {
				nextByteCount = remainingByteCount
			}
			remainingByteCount -= nextByteCount

			rec := Record{
				ByteCount: nextByteCount,
				Address:   offset,
				Type:      RecordTypeData,
				Data:      r.Data[(offset-r.Address)*2 : (offset-r.Address+nextByteCount)*2],
			}

			fmt.Fprintf(&b, ":%02X%04X%02X%s%s\n",
				rec.ByteCount,
				rec.Address,
				rec.Type,
				rec.Data,
				rec.checksum(),
			)

			offset += nextByteCount
		}
	case RecordTypeEOF:
		fmt.Fprintf(&b, ":00000001FF\n")
	}
	return b.Bytes()
}

func (r *Record) checksum() string {
	var sum uint8
	sum = uint8(r.ByteCount)
	sum += uint8(r.Address / 256)
	sum += uint8(r.Address % 256)
	sum += uint8(r.Type)
	for i := 0; i < len(r.Data); i += 2 {
		hexString := r.Data[i : i+2]
		num, err := strconv.ParseUint(hexString, 16, 8)
		if err != nil {
			return ""
		}
		sum += uint8(num)
	}
	return fmt.Sprintf("%02X", (^sum)+1) // two's complement
}

func (r *Record) String() string {
	switch r.Type {
	case RecordTypeData:
		return fmt.Sprintf("{Address:%04X ByteCount:%d Data:%s}", r.Address, r.ByteCount, r.Data)
	case RecordTypeEOF:
		return "EOF"
	default:
		return r.Data
	}
}

func (r *Record) canAppend(nextData *Record) bool {
	if r.Address+r.ByteCount == nextData.Address {
		return true
	}
	return false
}

func (r *Record) append(nextData *Record) {
	r.ByteCount += nextData.ByteCount
	r.Data += nextData.Data
}

type parser struct {
	fields               <-chan field
	records              chan Record
	currentRecord        *Record // nil-able
	cumulativeDataRecord *Record // nil-able
}

func Parse(fields <-chan field) (*parser, <-chan Record) {
	p := &parser{
		fields:  fields,
		records: make(chan Record),
	}
	go p.run() // Concurrently run state machine.
	return p, (<-chan Record)(p.records)
}

func ParseString(in string) (*parser, <-chan Record) {
	_, fields := lex(in)
	return Parse(fields)
}

type parseStateFunc func(*parser) parseStateFunc

// parse fields of the input by executing state functions until
// the state is nil.
func (p *parser) run() {
	for parserState := parseStart; parserState != nil; {
		parserState = parserState(p)
	}
	close(p.records)
}

// accept consumes the next field
// if it's from the valid set.
func (p *parser) accept(valids []fieldType) bool {
	next := p.next()
	for _, valid := range valids {
		if next.typ == valid {
			return true
		}
	}
	return false
}

func (p *parser) acceptValue(validFieldType fieldType) (string, bool) {
	next := p.next()
	if next.typ == validFieldType {
		return next.val, true
	}
	return "", false
}

// next returns the next rune in the input.
func (p *parser) next() field {
	f, ok := <-p.fields
	if !ok {
		return field{fieldEOF, ""}
	}
	// fmt.Println("field: ", f)
	return f
}

func (p *parser) errorf(format string, args ...interface{}) parseStateFunc {
	p.records <- Record{
		0,
		0,
		RecordTypeError,
		fmt.Sprintf(format, args...),
	}
	return parseAfter
}

func parseStart(p *parser) parseStateFunc {
	nextField := p.next()
	if nextField.typ == fieldEOF {
		return parseAfter
	} else if nextField.typ != fieldStartCode {
		return p.errorf("StartCode expected but got something else")
	}

	byteCountString, ok := p.acceptValue(fieldByteCount)
	if !ok {
		return p.errorf("ByteCount expected but got something else")
	}
	byteCount, err := strconv.ParseUint(byteCountString, 16, 8)
	if err != nil {
		return p.errorf("Failed to parse: %s", byteCountString)
	}

	addressString, ok := p.acceptValue(fieldAddress)
	if !ok {
		return p.errorf("Address expected but got something else")
	}
	address, err := strconv.ParseUint(addressString, 16, 16)
	if err != nil {
		return p.errorf("Failed to parse address: %s, %s", addressString, err)
	}

	recordTypString, ok := p.acceptValue(fieldRecordType)
	if !ok {
		return p.errorf("RecordType expected but got something else")
	}
	recordTyp, err := strconv.ParseInt(recordTypString, 16, 8)
	if err != nil {
		return p.errorf("Failed to parse: %s", recordTypString)
	}

	p.currentRecord = &Record{
		int64(byteCount),
		int64(address),
		RecordType(recordTyp),
		"",
	}
	return parseData
}

func parseData(p *parser) parseStateFunc {
	byteCount := p.currentRecord.ByteCount
	if byteCount > 0 {
		data, ok := p.acceptValue(fieldData)
		if !ok {
			return p.errorf("Data expected but got something else")
		}
		p.currentRecord.Data += data
	}
	return parseChecksum
}

func parseChecksum(p *parser) parseStateFunc {
	checksum, ok := p.acceptValue(fieldChecksum)
	if !ok {
		return p.errorf("Checksum expected but got something else")
	}
	calculatedChecksum := p.currentRecord.checksum()
	if calculatedChecksum != checksum {
		return p.errorf("Invalid checksum, got: %s but expected: %s", calculatedChecksum, checksum)
	}
	return parseAfter
}

func parseAfter(p *parser) parseStateFunc {
	if p.currentRecord == nil {
		if p.cumulativeDataRecord != nil {
			p.records <- *p.cumulativeDataRecord
			p.cumulativeDataRecord = nil
		}
		return nil
	} else if p.currentRecord.Type == RecordTypeData {
		if p.cumulativeDataRecord == nil {
			p.cumulativeDataRecord = p.currentRecord
		} else {
			if p.cumulativeDataRecord.canAppend(p.currentRecord) {
				p.cumulativeDataRecord.append(p.currentRecord)
			} else {
				p.records <- *p.cumulativeDataRecord
				p.cumulativeDataRecord = p.currentRecord
			}
		}
	} else {
		if p.cumulativeDataRecord != nil {
			p.records <- *p.cumulativeDataRecord
			p.cumulativeDataRecord = nil
		}

		p.records <- *p.currentRecord
		if p.currentRecord.Type == RecordTypeEOF {
			p.currentRecord = nil
			return nil
		}
	}
	p.currentRecord = nil
	return parseStart
}
