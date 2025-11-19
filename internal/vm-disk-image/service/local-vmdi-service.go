package service

import (
	"context"
	"errors"
	"math/rand"
	crdv1 "pelotech/data-sync-operator/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

// This service will act to simulate importing resources in the deployed cluster.
// When running locally we can just allocate empty disks so this will ensure that
// we still get some sync failures but it will let us know that they are correct.
// This is to test the error handling component of the application
type LocalVMDIOrchestrator struct {
	VMDiskImageOrchestrator
}

func (o LocalVMDIOrchestrator) AttemptSyncingOfResource(
	ctx context.Context,
	vmdi *crdv1.VMDiskImage,
) (ctrl.Result, error) {
	res, err := o.AttemptSyncingOfResource(ctx, vmdi)
	// Generate a random number between 1 and 10 (inclusive).
	// rand.Intn(10) generates a number in the range [0, 9].
	// Adding 1 shifts the range to [1, 10].
	randomNumber := rand.Intn(10)

	if err != nil {
		return res, err
	}

	// Check if the number is greater than 5.
	if randomNumber > 5 {
		return o.HandleSyncError(ctx, vmdi, errors.New("forced error"), "Forced")
	} else {
		return res, err
	}
}
