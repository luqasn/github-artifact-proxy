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
	Branch *string
	Event  *string
	Status *string
}

type Target struct {
	Token        *string
	Owner        string
	Repo         string
	Filename     string
	LatestFilter *LatestFilter

	lockChan chan struct{}
	runCache map[string]*Run
}

type Webhook struct {
	Path   string
	Secret string
}

type Http struct {
	Bind     string
	BasePath string
}

type Github struct {
	Tokens   map[string]string
	CacheTTL time.Duration
}

type Config struct {
	DownloadDir string
	Webhook     *Webhook
	Http        *Http
	Github      *Github
	Targets     map[string]*Target
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
