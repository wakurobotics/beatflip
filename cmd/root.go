package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "beatflip",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is '.beatflip.yml')")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cobra.OnInitialize(initConfig, initLogger)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		path, err := os.Getwd()
		cobra.CheckErr(err)

		viper.AddConfigPath(path)
		viper.SetConfigType("yml")
		viper.SetConfigName(".beatflip")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initLogger() {
	if viper.InConfig("log.level") {
		level, err := logrus.ParseLevel(viper.GetString("log.level"))
		if err != nil {
			log.Fatalf("parsing log level failed: %+v", err)
		}
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	switch viper.GetString("log.formatter") {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if viper.InConfig("file") {
		jack := &lumberjack.Logger{}
		err := viper.UnmarshalKey("file", jack)
		cobra.CheckErr(err)
		logrus.SetOutput(io.MultiWriter(os.Stdout, jack))
	}

}
