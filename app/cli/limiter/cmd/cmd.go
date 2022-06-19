// Package cmd is responsible initializing commands and the functions
// executed when those commands are called. All commands are using
// cobra and viper packages
package cmd

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/failure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"

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

	cobra.OnInitialize(initConfig)
	template := `{{printf "%s: %s - version %s\n" .Name .Short .Version}}`
	rootCmd.SetVersionTemplate(template)

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(strings.ToUpper(app.ServiceName))

	rootCmd.AddCommand(apiCmd)

	// api sub commands
	apiCmd.AddCommand(serveCmd)
}

func initConfig() {
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

// processConfigCLI will convert configuration defined in struct annotations
// and process them though cli flags, env vars and config files.
// CLI flag    - the highest priority
// ENV Var  	 - next highest
// Config file - lowest
func processConfigCLI(v *viper.Viper, c interface{}) error {
	if err := rsbConf.ProcessCLI(v, c); err != nil {
		return failure.Wrap(err, "rsbConf.ProcessCLI failed")
	}

	return nil
}

func bindCLI(c *cobra.Command, v *viper.Viper, config interface{}) {
	if err := rsbConf.BindCLI(c, v, config); err != nil {
		err = failure.Wrap(err, "rsbConf.BindCLI failed")
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func failureExit(log *zap.SugaredLogger, err error, cat, msg string) {
	err = failure.Wrap(err, msg)
	if log == nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: (%s)", err.Error())
		os.Exit(1)
	}

	log.Errorw(cat, "ERROR", err)
	os.Exit(1)
}
