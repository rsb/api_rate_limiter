// Package cmd is responsible initializing commands and the functions
// executed when those commands are called. All commands are using
// cobra and viper packages
package cmd

import (
	"github.com/joho/godotenv"
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/failure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"

	rsbConf "github.com/rsb/conf"
)

var (
	cfgFile string
	build   = "develop"
)

// rootCmd is the base cli command
var rootCmd = &cobra.Command{
	Use:   "limiter",
	Short: "cli tool to help develop, deploy and test the rate limiter example",
	Long: `limiter cli tool is used for the following:
- Aid in development of the limiter api service
- API server management
- Admin tasks
`,
	Version: build,
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("[init] no .env file used.")
	}
}

func InitConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		root := app.RootDir()
		viper.AddConfigPath(root)
		viper.SetConfigType("toml")
		viper.SetConfigName(app.ServiceName)
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	fileUsed := viper.ConfigFileUsed()
	if err != nil {
		fileUsed = "SKIPPED: " + fileUsed
	}
	log.Printf("cli-init, initConfig config-file: [%s]\n", fileUsed)
}

// Execute is the entry point for cobra cli apps
// b is the build version collected by the main.go via -X build flag
func Execute(b string) {
	build = b
	cobra.CheckErr(rootCmd.Execute())
}

// ProcessConfigCLI will convert configuration defined in struct annotations
// and process them though cli flags, env vars and config files.
// CLI flag    - the highest priority
// ENV Var  	 - next highest
// Config file - lowest
func ProcessConfigCLI(v *viper.Viper, c interface{}) error {
	if err := rsbConf.ProcessCLI(v, c); err != nil {
		return failure.Wrap(err, "rsbConf.ProcessCLI failed")
	}

	return nil
}
