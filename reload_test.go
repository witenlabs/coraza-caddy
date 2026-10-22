// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

package coraza

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/stretchr/testify/require"
)

func reloadContext(t *testing.T) caddy.Context {
	t.Helper()
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	t.Cleanup(cancel)
	return ctx
}

func writeReloadFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func reloadRule(path string, id int) string {
	return fmt.Sprintf("SecRule REQUEST_URI \"@streq %s\" \"id:%d,phase:1,deny,status:403,nolog\"\n", path, id)
}

func provisionReloadModule(t *testing.T, directives string, includes []string) *corazaModule {
	t.Helper()
	m := &corazaModule{Directives: directives, Include: includes}
	require.NoError(t, m.Provision(reloadContext(t)))
	t.Cleanup(func() { require.NoError(t, m.Cleanup()) })
	return m
}

func requestInterruption(t *testing.T, m *corazaModule, path string) *types.Interruption {
	t.Helper()
	tx := m.waf.NewTransaction()
	defer func() { require.NoError(t, tx.Close()) }()
	tx.ProcessURI(path, "GET", "HTTP/1.1")
	return tx.ProcessRequestHeaders()
}

func TestReloadReadsExternalPolicy(t *testing.T) {
	for _, kind := range []string{"include", "deprecated-include", "nested-include", "glob-membership", "phrase-file"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			leaf := filepath.Join(dir, "leaf.conf")
			writeReloadFile(t, leaf, reloadRule("/before", 70001))
			directives := "SecRuleEngine On\n"
			var includes []string
			change := func() { writeReloadFile(t, leaf, reloadRule("/after", 70002)) }
			switch kind {
			case "include":
				directives += fmt.Sprintf("Include %q\n", leaf)
			case "deprecated-include":
				includes = []string{leaf}
			case "nested-include":
				parent := filepath.Join(dir, "parent.conf")
				writeReloadFile(t, parent, fmt.Sprintf("Include %q\n", leaf))
				directives += fmt.Sprintf("Include %q\n", parent)
			case "glob-membership":
				directives += fmt.Sprintf("Include %q\n", filepath.Join(dir, "*.conf"))
				change = func() {
					require.NoError(t, os.Remove(leaf))
					writeReloadFile(t, filepath.Join(dir, "replacement.conf"), reloadRule("/after", 70002))
				}
			case "phrase-file":
				writeReloadFile(t, leaf, "/before\n")
				directives += fmt.Sprintf("SecRule REQUEST_URI \"@pmFromFile %s\" \"id:70001,phase:1,deny,status:403,nolog\"\n", leaf)
				change = func() { writeReloadFile(t, leaf, "/after\n") }
			}
			before := provisionReloadModule(t, directives, includes)
			require.NotNil(t, requestInterruption(t, before, "/before"))
			require.Nil(t, requestInterruption(t, before, "/after"))
			change()
			after := provisionReloadModule(t, directives, includes)
			require.NotNil(t, requestInterruption(t, after, "/after"), "reload must observe changed external bytes")
			require.Nil(t, requestInterruption(t, after, "/before"), "new policy must replace old matches")
			require.NotNil(t, requestInterruption(t, before, "/before"), "old generation remains valid until Caddy retires it")
		})
	}
}

func TestReloadRejectsBrokenExternalPolicy(t *testing.T) {
	for _, kind := range []string{"missing", "invalid"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "rules.conf")
			writeReloadFile(t, path, reloadRule("/before", 70001))
			directives := fmt.Sprintf("SecRuleEngine On\nInclude %q\n", path)
			active := provisionReloadModule(t, directives, nil)
			if kind == "missing" {
				require.NoError(t, os.Remove(path))
			} else {
				writeReloadFile(t, path, "InvalidDirective On\n")
			}
			candidate := &corazaModule{Directives: directives}
			require.Error(t, candidate.Provision(reloadContext(t)), "broken replacement must be compiled and rejected")
			require.NoError(t, candidate.Cleanup(), "failed provision must be safe to clean up")
			require.NotNil(t, requestInterruption(t, active, "/before"), "failed candidate must preserve active generation")
		})
	}
}
