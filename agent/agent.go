package agent

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	v1alpha1 "github.com/skillz/opvic/agent/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	// The interval between individual synchronizations
	Interval time.Duration
	// Agent Identifier
	ID string
	// Url of Control Plane API
	ControlPlaneUrl string
	// Token to authenticate with Control Plane API
	ControlPlaneAuthToken string
	// Tags
	Tags map[string]string
}

// VersionTrackerReconciler reconciles a VersionTracker object
type VersionTrackerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config *Config
}

//+kubebuilder:rbac:groups=vt.skillz.com,resources=versiontrackers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vt.skillz.com,resources=versiontrackers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vt.skillz.com,resources=versiontrackers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *VersionTrackerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("versiontracker", req.NamespacedName)
	start := time.Now()

	log.Info("starting reconciliation", "interval", r.Config.Interval)
	var v v1alpha1.VersionTracker
	var status v1alpha1.VersionTrackerStatus
	var sv SubjectVersion

	if err := r.Get(ctx, req.NamespacedName, &v); err != nil {
		log.Error(err, "unable to fetch VersionTracker")
		reconciliationErrorsTotal.Inc()
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Set defaults
	v.SetDefaults()
	// Validate the VersionTracker
	err := v.Validate()
	if err != nil {
		log.Error(err, "failed to validate VersionTracker")
		reconciliationErrorsTotal.Inc()
		return ctrl.Result{}, err
	}

	// Prepare options fro getting resources defined in the VersionTracker
	var opts []client.ListOption
	if len(v.Spec.Resources.Namespaces) > 0 {
		for _, ns := range v.Spec.Resources.Namespaces {
			opts = append(opts, client.InNamespace(ns))
		}
	}
	selector, err := metav1.LabelSelectorAsSelector(v.Spec.Resources.Selector)
	if err != nil {
		log.Error(err, "failed to convert label selector to selector")
		reconciliationErrorsTotal.Inc()
		return ctrl.Result{}, err
	}
	opts = append(opts, client.MatchingLabelsSelector{Selector: selector})

	// Get the resource object type based on the resource strategy of the VersionTracker
	resources, err := v.GetObjectList()
	if err != nil {
		log.Error(err, "failed to get resource ObjectList")
		reconciliationErrorsTotal.Inc()
		return ctrl.Result{}, err
	}

	// Get all resources
	err = r.List(ctx, resources, opts...)
	if err != nil {
		reconciliationErrorsTotal.Inc()
		log.Error(err, "failed to list pods")
		return ctrl.Result{}, err
	}

	status.ID = &v.Spec.Name
	status.Namespace = &v.ObjectMeta.Namespace
	status.LocalVersion = &v.Spec.LocalVersion
	status.RemoteVersion = &v.Spec.RemoteVersion

	// Get items based on the resource type
	items := GetItems(resources)
	if len(items) == 0 {
		log.Info("no resources found")
		count := 0
		status.TotalResourceCount = &count
	} else {
		// Extract versions from resources
		sv = r.ExtractSubjectVersion(v, items)
		var uniqVersions []*string
		for _, v := range sv.UniqVersions {
			uniqVersions = append(uniqVersions, &v)
		}
		status.TotalResourceCount = &sv.TotalResourceCount
		status.UniqVersions = uniqVersions
		status.Versions = sv.Versions
	}

	// Update the VersionTracker status
	if !reflect.DeepEqual(v.Status, status) {
		updated := v.DeepCopy()
		updated.Status = status
		if err := r.Status().Patch(ctx, updated, client.MergeFrom(&v)); err != nil {
			log.Info("Failed to patch VersionTracker", "error", err)
			return ctrl.Result{
				Requeue: true,
			}, nil
		}
	}

	// Ship the version information to the Control Plane
	if len(sv.Versions) > 0 && r.Config.ControlPlaneUrl != "" {
		err := r.ShipToControlPlane(sv)
		if err != nil {
			log.Error(err, "failed to ship the version to control plane")
			reconciliationErrorsTotal.Inc()
			return ctrl.Result{}, err
		}
	}

	elapsed := time.Since(start)
	lastReconciliationTimestamp.SetToCurrentTime()
	reconciliationDuration.Set(float64(elapsed.Milliseconds()))
	log.Info("done reconciling", "interval", r.Config.Interval)

	return ctrl.Result{
		RequeueAfter: r.Config.Interval,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VersionTrackerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VersionTracker{}).
		Complete(r)
}
