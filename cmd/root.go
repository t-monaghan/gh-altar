package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// BrokerAddress is the address of the altar broker being targeted by gh-altar.
var BrokerAddress string

// GithubToken is an optional token for providing authorisation to queries.
var GithubToken string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gh-altar",
	Short: "a github cli extension for usage with altar brokers",
	Long: `gh-altar is a tool for pushing github related information to your awtrix device via an altar broker.
e.g. "gh altar ci" will watch the actions for the PR on the branch you are currently cd'd into.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gh-altar.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVarP(&BrokerAddress, "broker-address",
		"a", "http://127.0.0.1:25827", "IP Address of your altar broker admin server")
	rootCmd.PersistentFlags().StringVar(&GithubToken, "token", "", "github auth token")
}
