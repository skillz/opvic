package controlplane

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/skillz/opvic/controlplane/providers"
	"github.com/skillz/opvic/controlplane/providers/github"
)

type Config struct {
	BindAddr                string
	Token                   *string
	GithubConfig            *github.Config
	CacheExpiration         time.Duration
	CacheReconcilerInterval time.Duration
}

type ControlPlane struct {
	bindAddr                string
	token                   *string
	cache                   *cache.Cache
	cacheReconcilerInterval time.Duration
	provider                *providers.Provider
	mutex                   sync.RWMutex
	reqCount                *prometheus.CounterVec
}

func (conf *Config) NewControlPlane() (*ControlPlane, error) {
	cache := cache.New(conf.CacheExpiration, cache.NoExpiration)
	ctx := context.Background()
	if conf.Token == nil {
		return nil, fmt.Errorf("missing token")
	}

	pConf := providers.Config{
		Github: conf.GithubConfig,
	}
	provider, err := pConf.Init(ctx, cache)
	if err != nil {
		return nil, err
	}
	return &ControlPlane{
		bindAddr:                conf.BindAddr,
		token:                   conf.Token,
		cache:                   cache,
		cacheReconcilerInterval: conf.CacheReconcilerInterval,
		provider:                provider,
		mutex:                   sync.RWMutex{},
		reqCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "requests_total",
			Help:      "The number of HTTP requests processed",
		}, []string{"method", "path", "status"}),
	}, nil
}

func (cp *ControlPlane) Start() {
	// Register counter metrics
	prometheus.MustRegister(cp.reqCount)
	// Setup Routers
	r := cp.SetupRouter()

	// Start background tasks
	go cp.executeCronJobs()

	// Start HTTP server
	r.Run(cp.bindAddr)
}
