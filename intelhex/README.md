# IntelHex

Decode Intel Hex file, encode data in Intel Hex format. Hex files are use to programm eeprom or microcontroller

## Usage

  intelhex [command]

Available Commands:

  * read        Read data from hex file
  * write       Write data from binary file


### Read
  intelhex read "path" [flags]

Flags:

      --binary   display binary

### Write
  intelhex write "path" [flags]

Flags:

      --immediate   immediate string in encoded
