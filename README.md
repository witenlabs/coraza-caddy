# Witen downstream Coraza Caddy connector

This is Witen Labs' focused fork of
[corazawaf/coraza-caddy](https://github.com/corazawaf/coraza-caddy), licensed
under Apache-2.0. See [FORK.md](FORK.md) for the patch scope, tested build,
module replacement instructions and remaining qualification work. The original
module path and Caddy configuration names remain compatible with upstream.
This initial downstream candidate is experimental and has not been deployed.

<!-- Modified by Witen Labs in 2026: identify the downstream project, build
instructions and qualified scope. Upstream attribution and license retained. -->

## Coraza WAF Caddy Module

[![Witen CI](https://github.com/witenlabs/coraza-caddy/actions/workflows/witen.yml/badge.svg)](https://github.com/witenlabs/coraza-caddy/actions/workflows/witen.yml)
<a href="https://pkg.go.dev/github.com/corazawaf/coraza-caddy" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>

[OWASP Coraza](https://github.com/corazawaf/coraza) Caddy Module provides Web Application Firewall capabilities for Caddy.

Coraza implements ModSecurity-style SecLang and integrates with OWASP CRS.
Qualify the selected engine, connector and rule versions for your workload.
## CRS Documentation

If you’re looking for installation guidance, tuning, false-positive handling, or
deployment best practices for CRS, refer to the official OWASP CRS documentation.
https://coreruleset.org/docs/
This repository focuses on Coraza integration and runtime behavior rather than
maintaining CRS rule documentation.

## Getting started

`go run mage.go -l` lists all the available commands:

```bash
go run mage.go -l
Targets:
  buildCaddy        builds the plugin.
  buildCaddyLinux   builds the plugin with GOOS=linux.
  buildExample       builds the example deployment.
  check              runs lint and tests.
  coverage           runs tests with coverage and race detector enabled.
  doc                runs godoc, access at http://localhost:6060
  e2e                runs e2e tests with a built plugin against the example deployment.
  format             formats code in this repository.
  ftw                runs CRS regressions tests.
  lint               verifies code quality.
  precommit          installs a git hook to run check when committing
  reloadExample      reload the test environment.
  runExample         spins up the test environment, access at http://localhost:8080.
  teardownExample    tears down the test environment.
  test               runs all tests.
```

## Plugin syntax

```caddy
coraza_waf {
 directives `
  Include /path/to/config.conf
  SecAction "id:1,pass,log"
 `
}
```

Sample usage:  

**Important:** `order coraza_waf first` must be always included in your Caddyfile for Coraza module to work

```caddy
{
    order coraza_waf first
}

http://127.0.0.1:8080 {
 coraza_waf {
  directives `
   SecAction "id:1,pass,log"
   SecRule REQUEST_URI "/test5" "id:2, deny, log, phase:1"
   SecRule REQUEST_URI "/test6" "id:4, deny, log, phase:3"
   Include file1.conf 
   Include file2.conf
   Include /some/path/*.conf
  `
 }
 reverse_proxy http://192.168.1.15:8080
}
```

## Build Caddy with Coraza WAF

Run:

```shell
go build -tags=coraza.no_memoize -buildvcs=false -trimpath -o build/caddy ./caddy
```

## Testing

You may run the test suite by executing:

```shell
GOFLAGS=-tags=coraza.no_memoize go run mage.go test
```

## Using OWASP Core Ruleset

You can load OWASP CRS by passing the field `load_owasp_crs` and then load the CRS files in the directives as described in the [coraza-coreruleset](https://github.com/corazawaf/coraza-coreruleset) documentation.

The `@`-prefixed paths are **not files on disk**. They are served from a ruleset
that is compiled into the binary, and `load_owasp_crs` is what makes them
resolvable. This has two consequences:

- You do not need to download CRS, mount a volume, or ship rule files alongside
  the binary. `xcaddy build --with github.com/corazawaf/coraza-caddy/v2` already
  contains them.
- Do not put a filesystem path in front of an `@` path.
  `Include /etc/caddy/rules/@owasp_crs/*.conf` will not work; the `@` prefix
  must start the path.

If you use an `@` path without `load_owasp_crs`, Caddy fails at startup with:

```text
invalid WAF config from string: failed to readfile:
  open @coraza.conf-recommended: no such file or directory
```

Adding `load_owasp_crs` to the `coraza_waf` block resolves it.

```caddy
:8080 {
 coraza_waf {
  load_owasp_crs
  directives `
   Include @coraza.conf-recommended
   Include @crs-setup.conf.example
   Include @owasp_crs/*.conf
   SecRuleEngine On
  `
 }

 reverse_proxy httpbin:8081
}
```

## Using Coraza plugins

Coraza plugins (extra operators, body processors, GeoIP lookups) are plain Go
packages that register with Coraza from an `init()`. xcaddy adds a blank import
for every `--with` package, so a plugin that registers itself and needs no
configuration can be added straight to the build:

```shell
GOFLAGS=-tags=coraza.no_memoize xcaddy build --with github.com/corazawaf/coraza-caddy/v2=. \
  --with github.com/some/coraza-plugin
```

A plugin that needs setup, such as
[coraza-geoip](https://github.com/corazawaf/coraza-geoip) which must be given a
database, needs a small package of your own that does the registration. Pass it
to xcaddy as a local module:

```go
// register/register.go
package register

import (
	_ "embed"
	"fmt"

	geo "github.com/corazawaf/coraza-geoip"
)

//go:embed GeoLite2-Country.mmdb
var db []byte

func init() {
	if err := geo.RegisterGeoDatabase(db, "country"); err != nil {
		panic(fmt.Errorf("registering GeoIP database: %w", err))
	}
}
```

```shell
# register/go.mod declares `module example.com/register` and requires coraza-geoip
xcaddy build --with github.com/corazawaf/coraza-caddy/v2 \
  --with example.com/register=./register
```

## Specifying transaction ID from downstream

You can specify the transaction ID by setting the field `tx_id_req_header` and then pass the value for that request header.
This is useful when running coraza behind another HTTP server and using the request ID issued by that server.

**Important:** `tx_id_req_header` must be set or overwritten only by a trusted upstream proxy, and that proxy must strip any client-provided value for the same header before it reaches coraza. Otherwise, clients can spoof the transaction ID (and any downstream consumers of it). Configure your front-end proxy to remove client-sent values and inject a trusted request ID instead.

```caddy
:8080 {
 coraza_waf {
  load_owasp_crs
  directives `
   Include @coraza.conf-recommended
   Include @crs-setup.conf.example
   Include @owasp_crs/*.conf
   SecRuleEngine On
  `
  tx_id_req_header "X-Transaction-ID"
 }

 reverse_proxy httpbin:8081
}
```


## Performance

Upstream Coraza enables compiled matcher memoization by default. This initial
Witen candidate requires `coraza.no_memoize` because Coraza 3.8.1 can otherwise
retain stale file-backed dictionaries across reloads. See [FORK.md](FORK.md).

| Build tag | Behaviour |
|---|---|
| *(none)* | Connector refuses activation with a build instruction |
| `coraza.no_memoize` | Qualified candidate mode; every pattern compiles fresh |

Older guidance recommends building with `-tags memoize_builders`. That tag was
how memoization was opted into up to Coraza v3.4.0, when the default was off. It
was removed in v3.5.0 once memoization became the default. Go ignores unknown
build tags silently, so passing it today neither fails nor does anything — drop
it from your build.

Every configuration now owns its WAF and calls Close when retired. Repeated
rule compilation and separate instances can increase reload time and peak
memory. Measure the intended site/rule count and sustained churn; a short
successful reload does not establish a memory bound for production.


## Running Example

### Docker

```bash
GOFLAGS=-tags=coraza.no_memoize go run mage.go buildExample runExample
curl -i localhost:8080/
```

### Local

```bash
# in terminal 1
go run github.com/mccutchen/go-httpbin/v2/cmd/go-httpbin@v2.9.0 -port 8081

# in terminal 2
GOFLAGS=-tags=coraza.no_memoize go run mage.go buildCaddy
./build/caddy run --config example/Caddyfile --adapter caddyfile

# in terminal 3
curl -i localhost:8080/
```

## Respond with custom message or HTML page

In order to respond with a custom message or HTML page, you can take advantage of [handle_errors](https://caddyserver.com/docs/caddyfile/directives/handle_errors) directive:

```caddy
handle_errors 403 {
 header X-Blocked "true"
 respond "Your request was blocked. Request ID: {http.request.header.x-request-id}"
}
```
or
```caddy
handle_errors {
 @block_codes `{err.status_code} in [403]`
 handle @block_codes {
  root    * /path/to/html/dir
  rewrite * /{err.status_code}.html
  file_server
 }
}
```

It is possible to use the [templates](https://caddyserver.com/docs/caddyfile/directives/templates) directive to render data dynamically. Take a look at [`example/403.html`](./example/403.html) file.  

When the WAF blocks a request, `{http.error.message}` is
`interruption triggered by rule <id> (<action>)`. Use it to tell WAF blocks
apart from application errors with the same status code, or to show the rule
ID on the error page:

```caddy
handle_errors {
 @waf expression `{http.error.message}.startsWith("interruption triggered")`
 respond @waf "Blocked by the WAF: {http.error.message}" 403
}
```

The `WAF rule violation detected` log entry carries the same information in
the `rule_id` and `action` fields, for blocks raised in the request phases and
in the response phases alike.
