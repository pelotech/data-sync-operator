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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	VMDiskImageControllerName = "data-sync-operator-vmdi-controller"
)

// Condition types and reasons
const (
	VMDiskImageTypeReady  string = "Ready"
	VMDiskImageTypeFailed string = "Failed"
)

// VMDiskImage Phases
const (
	VMDiskImagePhaseQueued    string = "Queued"
	VMDiskImagePhaseSyncing   string = "Syncing"
	VMDiskImagePhaseCompleted string = "Completed"
	VMDiskImagePhaseFailed    string = "Failed"
)

// VMDiskImage Labels
const (
	VMDiskImageOwnerLabel string = "owner"
)

// VMDiskImage Annotations
const (
	SyncStartTimeAnnotation = "sync-start-time"
)

const VMDiskImageFinalizer = "pelotech.ot/vm-disk-image-finalizer"

// VMDiskImageSpec defines the desired state of VMDiskImage.
type VMDiskImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:required
	// +kubebuilder:validation:minlength=1
	SecretRef string `json:"secretRef"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="not-provided"
	// +optional
	URL string `json:"url"`

	// NOTE: The "blank" type isn't used yet in production but is useful locally. This can be removed if we want
	// +kubebuilder:validation:Enum=s3;registry;blank
	SourceType string `json:"sourceType"`

	// DiskSize specifies the size of the disk, e.g., "10Gi", "500Mi".
	// +kubebuilder:validation:Pattern=`^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$`
	DiskSize string `json:"diskSize"`

	// +kubebuilder:validation:Optional
	StorageClass *string `json:"storageClass,omitempty"`

	// +kubebuilder:validation:Optional
	CertConfigMap *string `json:"certConfigMap,omitempty"`

	// +kubebuilder:validation:Optional
	SnapshotClass *string `json:"snapshotClass,omitempty"`
}

// VMDiskImageStatus defines the observed state of VMDiskImage.
type VMDiskImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=Queued;Syncing;Completed;Failed
	Phase string `json:"phase"`

	// A human-readable message providing more details about the current phase.
	Message string `json:"message,omitempty"`

	// Conditions of the VMDiskImage resource.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	FailureCount int `json:"failureCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=vmdiskimages,scope=Namespaced,shortName=vmdi,singular=vmdiskimage
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The current phase of the VMDiskImage."
// +kubebuilder:printcolumn:name="Resource Name",type="string",JSONPath=".spec.name",description="The name of the resource we are syncing."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VMDiskImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VMDiskImageSpec   `json:"spec,omitempty"`
	Status VMDiskImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type VMDiskImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VMDiskImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VMDiskImage{}, &VMDiskImageList{})
}
