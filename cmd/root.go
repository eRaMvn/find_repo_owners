package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "find_repo_owner",
	Short: "Gather the owners based on CODEOWNERS on github",
	Long: `A tool to get the owners for all repos in the organization in github
	Example:
	find_repo_owner -o eRaMvn
	find_repo_owner -o eRaMvn -f owners_to_watch.txt --of result
	`,
	Run: func(cmd *cobra.Command, args []string) {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Fatal("No token has been supplied. Please set access token to environment variable GITHUB_TOKEN!")
		}
		owner, _ = cmd.Flags().GetString("owner")
		outputFile, _ = cmd.Flags().GetString("of")
		if outputFile == "" {
			outputFile = "results_from_repos"
		}
		inputFile, _ = cmd.Flags().GetString("file")
		if inputFile == "" {
			inputFileSupplied = false
		} else {
			inputFileSupplied = true
		}

		ExecuteTask()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("owner", "o", "", "Specify the owner for the organization")
	_ = rootCmd.MarkPersistentFlagRequired("owner")
	rootCmd.PersistentFlags().StringP("file", "f", "", "Specify the list of owners of interest")
	rootCmd.PersistentFlags().String("of", "", "Specify the name of the output file. Default is 'results_from_repos'")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".find_repo_owner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".find_repo_owner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
