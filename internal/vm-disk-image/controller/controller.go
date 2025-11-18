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
	crdv1 "pelotech/data-sync-operator/api/v1"
	vmdiconfig "pelotech/data-sync-operator/internal/vm-disk-image/config"
	resourcegenservice "pelotech/data-sync-operator/internal/vm-disk-image/service/resource-gen-service"
	resourcemanagerservice "pelotech/data-sync-operator/internal/vm-disk-image/service/resource-manager-service"
	vmdiservice "pelotech/data-sync-operator/internal/vm-disk-image/service/vmdi-service"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// VMDiskImageReconciler reconciles a VMDiskImage object
type VMDiskImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	vmdiservice.VMDiskImageService
}

// RBAC for our CRD
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crd.pelotech.ot,resources=vmdiskimages/finalizers,verbs=update

// RBAC to preform CRUD operations on pvcs, datavolumes and volumesnapshots
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;list;watch;create;update;patch;delete

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

	var dataSync crdv1.VMDiskImage

	err := r.Get(ctx, req.NamespacedName, &dataSync)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("Resource has been deleted.")
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "Failed to get VMDiskImage")
		return ctrl.Result{}, err
	}

	// We don't have our finalizer and haven't been deleted
	if dataSync.GetDeletionTimestamp().IsZero() && !crutils.ContainsFinalizer(&dataSync, crdv1.VMDiskImageFinalizer) {
		crutils.AddFinalizer(&dataSync, crdv1.VMDiskImageFinalizer)

		err := r.Update(ctx, &dataSync)

		if err != nil {
			return r.HandleResourceUpdateError(ctx, &dataSync, err, "Failed to add finalizer to our resource")
		}

	}

	// We have been deleted with our finalizer
	if !dataSync.GetDeletionTimestamp().IsZero() {
		return r.VMDiskImageService.DeleteResource(ctx, &dataSync)
	}

	currentPhase := dataSync.Status.Phase
	logger.Info("Reconciling VMDiskImage", "Phase", currentPhase, "Name", dataSync.Name)

	switch currentPhase {
	case "":
		return r.QueueResourceCreation(ctx, &dataSync)
	case crdv1.VMDiskImagePhaseQueued:
		return r.AttemptSyncingOfResource(ctx, &dataSync)
	case crdv1.VMDiskImagePhaseSyncing:
		return r.TransitonFromSyncing(ctx, &dataSync)
	case crdv1.VMDiskImagePhaseCompleted, crdv1.VMDiskImagePhaseFailed:
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

	// Index resources by phase since we have to query these quite a bit
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &crdv1.VMDiskImage{}, ".status.phase", r.IndexVMDiskImageByPhase)

	if err != nil {
		return err
	}

	config := vmdiconfig.LoadVMDIControllerConfigFromEnv()

	resourceGenerator := &resourcegenservice.Generator{}

	client := mgr.GetClient()

	resourceManager := resourcemanagerservice.Manager{
		K8sClient:         client,
		ResourceGenerator: resourceGenerator,
		MaxSyncDuration:   config.MaxSyncDuration,
		RetryLimit:        config.RetryLimit,
	}

	service := vmdiservice.Service{
		Client:           client,
		Recorder:         mgr.GetEventRecorderFor("vmdi-controller"),
		ResourceManager:  resourceManager,
		ConcurrencyLimit: config.Concurrency,
		RetryLimit:       config.RetryLimit,
		RetryBackoff:     config.RetryBackoffDuration,
	}

	reconciler := &VMDiskImageReconciler{
		Client: client,
		Scheme: mgr.GetScheme(),
		VMDiskImageService: service,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.VMDiskImage{}).
		Named("vmdiskimage").
		Complete(reconciler)
}
