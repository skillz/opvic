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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/skillz/opvic/agent"
	"github.com/skillz/opvic/agent/api/v1alpha1"
	"gopkg.in/alecthomas/kingpin.v2"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	metricsAddr           = kingpin.Flag("metrics-bind-address", "The address the metric endpoint binds to.").Envar("METRICS_BIND_ADDRESS").Default(":8081").String()
	probeAddr             = kingpin.Flag("health-probe-bind-address", "The address the probe endpoint binds to.").Envar("HEALTH_PROBE_BIND_ADDRESS").Default(":8082").String()
	agentID               = kingpin.Flag("agent.identifier", "Agent unique identifier").Envar("AGENT_IDENTIFIER").Required().String()
	agentInterval         = kingpin.Flag("agent.interval", "Agent reconciliation interval").Envar("AGENT_INTERVAL").Default("60s").Duration()
	agentTags             = kingpin.Flag("agent.tags", "key:value pair to add to the agent tags. (you can pass this flag multiple times").Envar("AGENT_TAGS").PlaceHolder("KEY:VALUE").StringMap()
	controlPlaneUrl       = kingpin.Flag("controlplane.url", "Control Plane URL").Envar("CONTROLPLANE_URL").PlaceHolder("http(s)://CONTROLPLANE-ADDRESS").String()
	controlPlaneAuthToken = kingpin.Flag("controlplane.auth-token", "Control Plane Shared Auth Token").Envar("CONTROLPLANE_AUTH_TOKEN").String()
	logDevelopment        = kingpin.Flag("log.development", "Enable development logging").Envar("LOG_DEVELOPMENT").Default("false").Bool()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	opts := zap.Options{
		Development: *logDevelopment,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     *metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: *probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start agent")
		os.Exit(1)
	}

	conf := &agent.Config{
		Interval:              *agentInterval,
		ID:                    *agentID,
		ControlPlaneUrl:       *controlPlaneUrl,
		ControlPlaneAuthToken: *controlPlaneAuthToken,
		Tags:                  *agentTags,
	}
	if err = (&agent.VersionTrackerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log,
		Scheme: mgr.GetScheme(),
		Config: conf,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VersionTracker")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting agent")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running agent")
		os.Exit(1)
	}
}
