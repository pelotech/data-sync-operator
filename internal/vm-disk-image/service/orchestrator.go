package service

import (
	"context"
	"errors"
	"math"
	crdv1 "pelotech/data-sync-operator/api/v1alpha1"
	"time"

	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	crutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VMDiskImageOrchestrator interface {
	GetVMDiskImage(ctx context.Context, namespace types.NamespacedName, vmdi *crdv1.VMDiskImage) error
	AddControllerFinalizer(ctx context.Context, vmdi *crdv1.VMDiskImage) error
	IndexVMDiskImageByPhase(rawObj client.Object) []string
	ListVMDiskImagesByPhase(ctx context.Context, phase string) (*crdv1.VMDiskImageList, error)
	QueueResourceCreation(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	AttemptSyncingOfResource(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	TransitonFromSyncing(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	AttemptRetry(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	DeleteResource(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	HandleResourceUpdateError(ctx context.Context, ds *crdv1.VMDiskImage, originalErr error, message string) (ctrl.Result, error)
	HandleSyncError(ctx context.Context, vmdi *crdv1.VMDiskImage, originalErr error, message string) (ctrl.Result, error)
}

type Orchestrator struct {
	client.Client
	Recorder              record.EventRecorder
	Provisioner           VMDiskImageProvisioner
	MaxSyncAttemptBackoff time.Duration
	MaxSyncTime           time.Duration
	ConcurrentSyncLimit   int
}

func (o Orchestrator) GetVMDiskImage(ctx context.Context, namespace types.NamespacedName, vmdi *crdv1.VMDiskImage) error {
	return o.Get(ctx, namespace, vmdi)
}

func (o Orchestrator) AddControllerFinalizer(ctx context.Context, vmdi *crdv1.VMDiskImage) error {
	crutils.AddFinalizer(vmdi, crdv1.VMDiskImageFinalizer)

	return o.Update(ctx, vmdi)
}

func (o Orchestrator) ListVMDiskImagesByPhase(ctx context.Context, phase string) (*crdv1.VMDiskImageList, error) {
	list := &crdv1.VMDiskImageList{}

	listOpts := []client.ListOption{
		client.MatchingFields{".status.phase": phase},
	}
	if err := o.List(ctx, list, listOpts...); err != nil {
		return nil, err
	}

	return list, nil
}

func (o Orchestrator) IndexVMDiskImageByPhase(rawObj client.Object) []string {
	vmdi, ok := rawObj.(*crdv1.VMDiskImage)
	if !ok {
		return nil
	}

	if vmdi.Status.Phase == "" {
		return nil
	}

	return []string{vmdi.Status.Phase}
}

func (o Orchestrator) QueueResourceCreation(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error) {
	vmdi.Status.Phase = crdv1.PhaseQueued
	vmdi.Status.Message = "Request is waiting for an available worker."

	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  crdv1.ReasonQueued,
		Message: "The sync has been queued for processing.",
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		return o.HandleResourceUpdateError(ctx, vmdi, err, "Failed to update status to Queued")
	}

	o.Recorder.Eventf(vmdi, "Normal", "Queued", "Resource successfully queued for sync orchestration")

	return ctrl.Result{Requeue: true}, nil
}

func (o Orchestrator) AttemptSyncingOfResource(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	syncingList, err := o.ListVMDiskImagesByPhase(ctx, crdv1.PhaseSyncing)
	if err != nil {
		logger.Error(err, "Failed to list syncing resources")
		return ctrl.Result{}, err
	}
	if len(syncingList.Items) >= o.ConcurrentSyncLimit {
		o.Recorder.Eventf(vmdi, "Normal", "WaitingToSync", "No more than %d VMDiskImages can be syncing at once. Waiting...", o.ConcurrentSyncLimit)
		return ctrl.Result{RequeueAfter: o.MaxSyncAttemptBackoff}, nil
	}

	err = o.Provisioner.CreateResources(ctx, vmdi)
	if err != nil {
		o.Recorder.Eventf(vmdi, "Warning", "ResourceCreationFailed", "Failed to create resources: "+err.Error())
		return o.HandleResourceCreationError(ctx, vmdi, err)
	}

	vmdi.Status.Phase = crdv1.PhaseSyncing
	vmdi.Status.Message = "Syncing VM data for the workspace."
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  crdv1.ReasonSyncing,
		Message: "The sync is currently in progress.",
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		return o.HandleResourceUpdateError(ctx, vmdi, err, "Failed to update status to Syncing")
	}

	o.Recorder.Eventf(vmdi, "Normal", "SyncStarted", "Resource sync has started")

	return ctrl.Result{}, nil
}

func (o Orchestrator) TransitonFromSyncing(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// Check if there is an error occurring in the sync
	syncError := o.Provisioner.ResourcesHaveErrors(ctx, vmdi)
	if syncError != nil {
		logger.Error(syncError, "A sync error has occurred.")
		return o.HandleSyncError(ctx, vmdi, syncError, "A error has occurred while syncing")
	}

	// Check if the sync is done is not done
	isDone, err := o.Provisioner.ResourcesAreReady(ctx, vmdi)
	if err != nil {
		logger.Error(err, "Unable to verify if resource is ready or not.")
	}
	if !isDone {
		logger.Info("Sync is not complete. Requeuing.")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	vmdi.Status.Phase = crdv1.PhaseReady
	vmdi.Status.Message = "The data sync completed successfully."
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  crdv1.ReasonSynced,
		Message: "The sync finished successfully.",
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		return o.HandleResourceUpdateError(ctx, vmdi, err, "Failed to update status to Completed")
	}
	o.Recorder.Eventf(vmdi, "Normal", "SyncCompleted", "Resource sync completed successfully")

	return ctrl.Result{}, nil
}

func (o Orchestrator) AttemptRetry(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error) {
	syncDeadline := metav1.NewTime(vmdi.CreationTimestamp.Add(o.MaxSyncTime))
	exceededSyncDeadline := syncDeadline.Before(ptr.To(metav1.Now()))

	// Fail forever if we're past the deadline
	if exceededSyncDeadline {
		vmdi.Status.Phase = crdv1.PhaseFailed
		vmdi.Status.Message = "Exceeded overall sync retry window. Failed Permanently"
	} else {
		vmdi.Status.Phase = ""
		vmdi.Status.Message = ""
	}

	if err := o.Status().Update(ctx, vmdi); err != nil {
		return o.HandleResourceUpdateError(ctx, vmdi, err, "failed to reset VMDI on retry")
	}

	// Exponential retry
	var backoffInterval time.Duration
	nextBackoffMinutes := int(math.Floor(math.Pow(3, float64(vmdi.Status.FailureCount))))
	nextBackoffDuration := time.Duration(nextBackoffMinutes) * time.Minute
	backoffInterval = min(nextBackoffDuration, o.MaxSyncAttemptBackoff)

	return ctrl.Result{RequeueAfter: backoffInterval}, nil
}

func (o Orchestrator) DeleteResource(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	err := o.Provisioner.TearDownAllResources(ctx, vmdi)
	if err != nil {
		logger.Error(err, "failed to cleanup child resources of VMDiskImage.")
	}

	return ctrl.Result{}, nil
}

func (o Orchestrator) HandleResourceUpdateError(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
	originalErr error,
	message string,
) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Error(originalErr, message)

	// Mark the resource as Failed
	vmdi.Status.Phase = crdv1.PhaseRetryableFailure
	vmdi.Status.Message = "An error occurred during reconciliation: " + originalErr.Error()
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  crdv1.ReasonResouceUpdateFailed,
		Message: originalErr.Error(),
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed after an initial update error")
	}

	return ctrl.Result{}, originalErr
}

func (o Orchestrator) HandleResourceCreationError(ctx context.Context, vmdi *crdv1.VMDiskImage, originalErr error) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Handling a resource creation failure")
	logger.Error(originalErr, "Failed to create a resource.")

	o.Recorder.Eventf(vmdi, "Warning", "ResourceCreationFailed", "Failed to create resources.")
	vmdi.Status.Phase = crdv1.PhaseRetryableFailure
	vmdi.Status.Message = "Failed while creating resources: " + originalErr.Error()
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  crdv1.ReasonResourceCreationFailed,
		Message: originalErr.Error(),
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed resource creation failure")
	}

	err := o.Provisioner.TearDownAllResources(ctx, vmdi)
	if err != nil {
		logger.Error(err, "Failed to teardown resources.")
	}

	return ctrl.Result{}, originalErr
}

