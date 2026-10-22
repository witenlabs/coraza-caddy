// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

//go:build coraza.no_memoize

package coraza

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildQualification(t *testing.T) {
	m := &corazaModule{Directives: "SecRuleEngine On"}
	require.NoError(t, m.Provision(reloadContext(t)))
	require.NotNil(t, m.waf)
	require.NoError(t, m.Cleanup())
}
