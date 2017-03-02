package reader

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mash/go-intelhex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := MainCmd.Flags()

	flags.Bool("binary", false, "display binary")
	viper.BindPFlag("binary", flags.Lookup("binary"))
}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read data from hex file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("Missing file")
		}

		hexFile, err := os.Open(args[0])

		if err != nil {
			log.Fatal(err) //log.Fatal run an os.Exit
		}
		defer hexFile.Close()

		bytes, err := ioutil.ReadAll(hexFile)
		if err != nil {
			log.Fatal(err) //log.Fatal run an os.Exit
		}

		_, records := intelhex.ParseString(string(bytes))
		for record := range records {
			if viper.GetBool("binary") {
				src := strings.ToLower(record.Data)
				dst, err := hex.DecodeString(src)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("%s", dst)
			} else {
				fmt.Printf("%s\n", record.Data)
			}

		}

	},
}
