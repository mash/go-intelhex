package writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/mash/go-intelhex"
)

func init() {
	flags := MainCmd.Flags()

	flags.Bool("immediate", false, "immediate string is encoded")
	viper.BindPFlag("immediate", flags.Lookup("immediate"))
}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write data from binary file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("Missing file/immediate")
		}

        if viper.GetBool("immediate") {
			str := args[0]

			strB := []byte(str)
			encodedStr := hex.EncodeToString(strB)

			record := intelhex.Record{
				ByteCount : int64(len(str)),
				Address   : int64(0),
				Type      : intelhex.RecordTypeData,
				Data      : strings.ToUpper(encodedStr),
			}

			fmt.Printf("%s:00000001FF\n", string(record.Format(16)))

        } else {
            binFile, err := os.Open(args[0])

            if err != nil {
                log.Fatal(err) //log.Fatal run an os.Exit
            }
            defer binFile.Close()

			stats, statsErr := binFile.Stat()
			if statsErr != nil {
				log.Fatal("Cannot read file")
			}

			var size int64 = stats.Size()
			bytes := make([]byte, size)

			bufr := bufio.NewReader(binFile)
			_,err = bufr.Read(bytes)
			if err != nil {
                log.Fatal(err) //log.Fatal run an os.Exit
            }

			encodedStr := hex.EncodeToString(bytes)

			record := intelhex.Record{
				ByteCount : int64(size),
				Address   : int64(0),
				Type      : intelhex.RecordTypeData,
				Data      : strings.ToUpper(encodedStr),
			}

			fmt.Printf("%s:00000001FF\n", string(record.Format(16)))

        }

	},
}
