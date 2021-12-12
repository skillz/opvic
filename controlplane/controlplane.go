package controlplane

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
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
	LogHttpRequests         bool
	Logger                  logr.Logger
}

type ControlPlane struct {
	bindAddr                string
	token                   *string
	cache                   *cache.Cache
	cacheReconcilerInterval time.Duration
	provider                *providers.Provider
	mutex                   sync.RWMutex
	logHttpsRequests        bool
	logger                  logr.Logger
	reqCount                *prometheus.CounterVec
}

func (conf *Config) NewControlPlane() (*ControlPlane, error) {
	logger := conf.Logger
	logger.Info("Initializing the control plane")
	cache := cache.New(conf.CacheExpiration, cache.NoExpiration)
	ctx := context.Background()
	if conf.Token == nil {
		return nil, fmt.Errorf("missing token")
	}

	pConf := providers.Config{
		Logger: logger,
		Github: conf.GithubConfig,
	}
	logger.Info("Initializing the remote providers")
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
		logHttpsRequests:        conf.LogHttpRequests,
		logger:                  conf.Logger,
		reqCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "requests_total",
			Help:      "The number of HTTP requests processed",
		}, []string{"method", "path", "status"}),
	}, nil
}

func (cp *ControlPlane) Start() {
	cp.logger.Info("Starting the control plane")
	prometheus.MustRegister(cp.reqCount)

	cp.logger.Info("Starting the cache reconciler")
	r := cp.SetupRouter()

	cp.logger.Info("Starting the background cache reconciler")
	go cp.executeCronJobs()

	cp.logger.Info("Starting the HTTP server", "bind_addr", cp.bindAddr)
	r.Run(cp.bindAddr)
}
