package config

import (
	corecfg "pelotech/data-sync-operator/internal/core/config"
	"time"
)

const (
	defaultConcurrency     = 10
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
	concurrency := corecfg.GetIntEnvOrDefault("CONCURRENCY", defaultConcurrency)

	retryLimit := corecfg.GetIntEnvOrDefault("RETRY_LIMIT", defaultRetryLimit)

	retryBackoffDuration := corecfg.GetDurationEnvOrDefault("RETRY_BACKOFF_DURATION", defaultBackoffDuration)

	maxSyncDuration := corecfg.GetDurationEnvOrDefault("MAX_SYNC_DURATION", defaultMaxSyncDuration)

	return VMDiskImageControllerConfig{
		Concurrency:          concurrency,
		RetryLimit:           retryLimit,
		RetryBackoffDuration: retryBackoffDuration,
		MaxSyncDuration:      maxSyncDuration,
	}
}
