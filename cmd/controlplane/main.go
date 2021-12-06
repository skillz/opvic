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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/skillz/opvic/controlplane"
	"github.com/skillz/opvic/controlplane/providers/github"
	"gopkg.in/alecthomas/kingpin.v2"
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
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

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
	}
	cp, err := conf.NewControlPlane()
	if err != nil {
		os.Exit(1)
	}
	prometheus.MustRegister(cp)
	cp.Start()
}
