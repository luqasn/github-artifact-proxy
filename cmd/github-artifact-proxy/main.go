package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	log "github.com/sirupsen/logrus"
)

var (
	configFile            string
	envPrefix             string
	configFromEnvVariable string
	k                     = koanf.New(".")
	parser                = yaml.Parser()
)

func main() {
	flag.StringVar(&configFile, "config", "config.yml", "the filename of the configuration file")
	flag.StringVar(&envPrefix, "env-prefix", "GAP", "the prefix to use for reading environment variables")
	flag.StringVar(&configFromEnvVariable, "config-from-env-var", "", "name of env variable to load config from (instead of file)")
	flag.Parse()

	if err := k.Load(structs.Provider(
		Config{
			Http: Http{
				Bind:     ":8080",
				BasePath: "/",
			},
			Github: Github{
				CacheTTL: 5 * time.Minute,
			},
		}, ""), nil); err != nil {
		log.Fatalf("fatal error loading default values: %v", err)
	}

	if configFromEnvVariable != "" {
		rawConfig := os.Getenv(configFromEnvVariable)
		err := k.Load(rawbytes.Provider([]byte(rawConfig)), parser)
		if err != nil {
			log.Fatalf("fatal error loading config from env: %v", err)
		}
	} else {
		err := k.Load(file.Provider(configFile), parser)
		if err != nil {
			log.Fatalf("fatal error loading config: %v", err)
		}
	}

	if err := k.Load(env.Provider(envPrefix+"_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, envPrefix+"_")), "_", ".", -1)
	}), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("error loading config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %w", err)
	}

	log.WithField("addr", cfg.Http.Bind).Info("starting http server")

	server := NewServer(&cfg)
	log.Fatal(http.ListenAndServe(cfg.Http.Bind, server))
}
