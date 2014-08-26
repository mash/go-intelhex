package intelhex

import (
	"bytes"
	"testing"
)

type parseTestCase struct {
	in          string
	out         []Record
	recordCount int
}

var parseTestCases = []parseTestCase{
	{
		":01000000CB34\n",
		[]Record{
			{1, 0, RecordTypeData, "CB"},
		},
		1,
	},
	{
		":200000000C94AE040C94D6040C94D6040C94D6040C94D6040C94D6040C94D6040C94D60438\n" +
			":200020000C94D6040C94D6040C9474320C94FB320C94D6040C94D6040C94D6040C94D604D1\n",
		[]Record{
			{64, 0, RecordTypeData, "0C94AE040C94D6040C94D6040C94D6040C94D6040C94D6040C94D6040C94D6040C94D6040C94D6040C9474320C94FB320C94D6040C94D6040C94D6040C94D604"},
		},
		1,
	},
}

func TestParse(t *testing.T) {
	for _, testCase := range parseTestCases {
		// fmt.Printf("testCase: %v\n", testCase)
		_, records := ParseString(testCase.in)
		var recordCount int = 0
		var b bytes.Buffer
		for record := range records {
			for _, expectedRecord := range testCase.out {
				if record != expectedRecord {
					t.Errorf("parse failed\nparsed:%v\nexpected:%v", record, expectedRecord)
				}
			}
			recordCount++
			b.Write(record.Format(32))
		}
		if recordCount != testCase.recordCount {
			t.Errorf("recordCount diff\nresult:%v\nexpected:%v", recordCount, testCase.recordCount)
		}
		if b.String() != testCase.in {
			t.Errorf("Format diff\nresult:%v\nexpected:%v", b.String(), testCase.in)
		}
	}
}
