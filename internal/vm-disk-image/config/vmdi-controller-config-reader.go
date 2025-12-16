package config

import (
	corecfg "pelotech/data-sync-operator/internal/core/config"
	"time"
)

const (
	defaultConcurrency            = 5 // TODO: We will need to tune this default
	defaultMaxBackoffDelay        = 1 * time.Hour
	defaultMaxSyncDuration        = 12 * time.Hour
	defaultMaxSyncAttemptRetries  = 3
	defaultMaxSyncAttemptDuration = 1 * time.Hour
)

type VMDiskImageControllerConfig struct {
	Concurrency            int
	MaxBackoffDelay        time.Duration
	MaxSyncDuration        time.Duration
	MaxSyncAttemptDuration time.Duration
	MaxSyncAttemptRetry    int
}

// This function will allow us to get the required config variables from the environment.
// Locally this is your "env" and in production these values will come from a configmap
func LoadVMDIControllerConfigFromEnv() VMDiskImageControllerConfig {
	// The max amount of VMDIs we can have syncing at one time.
	concurrency := corecfg.GetIntEnvOrDefault("MAX_VMDI_SYNC_CONCURRENCY", defaultConcurrency)

	// The longest we will ever wait to retry.
	maxBackoffDelay := corecfg.GetDurationEnvOrDefault("MAX_SYNC_RETRY_BACKOFF_DURATION", defaultMaxBackoffDelay)

	// How long we will try to run a sync before we fail it forever.
	maxSyncDuration := corecfg.GetDurationEnvOrDefault("MAX_SYNC_DURATION", defaultMaxSyncDuration)

	// How long we will let a VMDI sit in syncing status.
	maxAttemptDuration := corecfg.GetDurationEnvOrDefault("MAX_SYNC_ATTEMPT_DURATION", defaultMaxSyncAttemptDuration)

	// How many times we will retry on a given attempt.
	maxRetriesPerAttempt := corecfg.GetIntEnvOrDefault("MAX_SYNC_ATTEMPT_RETRIES", defaultMaxSyncAttemptRetries)

	return VMDiskImageControllerConfig{
		Concurrency:            concurrency,
		MaxBackoffDelay:        maxBackoffDelay,
		MaxSyncAttemptDuration: maxAttemptDuration,
		MaxSyncAttemptRetry:    maxRetriesPerAttempt,
		MaxSyncDuration:        maxSyncDuration,
	}
}
