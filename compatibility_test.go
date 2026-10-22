// Copyright 2026 Witen Labs
// SPDX-License-Identifier: Apache-2.0

package coraza

import (
	"net/http"
	"testing"

	"github.com/caddyserver/caddy/v2/caddytest"
	"github.com/stretchr/testify/require"
)

// Exercise Caddy's CEL routing in the qualified dependency combination.
func TestCaddyExpressionCompatibility(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(`{
	admin localhost:2999
	order coraza_waf first
}
http://127.0.0.1:8080 {
	@health expression {http.request.method} == "GET"
	handle @health {
		coraza_waf {
			directives "SecRuleEngine DetectionOnly"
		}
		respond "cel-ok" 200
	}
	respond "fallback" 404
}`, "caddyfile")
	tester.AssertGetResponse(baseURL+"/health", 200, "cel-ok")
	request, err := http.NewRequest("POST", baseURL+"/health", nil)
	require.NoError(t, err)
	tester.AssertResponseCode(request, 404)
}
