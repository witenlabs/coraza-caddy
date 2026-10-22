// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

package coraza

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Verify actual descriptor ownership, including a transaction that outlives Cleanup.
func TestAuditDescriptorsAcrossReloads(t *testing.T) {
	if _, err := os.ReadDir("/proc/self/fd"); err != nil {
		t.Skip("Linux procfs required for descriptor ownership check")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	count := func() int {
		entries, err := os.ReadDir("/proc/self/fd")
		require.NoError(t, err)
		n := 0
		for _, entry := range entries {
			target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
			if err == nil && strings.HasPrefix(target, path) {
				n++
			}
		}
		return n
	}
	directives := fmt.Sprintf("SecRuleEngine DetectionOnly\nSecAuditEngine On\nSecAuditLogParts AZ\nSecAuditLog %q\n", path)
	active := provisionReloadModule(t, directives, nil)
	require.Equal(t, 1, count())
	for generation := 0; generation < 64; generation++ {
		txID := fmt.Sprintf("generation-%d", generation)
		tx := active.waf.NewTransactionWithID(txID)
		rotated := ""
		if generation%16 == 0 {
			rotated = fmt.Sprintf("%s.%d", path, generation)
			require.NoError(t, os.Rename(path, rotated))
		}
		candidate := provisionReloadModule(t, directives, nil)
		require.NoError(t, active.Cleanup())
		require.Equal(t, 2, count(), "pending transaction must retain its writer")
		tx.ProcessLogging()
		require.NoError(t, tx.Close())
		require.Equal(t, 1, count(), "retired generation must release its descriptor")
		writtenPath := path
		if rotated != "" {
			writtenPath = rotated
		}
		data, err := os.ReadFile(writtenPath)
		require.NoError(t, err)
		require.Contains(t, string(data), txID, "old generation must finish logging, including after rotation")
		if generation%8 == 0 {
			failed := &corazaModule{Directives: directives + "SecRequestBodyLimit 1\nSecRequestBodyInMemoryLimit 2\n"}
			require.Error(t, failed.Provision(reloadContext(t)))
			require.NoError(t, failed.Cleanup())
			require.Equal(t, 1, count(), "failed constructor must release the opened writer")
		}
		active = candidate
	}
	require.NoError(t, active.Cleanup())
	require.Zero(t, count(), "all audit descriptors must close when the final generation retires")
}
