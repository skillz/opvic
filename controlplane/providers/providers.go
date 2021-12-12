package providers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"
	"github.com/skillz/opvic/agent/api/v1alpha1"
	"github.com/skillz/opvic/controlplane/providers/github"
	"github.com/skillz/opvic/controlplane/providers/helm"
)

const (
	Github ProviderType = "github"
	Helm   ProviderType = "helm"
)

type ProviderType string

func (p ProviderType) String() string {
	return string(p)
}

type Config struct {
	Logger logr.Logger
	Github *github.Config
}

type Provider struct {
	log    logr.Logger
	Github *github.Provider
	Helm   *helm.Provider
}

func (c *Config) Init(ctx context.Context, cache *cache.Cache) (*Provider, error) {
	var err error
	p := &Provider{}
	logger := c.Logger.WithName("provider")
	p.Github, err = c.Github.NewProvider(ctx, cache, logger.WithName("github"))
	if err != nil {
		return nil, err
	}
	p.Helm = helm.NewProvider(cache, logger.WithName("helm"))
	p.log = logger
	return p, nil
}

func (p *Provider) GetVersions(conf v1alpha1.RemoteVersion) ([]string, error) {
	switch conf.Provider {
	case Github.String():
		return p.Github.GetVersions(conf)
	case Helm.String():
		return p.Helm.GetVersions(conf)
	default:
		return nil, fmt.Errorf("unknown provider %s", conf.Provider)
	}
}
