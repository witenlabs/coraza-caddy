// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

//go:build !coraza.no_memoize

package coraza

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildQualification(t *testing.T) {
	m := &corazaModule{Directives: "SecRuleEngine On"}
	require.ErrorContains(t, m.Provision(reloadContext(t)), "coraza.no_memoize")
	require.Nil(t, m.waf, "unsupported build must not activate an engine")
	require.NoError(t, m.Cleanup())
}
