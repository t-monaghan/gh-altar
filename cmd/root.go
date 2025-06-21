package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
func Execute(version string, commit string) {
	rootCmd.Version = fmt.Sprintf("%v@%v", version, commit)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to query user's home directory", "error", err)
		os.Exit(1)
	}

	viper.AddConfigPath(filepath.Join(home, ".config", "altar"))
	viper.AddConfigPath(filepath.Join(home, ".config"))
	viper.AddConfigPath(".")
	viper.SetConfigName("altar")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("could not read config", "error", err)
		os.Exit(1)
	}

	if a := viper.GetString("broker.address"); a == "" {
		slog.Error("could not find configured value for the altar broker address")
	}

	rootCmd.PersistentFlags().StringVar(&GithubToken, "token", "", "github auth token")
}
