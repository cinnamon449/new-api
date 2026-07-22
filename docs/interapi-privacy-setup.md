# InterAPI privacy & branding setup

This is the admin configuration checklist that makes the InterAPI storefront's
privacy claims **true**. The storefront copy is already worded as an honest
reframe (see `web/default/src/features/home/components/sections/`), but these
settings must be applied in **System Settings** for the claims to hold.

No backend code changes are required — every item below is a runtime option.

## 1. No email required at registration — true by default

Defaults already allow username-only registration. Verify in **System Settings**:

| Option | Value |
| --- | --- |
| `RegisterEnabled` | `true` |
| `PasswordRegisterEnabled` | `true` |
| `EmailVerificationEnabled` | `false` (default) |
| `TurnstileCheckEnabled` | `false` (see §3 — also a privacy setting) |

Optionally disable OAuth providers that require an email if you offer none.

Source of truth: `controller/user.go` `Register` — email is only validated
inside the `EmailVerificationEnabled` branch.

## 2. Crypto-only payments — Plisio

Plisio is a built-in cryptocurrency gateway (BTC, ETH, USDT, etc.).

1. Set `PlisioApiKey` in **System Settings → Payment**.
2. **One-time:** confirm payment-compliance terms so the gateway activates —
   `POST /api/option/payment_compliance` with
   `ComplianceTermsVersion = "v1"`. Without this, `isPlisioTopUpEnabled()`
   returns false.
3. Leave every other gateway's credentials blank so they don't activate:
   `StripeApiSecret`, `CreemApiKey`, `WaffoEnabled`, `PayAddress`/`EpayId`/
   `EpayKey`, `WaffoPancakeMerchantID`.
4. Clear the Alipay/WeChat placeholders from the `PayMethods` JSON so they
   don't render in the top-up UI.

Source: `controller/topup_plisio.go`, `controller/payment_webhook_availability.go`.

## 3. Providers see your prompt, not your identity — true by default

The relay builds a clean upstream request carrying **only** auth headers — no
`X-Forwarded-For`, no `User-Agent`, no account email
(`relay/channel/api_request.go`). Keep it that way:

- **Do not** configure wildcard/regex header-passthrough rules on any channel
  (`*` or `re:<regex>` in header override) — the skip-list does **not** block
  `X-Forwarded-For` / `X-Real-IP` / `User-Agent`, so a wildcard rule would leak
  them upstream.
- **Do not** template `{client_header:<name>}` IP/UA headers into upstream
  requests.
- Keep `TurnstileCheckEnabled = false` — when on, the client IP is sent to
  `challenges.cloudflare.com` as `remoteip`.

## 4. Minimal logging — honest reframe

new-api logs by default. The claim on the storefront is "minimal logging,"
**not** "zero logs," because some logging cannot be turned off via settings.

**Turn off (System Settings → Operation):**

| Option | Value | Effect |
| --- | --- | --- |
| `LogConsumeEnabled` | `false` | Stops per-request consume logs (model, tokens, quota, timing). **Note:** this degrades the usage dashboards and billing-detail logs. |
| `DataExportEnabled` | `false` | Stops the hourly `quota_data` aggregation used by quota charts. |

**Already private by default:**

- Prompt/message **content is never stored** — logs hold only billing metadata
  (token counts, ratios, timing). Source: `service/log_info_generate.go`.
- Per-request client **IP is not logged** unless a user opts in via `RecordIpLog`
  (defaults `false`). Source: `model/log.go` `RecordConsumeLog`.

**Cannot be disabled via settings (recorded unconditionally):**

- Login logs: client IP + User-Agent (`controller/user.go` `recordLoginAudit`).
- Top-up logs: client IP (`model/log.go` `RecordTopupLog`).
- Admin audit logs: client IP (`middleware/audit.go`).

This is why the public copy says "minimal — only what's required to operate and
bill," and leads with the claims that are fully true (prompts never stored,
request IPs off by default, no identity forwarded to providers). Fully
suppressing the login/top-up/audit IPs would require Go changes; that is the
"backend hardening" path, intentionally out of scope here.

## 5. Branding (InterAPI) — runtime config

The visible brand flows through `/api/status` and overrides the code-level
fallbacks at runtime, so **no protected source identifiers are edited**.

| Setting | Value |
| --- | --- |
| `system_name` | `InterAPI` |
| `logo` | InterAPI logo asset |
| `footer_html` | InterAPI console footer |
| `docs_link` | your docs URL |

Then set the theme preset to **InterAPI** (System Settings → the config drawer,
or per-user) so the console ships branded. The `interapi` preset is a warm
cream-on-near-black palette with a 4px radius, mirroring the storefront; the
console keeps its humanist sans for readability (`theme-presets.css`,
`lib/theme-customization.ts`).

## 6. Storefront notes

- The storefront aesthetic is **scoped** to a `.storefront` wrapper
  (`features/home/storefront.css`) — it never forces monospace onto the console.
- Font: JetBrains Mono (free Berkeley Mono substitute), loaded via
  `@fontsource-variable/jetbrains-mono`. Swap the family in `storefront.css`
  if you license Berkeley Mono.
- The storefront appears when **Home Page Content is empty** (System Settings).
  If you set a custom home-page URL/HTML/markdown, the storefront is bypassed.
