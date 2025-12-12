package vmdiskimagectrl

/*
Copyright 2025.

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

import (
	"context"
	crdv1 "pelotech/data-sync-operator/api/v1alpha1"

	vmdiconfig "pelotech/data-sync-operator/internal/vm-disk-image/config"
	vmdi "pelotech/data-sync-operator/internal/vm-disk-image/service"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	crutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// VMDiskImageReconciler reconciles a VMDiskImage object
type VMDiskImageReconciler struct {
	Scheme *runtime.Scheme
	vmdi.VMDiskImageOrchestrator
}

// RBAC for our CRD
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages,verbs=get;list;watch;create;update;patch;delete;delete;deletecollection
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages/status,verbs=get;update;patch;deletecollection
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages/finalizers,verbs=update;deletecollection

// RBAC to preform CRUD operations on pvcs, datavolumes and volumesnapshots
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VMDiskImage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *VMDiskImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	var VMDiskImage crdv1.VMDiskImage

	err := r.GetVMDiskImage(ctx, req.NamespacedName, &VMDiskImage)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Resource has been deleted.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Failed to get VMDiskImage")
		return ctrl.Result{}, err
	}

	resourceMarkedForDeletion := !VMDiskImage.GetDeletionTimestamp().IsZero()
	if resourceMarkedForDeletion {
		return r.VMDiskImageOrchestrator.DeleteResource(ctx, &VMDiskImage)
	}

	resourceHasFinalizer := !crutils.ContainsFinalizer(&VMDiskImage, crdv1.VMDiskImageFinalizer)
	if resourceHasFinalizer {
		err := r.AddControllerFinalizer(ctx, &VMDiskImage)
		if err != nil {
			return r.HandleResourceUpdateError(ctx, &VMDiskImage, err, "Failed to add finalizer to our resource")
		}

	}

	currentPhase := VMDiskImage.Status.Phase
	logger.Info("Reconciling VMDiskImage", "Phase", currentPhase, "Name", VMDiskImage.Name)
	switch currentPhase {
	case "":
		return r.QueueResourceCreation(ctx, &VMDiskImage)
	case crdv1.PhaseQueued:
		return r.AttemptSyncingOfResource(ctx, &VMDiskImage)
	case crdv1.PhaseSyncing:
		return r.TransitonFromSyncing(ctx, &VMDiskImage)
	case crdv1.PhaseRetryableFailure:
		return r.AttemptRetry(ctx, &VMDiskImage)
	case crdv1.PhaseReady, crdv1.PhaseFailed:
		return ctrl.Result{}, nil
	default:
		logger.Error(nil, "Unknown phase detected", "Phase", currentPhase)
		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
// By convention in kubebuilder this is where we are going to setup anything
// the controller requires
func (r *VMDiskImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger := mgr.GetLogger()
	config := vmdiconfig.LoadVMDIControllerConfigFromEnv()

	client := mgr.GetClient()

	resourceGenerator := &vmdi.Generator{}
	vmdiProvisioner := vmdi.K8sVMDIProvisioner{
		Client:                 client,
		ResourceGenerator:      resourceGenerator,
		MaxSyncAttemptDuration: config.MaxSyncAttemptDuration,
		MaxRetryPerAttempt:     config.MaxSyncAttemptRetry,
	}
	orchestrator := vmdi.Orchestrator{
		Client:                client,
		Recorder:              mgr.GetEventRecorderFor(crdv1.VMDiskImageControllerName),
		Provisioner:           vmdiProvisioner,
		MaxSyncAttemptBackoff: config.MaxBackoffDelay,
		ConcurrentSyncLimit:   config.Concurrency,
	}
	reconciler := &VMDiskImageReconciler{
		Scheme:                  mgr.GetScheme(),
		VMDiskImageOrchestrator: orchestrator,
	}

	// Index resources by phase since we have to query these quite a bit
	err := mgr.GetFieldIndexer().
		IndexField(
			context.TODO(),
			&crdv1.VMDiskImage{},
			".status.phase",
			reconciler.IndexVMDiskImageByPhase,
		)
	if err != nil {
		logger.Error(err, "Failed to created indexer")
		return err
	}

	controllerSetupError := ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.VMDiskImage{}).
		Named("vmdiskimage").
		Complete(reconciler)

	return controllerSetupError
}
