package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v44/github"
)

type Run struct {
	ID        int64
	Artifact  *github.Artifact
	FetchTime time.Time
}

type LatestFilter struct {
	Branch *string `koanf:"branch"`
	Event  *string `koanf:"event"`
	Status *string `koanf:"status"`
}

type Target struct {
	Token        *string       `koanf:"token"`
	Owner        string        `koanf:"owner"`
	Repo         string        `koanf:"repo"`
	Filename     string        `koanf:"filename"`
	LatestFilter *LatestFilter `koanf:"latest_filter"`

	lockChan chan struct{}
	runCache map[string]*Run
}

type Webhook struct {
	Path   string `koanf:"path"`
	Secret string `koanf:"secret"`
}

type Http struct {
	Bind     string `koanf:"bind"`
	BasePath string `koanf:"base_path"`
	// require callers to present this download_token in query args
	DownloadToken string `koanf:"download_token"`
}

type Github struct {
	Tokens   map[string]string `koanf:"tokens"`
	CacheTTL time.Duration     `koanf:"cache_ttl"`
}

type Config struct {
	DownloadDir string             `koanf:"download_dir"`
	Webhook     *Webhook           `koanf:"webhook"`
	Http        Http               `koanf:"http"`
	Github      Github             `koanf:"github"`
	Targets     map[string]*Target `koanf:"targets"`
}

func (t *Target) Lock(ctx context.Context) error {
	select {
	case t.lockChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Target) Unlock() {
	<-t.lockChan
}

func (config *Config) Validate() error {
	for id, target := range config.Targets {
		target.lockChan = make(chan struct{}, 1)
		target.runCache = make(map[string]*Run)

		if target.Token == nil {
			return fmt.Errorf("target '%s' requires an API token", id)
		}

		if _, ok := config.Github.Tokens[*target.Token]; !ok {
			return fmt.Errorf("token with id '%s' not found in tokens list", *target.Token)
		}
	}
	return nil
}
