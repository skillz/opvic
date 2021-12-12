/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/skillz/opvic/controlplane"
	"github.com/skillz/opvic/controlplane/providers/github"
	zaplib "go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

var (
	controlPlaneBindAddr         = kingpin.Flag("controlplane.bind-address", "The address the metric endpoint binds to.").Envar("CONTROLPLANE_BIND_ADDRESS").Default(":8080").String()
	controlPlaneAuthToken        = kingpin.Flag("controlplane.auth-token", "Control Plane Shared Auth Token").Envar("CONTROLPLANE_AUTH_TOKEN").Required().String()
	providerGithubToken          = kingpin.Flag("provider.github.token", "Github PAT for the github provider").Envar("PROVIDER_GITHUB_TOKEN").String()
	providerGithubAppID          = kingpin.Flag("provider.github.app-id", "Github App ID for the github provider").Envar("PROVIDER_GITHUB_GITHUB_APP_ID").Int64()
	providerGithubInstallationID = kingpin.Flag("provider.github.app-installation-id", "Github App ID for the github provider").Envar("PROVIDER_GITHUB_APP_INSTALLATION_ID").Int64()
	providerGithubAppPrivateKey  = kingpin.Flag("provider.github.app-private-key", "Github APP Private Key for github provider").Envar("PROVIDER_GITHUB_APP_PRIVATE_KEY").Default("").String()
	cacheExpiration              = kingpin.Flag("cache.expiration", "Cache expiration duration").Envar("CACHE_EXPIRATION").Default("1h").Duration()
	cacheReconcilerInterval      = kingpin.Flag("cache.reconciler-interval", "Cache reconciler interval").Envar("CACHE_RECONCILER_INTERVAL").Default("30s").Duration()
	logLevel                     = kingpin.Flag("log.level", "The verbosity of the logging. Valid values are `debug`, `info`, `warn`, `error`").Envar("LOG_LEVEL").Default("info").String()
	logHttpRequests              = kingpin.Flag("log.http-requests", "Enable HTTP request logging").Envar("LOG_HTTP_REQUESTS").Default("false").Bool()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := zap.New(func(o *zap.Options) {
		gin.SetMode(gin.ReleaseMode)
		switch *logLevel {
		case logLevelDebug:
			gin.SetMode(gin.DebugMode)
			o.Development = true
		case logLevelInfo:
			lvl := zaplib.NewAtomicLevelAt(zaplib.InfoLevel)
			o.Level = &lvl
		case logLevelWarn:
			lvl := zaplib.NewAtomicLevelAt(zaplib.WarnLevel)
			o.Level = &lvl
		case logLevelError:
			lvl := zaplib.NewAtomicLevelAt(zaplib.ErrorLevel)
			o.Level = &lvl
		}
	})

	ghConf := github.Config{
		Token:             *providerGithubToken,
		AppID:             *providerGithubAppID,
		AppInstallationID: *providerGithubInstallationID,
		AppPrivateKey:     *providerGithubAppPrivateKey,
	}

	conf := controlplane.Config{
		BindAddr:                *controlPlaneBindAddr,
		Token:                   controlPlaneAuthToken,
		GithubConfig:            &ghConf,
		CacheExpiration:         *cacheExpiration,
		CacheReconcilerInterval: *cacheReconcilerInterval,
		LogHttpRequests:         *logHttpRequests,
		Logger:                  logger.WithName("opvic-control-plane"),
	}
	cp, err := conf.NewControlPlane()
	if err != nil {
		logger.Error(err, "unable to create the control plane")
		os.Exit(1)
	}
	prometheus.MustRegister(cp)
	cp.Start()
}
