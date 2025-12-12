package service

import (
	"context"
	"errors"
	"fmt"
	crdv1 "pelotech/data-sync-operator/api/v1alpha1"
	"strings"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VMDiskImageProvisioner interface {
	CreateResources(ctx context.Context, resource *crdv1.VMDiskImage) error
	TearDownAllResources(ctx context.Context, resource *crdv1.VMDiskImage) error
	ResourcesAreReady(ctx context.Context, resource *crdv1.VMDiskImage) (bool, error)
	ResourcesHaveErrors(ctx context.Context, resource *crdv1.VMDiskImage) error
}

type K8sVMDIProvisioner struct {
	client.Client
	ResourceGenerator      VMDIResourceGenerator
	MaxSyncAttemptDuration time.Duration
	MaxRetryPerAttempt     int
}

const dataVolumeDonePhase = "Succeeded"

var ErrMissingSourceArtifact = errors.New("the requested artifact does not exist")
var ErrSyncAttemptExceedsRetries = errors.New("the sync attempt has failed beyond the retry limit")
var ErrSyncAttemptExceedsMaxDuration = errors.New("the sync attempt has lasted beyond its max duration")

// Create resources for a given VMDiskImage. Stops creating them if
// a single resource fails to create. Does not cleanup after itself
func (p K8sVMDIProvisioner) CreateResources(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) error {
	logger := logf.FromContext(ctx)

	vs, dv, err := p.ResourceGenerator.CreateStorageManifests(vmdi)
	if err != nil {
		logger.Error(err, "Failed to create the storage manifests for VMDiskImage", vmdi.Name)
		return err
	}

	// Create the Data volume
	err = p.Patch(ctx, dv, client.Apply, client.FieldOwner(crdv1.VMDiskImageControllerName), client.ForceOwnership)
	if err != nil {
		logger.Error(err, "Failed to create the backing datavolume for ", vmdi.Name, " within the cluster")
		return err
	}

	// Create the volume snapshot
	err = p.Patch(ctx, vs, client.Apply, client.FieldOwner(crdv1.VMDiskImageControllerName), client.ForceOwnership)
	if err != nil {
		logger.Error(err, "Failed to create the backing volumesnapshot for ", vmdi.Name, " within the cluster")
		return err
	}

	return err
}

// Tear down the resources associated with a given VMDiskImage.
func (p K8sVMDIProvisioner) TearDownAllResources(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) error {
	deleteByLabels := getLabelsToMatch(vmdi)

	// First we tear down the PVCs that back the data volumes
	err := p.DeleteAllOf(
		ctx,
		&corev1.PersistentVolumeClaim{},
		client.InNamespace(vmdi.Namespace),
		deleteByLabels,
	)
	if err != nil {
		return err
	}

	// Next the Datavolumes
	err = p.DeleteAllOf(
		ctx,
		&cdiv1beta1.DataVolume{},
		client.InNamespace(vmdi.Namespace),
		deleteByLabels,
	)
	if err != nil {
		return err
	}

	// Finally the volumesnapshots
	err = p.DeleteAllOf(
		ctx,
		&snapshotv1.VolumeSnapshot{},
		client.InNamespace(vmdi.Namespace),
		deleteByLabels,
	)
	if err != nil {
		return err
	}

	// If we have a finalizer remove it.
	if crutils.ContainsFinalizer(vmdi, crdv1.VMDiskImageFinalizer) {
		crutils.RemoveFinalizer(vmdi, crdv1.VMDiskImageFinalizer)
		if err := p.Update(ctx, vmdi); err != nil {
			return err
		}
	}

	return nil
}

// This function will check if the datavolumes assoicated with our VMDiskImage
// are ready. Currently we only check if the datavolumes are done syncing in our
// manual process. We do not check if any of the other resources are ready.
func (p K8sVMDIProvisioner) ResourcesAreReady(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) (bool, error) {
	searchLabels := getLabelsToMatch(vmdi)
	listOps := []client.ListOption{
		searchLabels,
	}

	dataVolumeList := &cdiv1beta1.DataVolumeList{}
	if err := p.List(ctx, dataVolumeList, listOps...); err != nil {
		return false, fmt.Errorf("failed to list data volumes with the vm disk image %s: %w", vmdi.Name, err)
	}

	dataVolumesReady := true
	for _, dv := range dataVolumeList.Items {
		if dv.Status.Phase != dataVolumeDonePhase {
			dataVolumesReady = false
			break
		}
	}

	return dataVolumesReady, nil
}

// Check if our resources have errors that would require us to
// scuttle the sync.
func (p K8sVMDIProvisioner) ResourcesHaveErrors(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) error {
	logger := logf.FromContext(ctx)

	condition := meta.FindStatusCondition(vmdi.Status.Conditions, crdv1.ConditionTypeReady)
	if condition == nil || condition.Reason != crdv1.ReasonSyncing {
		return fmt.Errorf("the VMDiskImage %s has no condition or is it's condition reason is not syncing", vmdi.Name)
	}

	now := time.Now()
	syncStartTime := condition.LastTransitionTime.Time

	var timeSyncing time.Duration
	if now.Before(syncStartTime) {
		skew := syncStartTime.Sub(now)
		logger.Info("Clock Skew Detected: Node time is behind resource start time",
			"node_time", now,
			"resource_start_time", syncStartTime,
			"skew_duration", skew,
		)

		// In the case of clock skew let the resource continue on. Don't fail it
		timeSyncing = 0
	} else {
		// Normal calculation
		timeSyncing = now.Sub(syncStartTime)
	}
	if timeSyncing > p.MaxSyncAttemptDuration {
		return ErrSyncAttemptExceedsMaxDuration
	}

	searchLabels := getLabelsToMatch(vmdi)
	listOps := []client.ListOption{
		searchLabels,
	}
	dataVolumeList := &cdiv1beta1.DataVolumeList{}

	if err := p.List(ctx, dataVolumeList, listOps...); err != nil {
		return fmt.Errorf("failed to list datavolumes with the VMDiskImage %s: %w", vmdi.Name, err)
	}

	for _, dv := range dataVolumeList.Items {
		for _, cond := range dv.Status.Conditions {
			missingSourceArtifact := strings.Contains(cond.Message, "404") || strings.Contains(strings.ToLower(cond.Message), "not found")
			if missingSourceArtifact {
				return ErrMissingSourceArtifact
			}
		}

		if dv.Status.RestartCount >= int32(p.MaxRetryPerAttempt) {
			return ErrSyncAttemptExceedsRetries
		}

	}

	return nil
}

func getLabelsToMatch(vmdi *crdv1.VMDiskImage) client.MatchingLabels {
	labelsToMatch := map[string]string{
		crdv1.VMDiskImageOwnerLabel: vmdi.Name,
	}

	return client.MatchingLabels(labelsToMatch)
}
