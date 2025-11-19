package service

// This service will act to simulate importing resources in the deployed cluster.
// When running locally we can just allocate empty disks so this will ensure that
// we still get some sync failures but it will let us know that they are correct.
// This is to test the error handling component of the application. I'll work on this
// while we review the PR
type LocalVMDIOrchestrator struct {
	VMDiskImageOrchestrator
}
