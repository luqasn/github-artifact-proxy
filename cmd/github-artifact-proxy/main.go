package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var configFile string
var envPrefix string

func main() {
	flag.StringVar(&configFile, "config", "", "the filename of the configuration file")
	flag.StringVar(&envPrefix, "env-prefix", "", "the prefix to use for reading environment variables")
	flag.Parse()

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(".")      // optionally look for config in the working directory
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	viper.SetDefault("Http.Bind", "0.0.0.0:8080")
	viper.SetDefault("Http.BasePath", "/")
	viper.SetDefault("Github.Tokens", map[string]string{})
	viper.SetDefault("Github.CacheTTL", "5m")
	viper.SetDefault("DownloadDir", "/tmp")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error loading config: %w", err))
	}
	log.Infof("Using config file: %q", viper.ConfigFileUsed())

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("error loading config: %w", err))
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Errorf("invalid config: %w", err))
	}

	log.WithField("addr", cfg.Http.Bind).Info("starting http server")

	server := NewServer(&cfg)
	log.Fatal(http.ListenAndServe(cfg.Http.Bind, server))
}
