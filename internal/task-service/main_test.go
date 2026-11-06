package main

import (
	"strings"
	"testing"
)

func TestRunRequiresDatabaseServiceAddress(t *testing.T) {
	t.Setenv(databaseServiceAddressEnv, "")

	err := run()

	if err == nil || !strings.Contains(err.Error(), databaseServiceAddressEnv) {
		t.Fatalf("run error = %v, want missing %s error", err, databaseServiceAddressEnv)
	}
}
