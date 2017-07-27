package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mash/go-intelhex/intelhex/reader"
	"github.com/mash/go-intelhex/intelhex/writer"
)

/* Main Command to parse
   command line */
var mainCommand = &cobra.Command{
	Use:   "intelhex",
	Short: "Intel hex parser / encoder",
	Long:  "Parse or encode data to intel HEX",
	Run: func(cmd *cobra.Command, args []string) {
		viper.AutomaticEnv()
		// Application statup here
		err := mainApp()
		if err != nil {
			fmt.Println(err)
		}
	},
}

/**
 * The Main application really starts here
 */
func mainApp() (err error) {

	return nil
}

func main() {
	mainCommand.Execute()
}

func init() {
	mainCommand.AddCommand(reader.MainCmd)
	mainCommand.AddCommand(writer.MainCmd)
}
