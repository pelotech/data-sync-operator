package config

import (
	corecfg "pelotech/data-sync-operator/internal/core/config"
	"time"
)

const (
	defaultConcurrency                 = 10 // TODO: We will need to tune this default
	defaultSyncRetryLimit              = 2
	defaultSyncBackoffDuration         = 10 * time.Second
	defaultMaxSyncDuration             = 1 * time.Hour
	defaultResurrectionTimeout         = 8 * time.Hour
	defaultResurrectionBackoffDuration = 30 * time.Minute
)

type VMDiskImageControllerConfig struct {
	Concurrency                 int
	SyncRetryLimit              int
	SyncRetryBackoffDuration    time.Duration
	MaxSyncDuration             time.Duration
	ResurrectionTimeout         time.Duration
	ResurrectionBackoffDuration time.Duration
}

// This function will allow us to get the required config variables from the environment.
// Locally this is your "env" and in production these values will come from a configmap
func LoadVMDIControllerConfigFromEnv() VMDiskImageControllerConfig {
	// The max amount of VMDIs we can have syncing at one time.
	concurrency := corecfg.GetIntEnvOrDefault("VMDI_SYNC_CONCURRENCY", defaultConcurrency)

	// How many times we will retry a failed sync.
	retryLimit := corecfg.GetIntEnvOrDefault("SYNC_RETRY_LIMIT", defaultSyncRetryLimit)

	// How long we want to wait before trying another sync on a VMDI in SYNCING status if it fails.
	retryBackoffDuration := corecfg.GetDurationEnvOrDefault("SYNC_RETRY_BACKOFF_DURATION", defaultSyncBackoffDuration)

	// How long we will let a VMDI sit in syncing status.
	maxSyncDuration := corecfg.GetDurationEnvOrDefault("MAX_SYNC_DURATION", defaultMaxSyncDuration)

	// How long we will attempt to resend a VMDI through loop until it is permanently failed
	resurrectionTimeout := corecfg.GetDurationEnvOrDefault("VMDI_RESURRECTION_TIMEOUT", defaultResurrectionTimeout)

	// How long we will wait before attempting to send a VMDI through the loop if it enters a failed state
	resurrectionBackoffDuration := corecfg.GetDurationEnvOrDefault("VMDI_RESURRECTION_BACKOFF_DURATION", defaultResurrectionBackoffDuration)

	return VMDiskImageControllerConfig{
		Concurrency:                 concurrency,
		SyncRetryLimit:              retryLimit,
		SyncRetryBackoffDuration:    retryBackoffDuration,
		MaxSyncDuration:             maxSyncDuration,
		ResurrectionTimeout:         resurrectionTimeout,
		ResurrectionBackoffDuration: resurrectionBackoffDuration,
	}
}
