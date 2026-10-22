# Host-trial preparation

This is an experimental candidate, not an unattended production deployment.
Trial one host first and widen only after a successful bounded trial. Every host
configuration change is explicit and opt-in. Building the module
does not change PHP-FPM, SELinux, logging, rule mode or installed services.

## Qualified components and checks

Use Go 1.26.10, the exact connector/core pins in go.mod, and
`GOFLAGS=-tags=coraza.no_memoize`. Consumers must set both replacements explicitly.
The Witen observation formatter remains outside the connector/core repositories.
Do not substitute native/JSON raw audit output for that reviewed formatter.

The Linux descriptor regression exercises 64 generations, pending transactions,
rename rotation and failed candidates. Core tests cover write/close failures and
failed construction. The real Caddy fixture lab also verifies body preservation,
inspection limits and privacy. A bounded 60-second HTTP soak varies 1-24 profiles,
forces repeated reloads, rejects invalid candidates, rotates the audit file and
requires exact canary delivery and one remaining audit descriptor after draining.
It checks settled live heap against a fixed budget. This is not an hours-long soak
or a full production capacity result.

The default and uncached core race suites and CRS tag matrix qualify engine
behavior. Full connector proxy/protocol coverage, a coordinated Caddy/CEL update,
multi-hour churn, production Warden ingestion and independent delivery alerting
remain gates before an unattended live trial.

## Required configuration choices

- Keep `SecRuleEngine DetectionOnly`, `SecResponseBodyAccess Off`, bounded request
  inspection/RAM limits and ProcessPartial. Preserve the upstream application's
  existing request-size limits; incomplete inspection is an observation.
- Keep rules reviewed, pinned and `nolog,auditlog`. Use the external Witen formatter
  with structured parts AFHKZ and validated site/policy identity. Emit no raw URI,
  headers, body, match data or credentials in findings. Do not enable debug/trace
  or raw rule callbacks to diagnose delivery.
- Use `SecDebugLogLevel 1` for operational errors and verify that the runtime sink
  receives the known audit-write failure signal. The lab found that level 0
  suppresses that signal. Review runtime error output for the actual policy too;
  arbitrary third-party rule/plugin errors are outside this privacy qualification.
- Give the audit file a private directory and mode 0600 with the service owner.
  Any Warden read access, host logging or SELinux change needs its own reviewed
  configuration. Verify delivery through the actual reader, not only file access.
- Explicitly configure each reverse proxy's HTTP transport:

  ```caddyfile
  reverse_proxy 127.0.0.1:8080 {
      transport http
  }
  ```

  In JSON use `"transport": {"protocol": "http"}`. The 2.11.4 lab's implicit
  default transport retained idle connections across reloads. Stacks showed
  net/http persistConn readers/writers and backend idle connections accumulating.
  Explicit module loading lets Caddy call HTTPTransport.Cleanup. The identical
  soak passes with this configuration and unchanged memory assertions. This is a
  configuration workaround, not a Coraza cache fix or proof of unbounded leakage.
- Rotate by rename/create with the same owner/mode, then force reload. The new
  generation opens the new file and old requests finish in the renamed file.
  Keep old files until readers drain. Do not unlink active logs or rely on
  copytruncate for exact delivery. A failed candidate must leave the old policy
  and old sink intact; repair the target and retry with a delivery canary.
- Trust only the direct peer initially. Check that spoofed forwarding headers
  cannot select a client IP. Public proxy/CDN trust requires a separate explicit
  configuration and tests against the real ingress topology.

## Trial sequence and rollback boundary

Inventory the target host read-only first: the installed Caddy version and
modules (preserve any DNS-provider or other non-default modules), the service
manager and account, and any file capabilities on the binary. A live replacement
must retain those exactly; the unprivileged high-port smoke deliberately needs none.

First run the exact binary unprivileged on a loopback high port with admin/config
persistence disabled, private scratch storage, a synthetic canary and no ACME.
Check module registration, detection-only delivery, spoofed-header behavior,
privacy, invalid reload rollback and audit rotation. Stop it and remove its scratch
data after preserving evidence. This does not qualify production TLS or replace
the running service.

Before any live replacement, preserve the exact binary, Caddyfile/includes,
service unit/config, file capabilities, ownership/modes and package identity.
Keep TLS storage private and unchanged; do not export it with test evidence.
Validate the candidate with the existing configuration and service identity.
Select one site, add only reviewed detection policy/logging configuration, and
prepare an exact rollback command using that host's verified service layout.
Do not assume copying a binary preserves its bind capabilities or restarting a
shell-owned process establishes service-manager ownership. Verify ordinary sites, TLS,
proxy streaming, Warden delivery and resource baselines before widening scope.
Rollback restores the saved binary/config/metadata and restarts the original
service; verify both normal traffic and monitoring afterward.

Containerized Caddy needs a digest-pinned custom image, preserved modules,
mounted config/log/storage compatibility, private network route checks and an
exact image/config rollback. Treat it as an infrastructure change, not part of an
application rollout.
