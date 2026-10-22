# Witen downstream scope

This fork starts from upstream main
`e268cf0f44e95d64da81c641aabebd21364be6f1` (2026-10-15).
The upstream source/history, Apache-2.0 license and attribution are preserved.
New downstream commits are authored by Witen Labs. Maintaining this candidate
does not establish an upstream maintainer role or production support promise.

## Initial patch

Each connector configuration owns a fresh Coraza WAF. Provision recompiles
the effective policy, and Cleanup closes the retired instance once and
releases its reference, including on a close error. The connector no longer
shares whole WAF instances across routes or configuration generations.

Previously, the process-global pool keyed directives and Include path strings.
An external file could change without changing that key, leaving stale policy
active even after a forced reload. Rebuilding handles direct, deprecated and
nested includes, glob membership and file-backed operators without attempting
an incomplete dependency scanner. It also gives each generation its current
logger. Missing/invalid replacements fail provisioning and leave the active
configuration available through Caddy's existing reload transaction.

Further regression testing found a second issue in Coraza 3.8.1: `pmFromFile`
and `pmf` matcher memoization uses file paths rather than contents. A fresh WAF
can therefore inherit a stale dictionary while the previous generation is
active. The initial qualified build requires `coraza.no_memoize`; without the
tag, Provision refuses activation with a clear error. This uses the engine's
supported build option and keeps a core fork out of this connector patch.
The dictionary regression retains its expected new match; it is not skipped.

The tradeoff is recompilation on every provision and separate WAF instances
for otherwise identical route configurations, plus disabled matcher memoization.
Large deployments must measure
reload duration, peak memory and sustained churn before adopting. This is a
deliberate correctness tradeoff; there is no opt-in switch that restores stale
cross-generation caching.

External file edits alone do not notify Caddy. Use an explicit forced reload,
for example `caddy reload --config Caddyfile --adapter caddyfile --force`.
Keep the full external policy set stable during validation and reload; this
fix does not make simultaneous edits to multiple files an atomic operation.

## Build and configuration compatibility

The module remains `github.com/corazawaf/coraza-caddy/v2`. The module ID remains
`http.handlers.waf` and the Caddyfile directive remains `coraza_waf`. Compile
one implementation into a Caddy binary. Warden/catalog adapters stay outside
this repository.

The initial candidate uses Go 1.26.10, Caddy 2.11.4 and Coraza 3.8.1, with
`golang.org/x/net` 0.60.0 and its required companion dependency versions.
The upstream default branch also uses the split embedded Coraza configuration
and CRS modules; preserve their pinned versions rather than floating rules.

Host-trial qualification adds a separately pinned [Witen core companion](https://github.com/witenlabs/coraza/tree/witen-v3.8.1)
at `7fda999742423c2cc64092103bdd714bf183d3ca`, based on Coraza 3.8.1.
It reports serial audit write errors and closes initialized writers after active
transactions drain, including failed-construction cleanup. Module consumers must
copy the core replacement from this go.mod: dependency replacements are not
inherited transitively. Its generated Go pseudo-version uses the fork's module
history; the reviewed upstream source base remains 3.8.1.

The official scanner reports no reachable vulnerability in this build. It
also reports GO-2026-6094 in CEL 0.28.1's unused NativeTypes/ParseStructTag paths
and GO-2026-5932 in the unimported x/crypto/openpgp package. CEL 0.30.0 fixes
the former but its changed interpreter API fails to compile with Caddy 2.11.4;
retain the compatible pin and track a coordinated Caddy/CEL update. No stable
security qualification is claimed for this experimental candidate.

From this checkout:

```sh
export GOFLAGS=-tags=coraza.no_memoize
go test -race -count=1 ./...
go vet ./...
go run mage.go lint
go build -buildvcs=false -trimpath -o build/caddy ./caddy
./build/caddy list-modules
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -test ./...
```

For another Go project, use a local replacement during development:

```go
replace github.com/corazawaf/coraza-caddy/v2 => ../coraza-caddy

replace github.com/corazawaf/coraza/v3 => github.com/witenlabs/coraza/v3 v3.0.0-20261022033216-7fda99974242
```

For xcaddy, pin the downstream commit, substituting its full SHA:

```sh
GOFLAGS=-tags=coraza.no_memoize xcaddy build v2.11.4 \
  --with github.com/corazawaf/coraza-caddy/v2=github.com/witenlabs/coraza-caddy/v2@DOWNSTREAM_COMMIT \
  --with github.com/corazawaf/coraza/v3=github.com/witenlabs/coraza/v3@7fda999742423c2cc64092103bdd714bf183d3ca
```

There is no stable downstream tag yet. Upstream version tags identify upstream
history and must not be mistaken for a tested Witen release. Distribution must
retain LICENSE and applicable third-party notices and identify changed files.
Dependencies and optional rulesets have their own license obligations.

## Qualification and maintenance boundaries

New regressions cover changed direct/nested/deprecated includes, glob membership,
phrase files, missing/invalid replacement rejection, independence of active
and candidate generations, and cleanup/idempotency/error propagation. The
Witen HTTP fixture lab supplies actual Caddy reload, request-body preservation,
privacy, resource and profile checks in addition to these connector tests.

DetectionOnly does not automatically make runtime/audit logging private.
Reviewed observation rules can use `nolog,auditlog` with a qualified redacting
formatter. Do not enable raw access logs, match data or unreviewed rules through
a metadata signature update. Keep source collection and enforcement separate.

The companion core fixes audit writer errors/ownership without introducing a
second core implementation here. Operational error output requires
`SecDebugLogLevel 1`; level 0 suppresses it. Reviewed rules still use `nolog,auditlog`
and the external redacting formatter. Disk-write errors do not prove delivery or
change HTTP responses. Independent canaries and Warden ingestion checks remain
required. See [HOST-TRIAL.md](HOST-TRIAL.md) for rotation and the explicit proxy
transport workaround qualified by the sustained lab.

Representative long-running tenant churn, HTTP/2/3, proxy address trust
and streaming still need qualification before a live host trial. No production host,
PHP-FPM, SELinux or logging configuration is changed by building this fork.

The Witen CI workflow is read-only and runs tests, vet, the existing lint and
dependency consistency gate, a custom Caddy build and the official vulnerability
checker. Upstream-specific automation remains in the
tree but is restricted to the upstream repository; no automatic downstream tag,
release, stale-issue message or upstream contact is configured. Keep patches
separable so individual fixes can later be proposed upstream and retired here.
