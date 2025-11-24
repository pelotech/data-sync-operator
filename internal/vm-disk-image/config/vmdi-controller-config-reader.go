package config

import (
	corecfg "pelotech/data-sync-operator/internal/core/config"
	"time"
)

const (
	defaultConcurrency     = 10 // TODO: We will need to tune this default
	defaultRetryLimit      = 2
	defaultBackoffDuration = 10 * time.Second
	defaultMaxSyncDuration = 1 * time.Hour
)

type VMDiskImageControllerConfig struct {
	Concurrency          int
	RetryLimit           int
	RetryBackoffDuration time.Duration
	MaxSyncDuration      time.Duration
}

// This function will allow us to get the required config variables from the environment.
// Locally this is your "env" and in production these values will come from a configmap
func LoadVMDIControllerConfigFromEnv() VMDiskImageControllerConfig {
	// The max amount of VMDIs we can have syncing at one time.
	concurrency := corecfg.GetIntEnvOrDefault("CONCURRENCY", defaultConcurrency)

	// How many times we will retry a failed sync.
	retryLimit := corecfg.GetIntEnvOrDefault("RETRY_LIMIT", defaultRetryLimit)

	// How long we want to wait before trying to resync a failed VMDI.
	retryBackoffDuration := corecfg.GetDurationEnvOrDefault("RETRY_BACKOFF_DURATION", defaultBackoffDuration)

	// How long we will let a VMDI sit in syncing status.
	maxSyncDuration := corecfg.GetDurationEnvOrDefault("MAX_SYNC_DURATION", defaultMaxSyncDuration)

	return VMDiskImageControllerConfig{
		Concurrency:          concurrency,
		RetryLimit:           retryLimit,
		RetryBackoffDuration: retryBackoffDuration,
		MaxSyncDuration:      maxSyncDuration,
	}
}