func (o Orchestrator) HandleSyncError(ctx context.Context, vmdi *crdv1.VMDiskImage, originalErr error, message string) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Error(originalErr, message)

	o.Recorder.Eventf(vmdi, "Warning", "SyncErrorOccurred", originalErr.Error())

	vmdi.Status.FailureCount += 1
	vmdi.Status.Phase = crdv1.PhaseRetryableFailure
	vmdi.Status.Message = "An error occurred during reconciliation: " + originalErr.Error()

	reason := crdv1.ReasonUnknownSyncFailure
	switch {
	case errors.Is(originalErr, ErrSyncAttemptExceedsMaxDuration):
		reason = crdv1.ReasonSyncAttemptDurationExceeded
	case errors.Is(originalErr, ErrSyncAttemptExceedsRetries):
		reason = crdv1.ReasonRetryLimitExceeded
	case errors.Is(originalErr, ErrMissingSourceArtifact):
		reason = crdv1.ReasonMissingSourceArtifact
	}

	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: originalErr.Error(),
	})

	if err := o.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed after a sync error")
	}

	err := o.Provisioner.TearDownAllResources(ctx, vmdi)
	if err != nil {
		logger.Error(err, "Failed to teardown resources.")
	}

	return ctrl.Result{}, originalErr
}
