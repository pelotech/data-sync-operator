package vmdiservice

import (
	"context"
	crdv1 "pelotech/data-sync-operator/api/v1"
	resourcemanagerservice "pelotech/data-sync-operator/internal/vm-disk-image/service/resource-manager-service"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VMDiskImageService interface {
	IndexVMDiskImageByPhase(rawObj client.Object) []string
	ListVMDiskImagesByPhase(ctx context.Context, c client.Client, phase string) (*crdv1.VMDiskImageList, error)
	QueueResourceCreation(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	AttemptSyncingOfResource(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	TransitonFromSyncing(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	DeleteResource(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error)
	HandleResourceUpdateError(ctx context.Context, ds *crdv1.VMDiskImage, originalErr error, message string) (ctrl.Result, error)
}

type Service struct {
	client.Client
	Recorder         record.EventRecorder
	ResourceManager  resourcemanagerservice.VMDIResourceManager
	ConcurrencyLimit int
	RetryLimit       int
	RetryBackoff     time.Duration
}

func (s *Service) ListVMDiskImagesByPhase(ctx context.Context, phase string) (*crdv1.VMDiskImageList, error) {
	list := &crdv1.VMDiskImageList{}

	listOpts := []client.ListOption{
		client.MatchingFields{".status.phase": phase},
	}

	if err := s.Client.List(ctx, list, listOpts...); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) IndexVMDiskImageByPhase(rawObj client.Object) []string {
	vmdi, ok := rawObj.(*crdv1.VMDiskImage)

	if !ok {
		return nil
	}

	if vmdi.Status.Phase == "" {
		return nil
	}

	return []string{vmdi.Status.Phase}
}

func (s *Service) QueueResourceCreation(ctx context.Context, vmdi *crdv1.VMDiskImage) (ctrl.Result, error) {
	vmdi.Status.Phase = crdv1.VMDiskImagePhaseQueued
	vmdi.Status.Message = "Request is waiting for an available worker."

	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:   crdv1.VMDiskImageTypeReady,
		Status: metav1.ConditionFalse, Reason: "Queued",
		Message: "The sync has been queued for processing.",
	})

	if err := s.Status().Update(ctx, vmdi); err != nil {
		return s.HandleResourceUpdateError(ctx, vmdi, err, "Failed to update status to Queued")
	}

	s.Recorder.Eventf(vmdi, "Normal", "Queued", "Resource successfully queued for sync orchestration")

	return ctrl.Result{Requeue: true}, nil
}

func (s *Service) HandleResourceUpdateError(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
	originalErr error,
	message string,
) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Error(originalErr, message)

	// Mark the resource as Failed
	vmdi.Status.Phase = crdv1.VMDiskImagePhaseFailed
	vmdi.Status.Message = "An error occurred durng reconciliation: " + originalErr.Error()
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.VMDiskImageTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  "UpdateError",
		Message: originalErr.Error(),
	})

	if err := s.Client.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed after an initial update error")
	}

	return ctrl.Result{}, originalErr
}

func (s *Service) HandleResourceCreationError(ctx context.Context, vmdi *crdv1.VMDiskImage, originalErr error) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Handling a reousrce creation failure")
	logger.Error(originalErr, "Failed to create a resource when trying to intiate resource sync")

	s.Recorder.Eventf(vmdi, "Warning", "ResourceCreationFailed", "Failed to create resources.")

	vmdi.Status.Phase = crdv1.VMDiskImagePhaseFailed
	vmdi.Status.Message = "Failed while creating resources: " + originalErr.Error()
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.VMDiskImageTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  "ResourceCreationFailed",
		Message: originalErr.Error(),
	})

	if err := s.Client.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed resource creation failure")
	}

	err := s.ResourceManager.TearDownAllResources(ctx, vmdi)

	if err != nil {
		logger.Error(err, "Failed to teardown resources.")
	}

	return ctrl.Result{}, originalErr
}

func (s *Service) HandleSyncError(ctx context.Context, vmdi *crdv1.VMDiskImage, originalErr error, message string) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Error(originalErr, message)

	s.Recorder.Eventf(vmdi, "Warning", "SyncErrorOccurred", originalErr.Error())

	vmdi.Status.FailureCount += 1

	if err := s.Client.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Failed to update resource failure count")
	}

	if vmdi.Status.FailureCount < s.RetryLimit {
		return ctrl.Result{RequeueAfter: s.RetryBackoff}, nil
	}

	s.Recorder.Eventf(vmdi, "Warning", "SyncExceededRetryCount", "The sync has failed beyond the set retry limit of %d", s.RetryLimit)

	vmdi.Status.Phase = crdv1.VMDiskImagePhaseFailed
	vmdi.Status.Message = "An error occurred durng reconciliation: " + originalErr.Error()
	meta.SetStatusCondition(&vmdi.Status.Conditions, metav1.Condition{
		Type:    crdv1.VMDiskImageTypeFailed,
		Status:  metav1.ConditionTrue,
		Reason:  "SyncFailure",
		Message: originalErr.Error(),
	})

	if err := s.Client.Status().Update(ctx, vmdi); err != nil {
		logger.Error(err, "Could not update status to Failed after an sync error")
	}

	err := s.ResourceManager.TearDownAllResources(ctx, vmdi)

	if err != nil {
		logger.Error(err, "Failed to teardown resources.")
	}

	return ctrl.Result{}, originalErr
}
