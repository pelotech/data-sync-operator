package resourcegenservice

import (
	"errors"

	crdv1 "pelotech/data-sync-operator/api/v1"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VMDIResourceGenerator interface {
	CreateStorageManifests(vmdi *crdv1.VMDiskImage) (*snapshotv1.VolumeSnapshot, *cdiv1beta1.DataVolume, error)
}

type Generator struct{}

func (g *Generator) CreateStorageManifests(
	vmdi *crdv1.VMDiskImage,
) (*snapshotv1.VolumeSnapshot, *cdiv1beta1.DataVolume, error) {
	volumeSnapshot := createVolumeSnapshot(vmdi)

	dataVolume, err := createDataVolume(vmdi)

	if err != nil {
		return nil, nil, err
	}

	return volumeSnapshot, dataVolume, nil
}

func createDataVolume(vmdi *crdv1.VMDiskImage) (*cdiv1beta1.DataVolume, error) {
	blockOwnerDeletion := true
	ownerReferences := []metav1.OwnerReference{
		{
			APIVersion:         "v1",
			BlockOwnerDeletion: &blockOwnerDeletion,
			Kind:               "VMDiskImage",
			Name:               vmdi.Name,
			UID:                vmdi.UID,
		},
	}

	meta := metav1.ObjectMeta{
		Name:            vmdi.Spec.Name,
		Namespace:       vmdi.Namespace,
		Labels:          withOperatorLabels(vmdi.Labels, vmdi.Name),
		OwnerReferences: ownerReferences,
		Annotations: map[string]string{
			"cdi.kubevirt.io/storage.bind.immediate.requested": "true",
		},
	}

	diskSizeResource, err := resource.ParseQuantity(vmdi.Spec.DiskSize)

	if err != nil {
		return nil, err
	}

	pvc := &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: diskSizeResource,
			},
		},
	}

	if vmdi.Spec.StorageClass != nil {
		pvc.StorageClassName = vmdi.Spec.StorageClass
	}

	var source *cdiv1beta1.DataVolumeSource

	if vmdi.Spec.SourceType == "s3" {
		source = &cdiv1beta1.DataVolumeSource{
			S3: &cdiv1beta1.DataVolumeSourceS3{
				URL:       vmdi.Spec.URL,
				SecretRef: vmdi.Spec.SecretRef,
			},
		}
	} else {
		if vmdi.Spec.CertConfigMap == nil {
			errMsg := "attempted to create a datavolume without a registry but no certConfigMap was provided"
			return nil, errors.New(errMsg)
		}
		source = &cdiv1beta1.DataVolumeSource{
			Registry: &cdiv1beta1.DataVolumeSourceRegistry{
				URL:           &vmdi.Spec.URL,
				CertConfigMap: vmdi.Spec.CertConfigMap,
				SecretRef:     &vmdi.Spec.SecretRef,
			},
		}
	}

	spec := cdiv1beta1.DataVolumeSpec{
		PVC:    pvc,
		Source: source,
	}

	dv := &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cdi.kubevirt.io/v1beta1",
			Kind:       "DataVolume",
		},
		ObjectMeta: meta,
		Spec:       spec,
	}

	return dv, nil
}

func createVolumeSnapshot(vmdi *crdv1.VMDiskImage) *snapshotv1.VolumeSnapshot {
	blockOwnerDeletion := true
	ownerReferences := []metav1.OwnerReference{
		{
			APIVersion:         "v1",
			BlockOwnerDeletion: &blockOwnerDeletion,
			Kind:               "VMDiskImage",
			Name:               vmdi.Name,
			UID:                vmdi.UID,
		},
	}

	meta := metav1.ObjectMeta{
		Name:            vmdi.Spec.Name,
		Namespace:       vmdi.Namespace,
		Labels:          withOperatorLabels(vmdi.Labels, vmdi.Name),
		OwnerReferences: ownerReferences,
	}

	spec := snapshotv1.VolumeSnapshotSpec{
		Source: snapshotv1.VolumeSnapshotSource{
			PersistentVolumeClaimName: &vmdi.Spec.Name,
		},
	}

	if vmdi.Spec.SnapshotClass != nil {
		spec.VolumeSnapshotClassName = vmdi.Spec.SnapshotClass
	}

	return &snapshotv1.VolumeSnapshot{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "snapshot.storage.k8s.io/v1",
			Kind:       "VolumeSnapshot",
		},
		ObjectMeta: meta,
		Spec:       spec,
	}
}

func withOperatorLabels(labels map[string]string, ownerName string) map[string]string {
	labels[crdv1.VMDiskImageOwnerLabel] = ownerName

	return labels
}
