package cmd

import (
	"os"

	"github.com/cornelk/goscrape/appcontext"
	"github.com/cornelk/goscrape/scraper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string

	// program parameters
	depth        uint
	includes     []string
	excludes     []string
	output       string
	imageQuality uint
	verbose      bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "goscrape http://website.com",
	Short: "Scrape a website and create an offline browseable version on the disk",
	Run:   startScraper,
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goscrape.yaml)")

	RootCmd.Flags().StringArrayVarP(&includes, "include", "n", nil, "only include URLs with PERL Regular Expressions support")
	RootCmd.Flags().StringArrayVarP(&excludes, "exclude", "x", nil, "exclude URLs with PERL Regular Expressions support")
	RootCmd.Flags().StringVarP(&output, "output", "o", "", "output directory to write files to")
	RootCmd.Flags().UintVarP(&imageQuality, "imagequality", "i", 0, "image quality, 0 to disable reencoding")
	RootCmd.Flags().UintVarP(&depth, "depth", "d", 10, "download depth, 0 for unlimited")
	RootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".goscrape") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	_ = viper.ReadInConfig()
}

func startScraper(cmd *cobra.Command, args []string) {
	log := appcontext.Logger
	if verbose {
		appcontext.LogLevel.SetLevel(zap.DebugLevel)
	}

	if len(args) == 0 {
		cmd.Help()
		return
	}

	for _, url := range args {
		sc, err := scraper.New(url)
		if err != nil {
			log.Fatal("Initializing scraper failed", zap.Error(err))
		}

		err = sc.SetIncludes(includes)
		if err != nil {
			log.Fatal("Configuring includes failed", zap.Error(err))
		}

		err = sc.SetExcludes(excludes)
		if err != nil {
			log.Fatal("Configuring excludes failed", zap.Error(err))
		}

		if imageQuality >= 100 {
			imageQuality = 0
		}

		sc.ImageQuality = imageQuality
		sc.MaxDepth = depth
		sc.OutputDirectory = output

		log.Info("Scraping", zap.Stringer("URL", sc.URL))
		err = sc.Start()
		if err != nil {
			log.Error("Scraping failed", zap.Error(err))
		}
	}
}
