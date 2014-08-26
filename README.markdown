Intel HEX file format parser/formatter in golang
================================================

## Description

[Intel HEX is a file format that conveys binary information in ASCII text form](http://en.wikipedia.org/wiki/Intel_HEX), used to represent firmware/eeprom images in text.  
This library is useful to post process those images.

## Usage

``` golang
package main

import (
	"fmt"
	"github.com/mash/go-intelhex"
)

func main() {
	_, records := intelhex.ParseString(":01000000CB34")
	for record := range records {
		fmt.Printf("%v\n", record)
		// prints "{1 0 0 CB}"

		fmt.Println(record.String())
		// prints "{Address:0000 ByteCount:1 Data:CB}"

		fmt.Println(string(record.Format(32)))
		// prints ":01000000CB34"
	}
}
```
