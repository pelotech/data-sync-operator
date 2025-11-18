package resourcemanagerservice

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	crdv1 "pelotech/data-sync-operator/api/v1"
	resourcegenservice "pelotech/data-sync-operator/internal/vm-disk-image/service/resource-gen-service"

	corev1 "k8s.io/api/core/v1"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	crutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type VMDIResourceManager interface {
	CreateResources(ctx context.Context, resource *crdv1.VMDiskImage) error
	TearDownAllResources(ctx context.Context, resource *crdv1.VMDiskImage) error
	ResourcesAreReady(ctx context.Context, resource *crdv1.VMDiskImage) (bool, error)
	ResourcesHaveErrors(ctx context.Context, resource *crdv1.VMDiskImage) error
}

type Manager struct {
	K8sClient         client.Client
	ResourceGenerator resourcegenservice.VMDIResourceGenerator
	MaxSyncDuration   time.Duration
	RetryLimit        int
}

const dataVolumeDonePhase = "Succeeded"

// Create resources for a given VMDiskImage. Stops creating them if
// a single resource fails to create. Does not cleanup after itself
func (m Manager) CreateResources(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) error {
	vs, dv, err := m.ResourceGenerator.CreateStorageManifests(vmdi)

	if err != nil {
		return err
	}

	err = m.K8sClient.Patch(ctx, dv, client.Apply, client.FieldOwner("data-sync-operator"))

	if err != nil {
		return err
	}

	err = m.K8sClient.Patch(ctx, vs, client.Apply, client.FieldOwner("data-sync-operator"))

	if err != nil {
		return err
	}

	return nil
}

// Tear down the resources associated with a given VMDiskImage.
func (m Manager) TearDownAllResources(
	ctx context.Context,
	ds *crdv1.VMDiskImage,
) error {
	deleteByLabels := getLabelsToMatch(ds)

	// First we tear down the PVCs that back the data volumes
	err := m.K8sClient.DeleteAllOf(
		ctx,
		&corev1.PersistentVolumeClaim{},
		client.InNamespace(ds.Namespace),
		deleteByLabels,
	)

	if err != nil {
		return err
	}

	// Next the Datavolumes
	err = m.K8sClient.DeleteAllOf(
		ctx,
		&cdiv1beta1.DataVolume{},
		client.InNamespace(ds.Namespace),
		deleteByLabels,
	)

	if err != nil {
		return err
	}

	// Finally the volumesnapshots
	err = m.K8sClient.DeleteAllOf(
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
		if err := m.K8sClient.Update(ctx, ds); err != nil {
			return err
		}
	}

	return nil
}

// This function will check if the datavolumes assoicated with our VMDiskImage
// are ready. Currently we only check if the datavolumes are done syncing in our
// manual process. We do not check if any of the other resources are ready.
func (m Manager) ResourcesAreReady(
	ctx context.Context,
	ds *crdv1.VMDiskImage,
) (bool, error) {

	searchLabels := getLabelsToMatch(ds)

	listOps := []client.ListOption{
		searchLabels,
	}

	dataVolumeList := &cdiv1beta1.DataVolumeList{}

	if err := m.K8sClient.List(ctx, dataVolumeList, listOps...); err != nil {
		return false, fmt.Errorf("failed to list data volumes with the vm disk image %s: %w", ds.Name, err)
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
func (m Manager) ResourcesHaveErrors(
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

	if timeSyncing > m.MaxSyncDuration {
		return fmt.Errorf("the VMDiskImage %s has been syncing longer than the allowed sync time.", ds.Name)
	}

	searchLabels := getLabelsToMatch(ds)

	listOps := []client.ListOption{
		searchLabels,
	}

	dataVolumeList := &cdiv1beta1.DataVolumeList{}

	if err := m.K8sClient.List(ctx, dataVolumeList, listOps...); err != nil {
		return fmt.Errorf("failed to list datavolumes with the VMDiskImage %s: %w", ds.Name, err)
	}

	for _, dv := range dataVolumeList.Items {
		if dv.Status.RestartCount >= int32(m.RetryLimit) {
			return fmt.Errorf("a datavolume has restarted more than the max for a sync.")
		}
	}

	return nil
}

func getLabelsToMatch(ds *crdv1.VMDiskImage) client.MatchingLabels {
	labelsToMatch := map[string]string{
		crdv1.VMDiskImageOwnerLabel: ds.Name,
	}

	return client.MatchingLabels(labelsToMatch)
}
