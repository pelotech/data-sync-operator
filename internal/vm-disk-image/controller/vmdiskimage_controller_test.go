// TODO: Make these tests meaningful and work post MVP
// /*
// Copyright 2025.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package vmdiskimagectrl

// import (
// 	"context"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	"k8s.io/apimachinery/pkg/api/errors"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/tools/record"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"

// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// 	crdv1alpha1 "pelotech/data-sync-operator/api/v1alpha1"
// 	vmdicfg "pelotech/data-sync-operator/internal/vm-disk-image/config"
// 	vmdi "pelotech/data-sync-operator/internal/vm-disk-image/service"
// )

// var _ = Describe("VMDiskImage Controller", func() {
// 	Context("When reconciling a resource", func() {
// 		const resourceName = "test-resource"

// 		ctx := context.Background()

// 		typeNamespacedName := types.NamespacedName{
// 			Name:      resourceName,
// 			Namespace: "default", // TODO(user):Modify as needed
// 		}
// 		VMDiskImage := &crdv1alpha1.VMDiskImage{}

// 		snapshotClass := "daily-snapshots"

// 		VMDiskImage.Spec = crdv1alpha1.VMDiskImageSpec{
// 			SecretRef:     "test-secret",
// 			SourceType:    "blank",
// 			DiskSize:      "1Mi",
// 			SnapshotClass: &snapshotClass,
// 			URL:           "test",
// 		}

// 		BeforeEach(func() {
// 			By("creating the custom resource for the Kind VMDiskImage")
// 			err := k8sClient.Get(ctx, typeNamespacedName, VMDiskImage)
// 			if err != nil && errors.IsNotFound(err) {
// 				resource := &crdv1alpha1.VMDiskImage{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Name:      resourceName,
// 						Namespace: "default",
// 					},
// 					// TODO(user): Specify other spec details if needed.
// 				}
// 				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
// 			}
// 		})

// 		AfterEach(func() {
// 			// TODO(user): Cleanup logic after each test, like removing the resource instance.
// 			resource := &crdv1alpha1.VMDiskImage{}

// 			snapshotClass := "daily-snapshots"

// 			VMDiskImage.Spec = crdv1alpha1.VMDiskImageSpec{
// 				SecretRef:     "test-secret",
// 				SourceType:    "blank",
// 				DiskSize:      "1Mi",
// 				SnapshotClass: &snapshotClass,
// 				URL:           "test",
// 			}

// 			err := k8sClient.Get(ctx, typeNamespacedName, resource)

// 			Expect(err).NotTo(HaveOccurred())

// 			By("Cleanup the specific resource instance VMDiskImage")
// 			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
// 		})

// 		It("should successfully reconcile the resource", func() {
// 			By("Reconciling the created resource")

// 			client := k8sClient

// 			config := vmdicfg.LoadVMDIControllerConfigFromEnv()

// 			resourceGenerator := &vmdi.Generator{}
// 			vmdiProvisioner := vmdi.K8sVMDIProvisioner{
// 				Client:            client,
// 				ResourceGenerator: resourceGenerator,
// 				MaxSyncDuration:   config.MaxSyncDuration,
// 				RetryLimit:        config.RetryLimit,
// 			}

// 			recorder := &record.FakeRecorder{}

// 			orchestrator := vmdi.Orchestrator{
// 				Client:       client,
// 				Recorder:     recorder,
// 				Provisioner:  vmdiProvisioner,
// 				RetryLimit:   config.RetryLimit,
// 				RetryBackoff: config.RetryBackoffDuration,
// 				SyncLimit:    config.Concurrency,
// 			}

// 			controllerReconciler := &VMDiskImageReconciler{
// 				Scheme:                  k8sClient.Scheme(),
// 				VMDiskImageOrchestrator: orchestrator,
// 			}

// 			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
// 				NamespacedName: typeNamespacedName,
// 			})

// 			Expect(err).NotTo(HaveOccurred())
// 			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
// 			// Example: If you expect a certain status condition after reconciliation, verify it here.
// 		})
// 	})
// })
