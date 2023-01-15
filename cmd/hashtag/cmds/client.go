package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/pkg/cli"
	"net/http"
	"os"
	"strings"
)

var CompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete one or more hashtags",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		debug, err := cmd.Flags().GetBool("debug")
		cobra.CheckErr(err)

		count, err := cmd.Flags().GetInt("count")
		cobra.CheckErr(err)

		server, err := cmd.Flags().GetString("server")
		cobra.CheckErr(err)

		inputs := []string{}
		for _, arg := range args {
			// if arg start with @, load from file
			if len(arg) >= 2 && arg[0] == '@' {
				// load from file
				s, err := os.ReadFile(arg[1:])
				cobra.CheckErr(err)

				for _, line := range strings.Split(string(s), "\n") {
					inputs = append(inputs, line)
				}
			} else {
				inputs = append(inputs, arg)
			}
		}

		completeRequest := CompleteRequest{
			Inputs: inputs,
			Count:  count,
			Debug:  debug,
		}

		bytes, err := json.Marshal(completeRequest)
		res, err := http.Post(server+"/complete",
			"application/json",
			strings.NewReader(string(bytes)))
		cobra.CheckErr(err)

		completeResponses := []CompleteResponse{}
		err = json.NewDecoder(res.Body).Decode(&completeResponses)
		cobra.CheckErr(err)

		gp, of, err := cli.SetupProcessor(cmd)
		cobra.CheckErr(err)

		// TODO handle debug

		for _, response := range completeResponses {
			for _, result := range response.Hashtags {
				obj := make(map[string]interface{})
				obj["Input"] = response.Input
				obj["Words"] = result.Words
				obj["String"] = result.String
				err = gp.ProcessInputObject(obj)
				cobra.CheckErr(err)
			}
		}

		s, err := of.Output()
		cobra.CheckErr(err)

		fmt.Println(s)
	},
}

func init() {
	CompleteCmd.Flags().String("server", "http://localhost:3333", "Server to use")
	CompleteCmd.Flags().Int("count", 5, "Number of results to return")
	CompleteCmd.Flags().Bool("debug", false, "Enable debug output")

	flagDefaults := cli.NewFlagsDefaults()
	cli.AddFlags(CompleteCmd, flagDefaults)
}
