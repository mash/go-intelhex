package reader

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

		scanner := bufio.NewScanner(hexFile)
		for scanner.Scan() {
			line := scanner.Text()
			if line != ":00000001FF" && len(line) > 4 {
				_, records := intelhex.ParseString(line)
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
			}


		}
	},
}
