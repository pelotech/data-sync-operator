package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// CoreConfig holds all the configuration settings for the operator
// that are not specific to any controller in particular
type CoreConfig struct {
	MetricsAddr          string
	MetricsCertPath      string
	MetricsCertName      string
	MetricsCertKey       string
	WebhookCertPath      string
	WebhookCertName      string
	WebhookCertKey       string
	EnableLeaderElection bool
	ProbeAddr            string
	SecureMetrics        bool
	EnableHTTP2          bool
}

// LoadCoreConfigFromEnv reads the configuration from environment variables.
// If an environment variable is not set, it uses the specified default value.
func LoadCoreConfigFromEnv() CoreConfig {
	cfg := CoreConfig{}

	// --- General Configuration ---

	// METRICS_BIND_ADDRESS: The address the metrics endpoint binds to.  Default: "0"
	// Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service
	cfg.MetricsAddr = GetStringEnvOrDefault("METRICS_BIND_ADDRESS", "0")

	// HEALTH_PROBE_BIND_ADDRESS: The address the probe endpoint binds to. Default: ":8081"
	cfg.ProbeAddr = GetStringEnvOrDefault("HEALTH_PROBE_BIND_ADDRESS", ":8081")

	// LEADER_ELECT: Enable leader election for controller manager. Default: false
	cfg.EnableLeaderElection = GetBoolEnvOrDefault("LEADER_ELECT", false)

	// METRICS_SECURE: If set, the metrics endpoint is served securely via HTTPS. Default: true
	cfg.SecureMetrics = GetBoolEnvOrDefault("METRICS_SECURE", true)

	// ENABLE_HTTP2: If set, HTTP/2 will be enabled for the metrics and webhook servers. Default: false
	cfg.EnableHTTP2 = GetBoolEnvOrDefault("ENABLE_HTTP2", false)

	// --- Webhook Configuration ---

	// WEBHOOK_CERT_PATH: The directory that contains the webhook certificate. Default: ""
	cfg.WebhookCertPath = GetStringEnvOrDefault("WEBHOOK_CERT_PATH", "")

	// WEBHOOK_CERT_NAME: The name of the webhook certificate file. Default: "tls.crt"
	cfg.WebhookCertName = GetStringEnvOrDefault("WEBHOOK_CERT_NAME", "tls.crt")

	// WEBHOOK_CERT_KEY: The name of the webhook key file. Default: "tls.key"
	cfg.WebhookCertKey = GetStringEnvOrDefault("WEBHOOK_CERT_KEY", "tls.key")

	// --- Metrics TLS Configuration ---

	// METRICS_CERT_PATH: The directory that contains the metrics server certificate. Default: ""
	cfg.MetricsCertPath = GetStringEnvOrDefault("METRICS_CERT_PATH", "")

	// METRICS_CERT_NAME: The name of the metrics server certificate file. Default: "tls.crt"
	cfg.MetricsCertName = GetStringEnvOrDefault("METRICS_CERT_NAME", "tls.crt")

	// METRICS_CERT_KEY: The name of the metrics server key file. Default: "tls.key"
	cfg.MetricsCertKey = GetStringEnvOrDefault("METRICS_CERT_KEY", "tls.key")

	return cfg
}

// LogLevelOptions maps the KubeBuilder flag to an environment variable.
func LoadLoggerOptionsFromEnv() zap.Options {

	// Default to "debug" if the environment variable is not set
	logLevel := GetStringEnvOrDefault("LOG_LEVEL", "debug")

	opts := zap.Options{}
	opts.Development = false // Default to production mode

	// Map the environment variable string to the core zapcore.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		// Setting Development=true here gives you a more human-readable,
		// non-JSON output format, which is standard for debugging.
		opts.Development = true
		opts.Level = zapcore.DebugLevel
	case "info":
		opts.Level = zapcore.InfoLevel
	case "warn", "warning":
		opts.Level = zapcore.WarnLevel
	case "error":
		opts.Level = zapcore.ErrorLevel
	}

	return opts
}

func GetStringEnvOrDefault(name, defaultValue string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}
	return defaultValue
}

func GetBoolEnvOrDefault(name string, defaultValue bool) bool {
	if valueStr, ok := os.LookupEnv(name); ok {
		// strconv.ParseBool is strict and only accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
		value, err := strconv.ParseBool(valueStr)
		if err != nil {
			panic(fmt.Sprintf("invalid boolean format for environment variable %s='%s': %v", name, valueStr, err))
		}
		return value
	}
	return defaultValue
}

func GetIntEnvOrDefault(name string, defaultValue int) int {
	if valueStr, ok := os.LookupEnv(name); ok {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			panic(fmt.Sprintf("invalid integer format for environment variable %s='%s': %v", name, valueStr, err))
		}
		return value
	}
	return defaultValue
}

func GetDurationEnvOrDefault(name string, defaultValue time.Duration) time.Duration {
	if valueStr, ok := os.LookupEnv(name); ok {
		value, err := time.ParseDuration(valueStr)
		if err != nil {
			panic(fmt.Sprintf("invalid duration format for environment variable %s='%s': %v", name, valueStr, err))
		}
		return value
	}
	return defaultValue
}
