//nolint:all // This is all test code so we don't care

package common

import (
	"flag"
	"os"
)

func ConfigFromEnv() Config {
	if !flag.Parsed() {
		flag.Parse()
	}

	cfg := Config{
		ProjectName: "backend-tests",
	}

	if s := os.Getenv("PROJECT_NAME"); s != "" {
		cfg.ProjectName = s
	}

	if os.Getenv("SKIP_CLEANUP") != "" {
		cfg.SkipCleanup = true
	}

	if os.Getenv("SKIP_CLEANUP_ON_FAILURE") != "" {
		cfg.SkipCleanupOnFailure = true
	}

	cfg.PrintServiceLogsOnFailure = os.Getenv("PRINT_SERVICE_LOGS_ON_FAILURE")
	cfg.PrintServiceStatusOnFailure = os.Getenv("PRINT_SERVICE_STATUS_ON_FAILURE")

	return cfg
}
