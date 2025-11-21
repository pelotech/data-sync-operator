package service

import (
	"context"
	"fmt"
	crdv1 "pelotech/data-sync-operator/api/v1alpha1"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
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
	ResourceGenerator VMDIResourceGenerator
	MaxSyncDuration   time.Duration
	RetryLimit        int
}

const dataVolumeDonePhase = "Succeeded"

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
	ds *crdv1.VMDiskImage,
) error {
	deleteByLabels := getLabelsToMatch(ds)

	// First we tear down the PVCs that back the data volumes
	err := p.DeleteAllOf(
		ctx,
		&corev1.PersistentVolumeClaim{},
		client.InNamespace(ds.Namespace),
		deleteByLabels,
	)

	if err != nil {
		return err
	}

	// Next the Datavolumes
	err = p.DeleteAllOf(
		ctx,
		&cdiv1beta1.DataVolume{},
		client.InNamespace(ds.Namespace),
		deleteByLabels,
	)

	if err != nil {
		return err
	}

	// Finally the volumesnapshots
	err = p.DeleteAllOf(
		ctx,
		&snapshotv1.VolumeSnapshot{},
		client.InNamespace(ds.Namespace),
		deleteByLabels,
	)

	if err != nil {
		return err
	}

	// If we have a finalizer remove it.
	if crutils.ContainsFinalizer(ds, crdv1.VMDiskImageFinalizer) {
		crutils.RemoveFinalizer(ds, crdv1.VMDiskImageFinalizer)
		if err := p.Update(ctx, ds); err != nil {
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
	ds *crdv1.VMDiskImage,
) error {
	// Check if our VMDiskImage has been syncing for too long
	now := time.Now()

	syncStartTimeStr, exists := ds.Annotations[crdv1.SyncStartTimeAnnotation]

	if !exists {
		return fmt.Errorf("the VMDiskImage %s does not have a recorded sync start time.", ds.Name)
	}

	syncStartTime, err := time.Parse(time.RFC3339, syncStartTimeStr)

	if err != nil {
		return fmt.Errorf("the VMDiskImage %s does not have a parseable sync start time.", ds.Name)
	}

	timeSyncing := now.Sub(syncStartTime)

	if timeSyncing > p.MaxSyncDuration {
		return fmt.Errorf("the VMDiskImage %s has been syncing longer than the allowed sync time.", ds.Name)
	}

	searchLabels := getLabelsToMatch(ds)

	listOps := []client.ListOption{
		searchLabels,
	}

	dataVolumeList := &cdiv1beta1.DataVolumeList{}

	if err := p.List(ctx, dataVolumeList, listOps...); err != nil {
		return fmt.Errorf("failed to list datavolumes with the VMDiskImage %s: %w", ds.Name, err)
	}

	for _, dv := range dataVolumeList.Items {
		if dv.Status.RestartCount >= int32(p.RetryLimit) {
			return fmt.Errorf("a datavolume has restarted more than the max for a sync.")
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
