# How detection works

SecureVibe detects threats in two fundamentally different ways, and it is worth
being precise about which is which — because only one of them can catch
something *novel*.

| Layer | Mechanism | Catches | Database-bound? |
|-------|-----------|---------|-----------------|
| **Deterministic scanners** | exact lookups, regex, edit-distance, version-range eval | **known** threats (and close misspellings of known-popular names) | yes |
| **Generation-time skills** | an AI assistant reasons over your code using `SKILL.md` rules | **novel** code flaws with no CVE (SSRF, broken authz, …) | no — pattern reasoning |

Everything on this page about typosquats, malicious packages, secrets, and CVEs
is the **deterministic** layer: fast, cheap, no false drama, but bounded by its
data. The skill layer is covered in [What makes SecureVibe different](features.md);
it is the only layer that is not database-bound.

---

## Typosquat detection

A **typosquat** is a malicious package deliberately named one or two letters off
a popular one — `request` instead of `requests`, `crossenv` instead of
`cross-env` — hoping you fat-finger the name and install the attacker's package.

SecureVibe checks every dependency name two independent ways, kept deliberately
separate by confidence level.

```mermaid
flowchart TD
    P["package name<br/>(e.g. request)"] --> C1
    P --> C2
    subgraph C1["1 · Curated DB — known squats"]
        D1["exact match against<br/>known_typosquats.json"]
    end
    subgraph C2["2 · Runtime distance — novel squats"]
        D2["edit-distance ≤ 2 from a<br/>name on the popular list"]
    end
    D1 -->|match| H["HIGH confidence<br/>→ finding → gate fails"]
    D2 -->|0 < distance ≤ 2| L["LOW confidence<br/>→ advisory hint"]
    L -.->|review / promote| D1
    style H fill:#fdd,stroke:#c00
    style L fill:#ffd,stroke:#aa0
```

### 1. Curated DB — the "known-bad list"

`vulnerabilities/supply-chain/typosquat-db/known_typosquats.json` is a
human-reviewed list. Each row maps a known squat to the legitimate target it
imitates:

```json
{ "target": "requests", "typosquat": "request", "ecosystem": "pypi",
  "levenshtein_distance": 1, "status": "removed", "discovered": "2017-09-12",
  "references": ["https://…advisory…"] }
```

Detection is a plain **case-insensitive exact match** of the package name
against the `target` or `typosquat` field. Because a human vetted the row, the
match is structural — **`confidence: high`**. In a lockfile scan only the *squat*
side is reported as a finding: depending on the real `requests` is not a
vulnerability.

> **Honest limit.** This catches only squats someone has already filed. The
> stored `levenshtein_distance` is display metadata — the match itself is string
> equality, not a computation.

### 2. Runtime edit-distance — the "looks like a popular name" net

This is the part that catches squats **nobody has filed yet**. For each
dependency, SecureVibe computes the
[Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance)
(the number of single-character inserts/deletes/substitutions) between the name
and every entry on the per-ecosystem **popular-packages** list
(`vulnerabilities/supply-chain/popular-packages/<eco>.json`).

```
   requests   ← real, popular package (on the popular list)
   request    ← what you typed
   ^^^^^^^
   1 letter missing  →  edit distance = 1  →  suspicious (0 < d ≤ 2)
```

A name within distance 1 or 2 of a popular package — but not equal to it — is
surfaced as a **`confidence: low`** suggestion. Three guards keep it quiet:

- **Explicit ecosystem required** — otherwise an npm name that resembles a PyPI
  one produces cross-language noise.
- **Name normalization** (`typosquatCompareKey`) — lower-cased; for **Go**,
  stripped to the final import-path segment (`bolt` in
  `github.com/boltdb/bolt`) so legitimate forks under other owners don't trip it.
- **Self-popular skip** — if the name is *itself* on the popular list, the sweep
  is skipped. Popular names sit within distance 2 of each other (`chalk`,
  `react`, `lodash`), so without this every popular dependency would false-alarm.

> **Honest limit.** Bounded three ways: the target must be on the popular list,
> the distance must be ≤ 2, and only *names* are compared — there is no
> behavioural analysis.

### The two mechanisms are complementary, not redundant

Three real entries from the shipped database show why both are needed:

| Input | Curated DB | Runtime distance | Caught by |
|-------|-----------|------------------|-----------|
| `crossenv` → `cross-env` (npm) | hit, d=1 | also d=1 | **both** |
| `crossen` (novel, not yet filed) | miss | d=2 from `cross-env` | **runtime net only** |
| `boltdb-go/bolt` → `boltdb/bolt` (Go) | hit, d=3 (full path) | segment `bolt`==`bolt` → d=0, invisible | **curated DB only** |

The last row is the key insight: the attack lives in the **owner** segment
(`boltdb` → `boltdb-go`), which the Go normalization discards to avoid
flagging forks. The runtime net is blind to it; only the curated DB — fed by
human review and upstream feeds — covers that class. Conversely, `crossen` has
never been filed, so only the runtime net sees it.

---

## The other deterministic detectors

Typosquats are one of several deterministic passes the scanner and the CI
[`gate`](../reference/cli.md) run. They all share the same four confidence
bands, set by *how the match was made* — not how bad the issue is:

| Band | Meaning |
|------|---------|
| `confirmed` | hand-reviewed curated evidence, **or** an OSV record whose version range covers the resolved version |
| `high` | unambiguous structured hit not individually triaged (OSSF feed, curated typosquat, a hardening regex) |
| `medium` | pattern-only signal (substring CVE-name match, runtime typosquat suggestion) |
| `low` | weakest tier — fuzzy heuristics |

Overview, then the algorithm for each:

| Detector | How it matches | Confidence |
|----------|----------------|-----------|
| **Malicious packages** | exact name (+ version range) against `malicious-packages/<eco>.json` (OpenSSF + curated + local overlay) | `confirmed` / `high` |
| **OSV advisories** | resolved version evaluated against each advisory's `affected[].ranges` event stream | `confirmed` (in range) / `high` (version unconfirmed) |
| **CVE patterns** | substring match on curated CVE name/description, filtered by ecosystem + version | `medium` |
| **Secrets** | regex token shape → Shannon entropy gate → hotword proximity → denylist | per-rule |
| **Dockerfile / GitHub Actions** | regex pass + structured (AST) pass for hardening anti-patterns | `high` / per-rule |

The whole dependency pipeline runs every resolved `(name, version, ecosystem)`
tuple from a lockfile through four checks at once:

```mermaid
flowchart LR
    L["lockfile<br/>(package-lock.json, go.sum, …)"] --> P["parse → (name, version, ecosystem)"]
    P --> M["malicious-packages DB"]
    P --> T["typosquat DB + popular list"]
    P --> O["OSV advisories"]
    P --> C["CVE patterns"]
    M --> F["findings → SARIF / gate exit code"]
    T --> F
    O --> F
    C --> F
```

### Malicious packages

A curated + imported canon of packages known to be malicious, one file per
ecosystem (`vulnerabilities/supply-chain/malicious-packages/<eco>.json`). The
match is **exact name**, then an optional **version-range** filter:

```
for each entry in malicious-packages/<eco>.json:
    if not EqualFold(entry.name, dep.name):  skip
    if dep.version set and entry.versions_affected set
       and version NOT in any range:         skip          # already on a safe version
    → finding
```

The version grammar is permissive (`all` / `*`, `pre-X.Y.Z`, `>=`, `<=`,
ranges `A - B`, exact) and falls back to the ecosystem semver matcher. The
**local contribution overlay** (`.skills-check/overlay.json`, the LEARN loop) is
consulted here too, so a package you flagged with `contribute add` blocks
exactly like a curated row.

Confidence depends on provenance:

```
curated row (no upstream source) ......... confirmed   ← hand-reviewed
imported source: ossf-malicious-packages . high        ← structured, not triaged
local overlay (user-asserted) ............ high
```

### OSV advisories

The OSV layer answers "is this exact version inside a published advisory's
affected range?" — versions, not just names. Data comes from a local cache
(`vulnerabilities/osv/<eco>/` or the larger `osv-cache.tar.gz` / `fetch-vulns`
download), optionally `api.osv.dev` directly.

```mermaid
flowchart TD
    A["(name, version, ecosystem)"] --> B["index.json lookup by lowercase name"]
    B -->|no entry| Z["no advisory"]
    B -->|candidates| C["for each advisory: evaluate affected[]"]
    C --> D{"version vs<br/>versions[] + ranges[]"}
    D -->|in an evaluable range| IN["IN RANGE → keep,<br/>VersionConfirmed → confirmed"]
    D -->|all evaluable ranges exclude it| OUT["NOT AFFECTED → drop"]
    D -->|nothing evaluable / no version| UNK["UNKNOWN → keep,<br/>fail-open → high"]
```

Each `affected[].ranges[].events[]` list is an **ordered state machine** over
`introduced` / `fixed` / `last_affected` bounds:

```
affected = false
for each event in order:
    introduced=0   → affected = true            # from the beginning
    introduced=X   → if version >= X: affected = true
    fixed=X        → if version >= X: affected = false   # exclusive upper bound
    last_affected=X→ if version  > X: affected = false   # inclusive upper bound
# unparseable bound on either side → fail open (kept as "unknown")
```

Range types `ECOSYSTEM` and `SEMVER` are evaluated; `GIT` ranges are skipped.
Version comparison is **ecosystem-native** — npm (`^`, `~`, X-ranges, pre-release
rules), PyPI (PEP 440 `~=`, `.postN`, `.devN`), Go (incl. pseudo-versions) — with
a generic dotted-numeric fallback for crates/maven/nuget/rubygems/etc. Severity
is read from `database_specific.severity` (GitHub band) first, else parsed from a
CVSS v3/v2 vector and bucketed (`≥9.0` critical, `≥7.0` high, `≥4.0` medium).

> **Why two bands.** If the version intersects an evaluable range →
> `confirmed`. If the version is absent or the range grammar can't be evaluated,
> the advisory is kept on the *name* match alone (fail-open) → `high`. A version
> proven *outside* every range is dropped entirely.

### CVE patterns

A lighter, name-oriented pass over a curated CVE list. It is a **substring**
match, deliberately fenced so it stays suggestive rather than noisy:

```
needle = lower(package name)
for each cve entry:
    if needle NOT in lower(entry.name + " " + entry.description):  skip
    if entry.languages don't overlap this ecosystem:               skip   # "express" in "OGNL expressions"
    if version known and entry pins affected versions
       and version not affected:                                   skip
    → finding (medium)
```

Because the underlying patterns describe code shapes, not pinned versions, every
hit is **`medium`** — a strict `gate --severity-floor high` won't fail on it.

### Secrets

Secret detection is regex-first, then three filters that cut the false positives
regex alone produces. Each rule
(`skills/secret-detection/checklists/secret_detection.yaml`) carries a `pattern`,
`severity`, `entropy_min`, `hotwords` / `hotword_window` / `hotword_boost` /
`require_hotword`, and `denylist_substrings`.

```mermaid
flowchart TD
    R["regex match (e.g. AKIA[0-9A-Z]{16})"] --> D{"denylisted substring?<br/>(EXAMPLE, AKIAIOSFODNN7)"}
    D -->|yes| X1["drop"]
    D -->|no| E{"Shannon entropy ≥ entropy_min?"}
    E -->|no| X2["drop (low-randomness → not a real key)"]
    E -->|yes| H{"hotword within ±window?<br/>(aws, secret, credentials…)"}
    H -->|require_hotword and none| X3["drop"]
    H -->|hit| S["score = score_weight + hotword_boost"]
    H -->|miss, not required| S2["score = score_weight"]
    S --> F["finding (+ known-false-positive flag)"]
    S2 --> F
```

Shannon entropy is computed over the matched bytes,
`H = −Σ p·log₂(p)` (0–8 bits; base64-style secrets land ~4–6, prose/repeats
below ~4). The hotword window scans `text[start−window : end+window]` for context
words. A real rule:

```yaml
- id: aws-access-key
  severity: critical
  pattern: AKIA[0-9A-Z]{16}
  score_weight: 1
  hotwords: [aws, access_key, credentials, iam, secret]
  hotword_window: 200
  hotword_boost: 2
  entropy_min: 3.5
  denylist_substrings: [EXAMPLE, AKIAIOSFODNN7]   # the canonical AWS sample key
```

So `AKIAIOSFODNN7EXAMPLE` (AWS's own docs key) is dropped by the denylist, while a
high-entropy `AKIA…` sitting next to the word `credentials` scores `1 + 2 = 3`.

### Dockerfile / GitHub Actions

Two hardening scanners, each a **regex pass plus a structured (AST) pass**. The
regex pass catches per-line anti-patterns; the AST pass is the *canonical*
detector for rules that need structure, so the two are split to avoid
double-firing.

```
Dockerfile  (regex, "high"):  dkr-explicit-latest-tag, dkr-no-secrets-in-env,
                              dkr-no-add-remote, dkr-no-curl-pipe-sh, dkr-apt-version-pin …
            (AST, multi-stage aware):  dkr-pinned-base-digest, dkr-non-root-user
                              (e.g. final stage with NO USER → implicit root,
                               which the regex "USER root" check can't see)

GitHub Actions (regex):  gha-default-permissions-read, gha-oidc-cloud-credentials,
                         gha-no-curl-pipe-bash, gha-no-untrusted-script-injection …
               (AST, canonical):  gha-pin-actions-by-sha (tag vs 40-hex SHA),
                         gha-pr-target-no-untrusted-checkout (pull_request_target
                         + checkout of the untrusted head ref = "pwn request")
```

Expression-injection deliberately stays a regex check — the structured walk
missed cases the regex catches.

---

## What none of them can do

Every detector above matches a **signature, name, version, or shape**. None
trace **data-flow** — a server-side request forgery, a broken authorization
check, or an injection that crosses function boundaries has nothing to look up.
Those are the job of the generation-time **skills**, which encode a
*generalizable* pattern ("never fetch a client-supplied URL without an
allowlist") so the assistant recognizes an instance that has no CVE. That is the
only non-database-bound layer — see [What makes SecureVibe different](features.md).

---

## Expanding the databases

Detection coverage grows three ways. After editing any data file, run
`skills-check validate` and `skills-check manifest compute --path . --write`;
releases are then Ed25519-signed out-of-band (see
[SIGNING.md](https://github.com/shieldnet-360/securevibe/blob/main/SIGNING.md)).

### Automated bulk ingestion (maintainer refresh)

```bash
python3 scripts/ingest-malicious-packages.py     # refresh malicious-packages DB (OpenSSF)
python3 scripts/derive-typosquats-from-ossf.py    # derive typosquat rows by name-similarity
python3 scripts/ingest-osv.py                      # refresh OSV cache  (or: skills-check fetch-vulns)
```

`derive-typosquats-from-ossf.py` measures every malicious name against the
popular list and keeps a pair only when **distance ∈ {1, 2}** *and* the length
difference is ≤ 2. Each derived row carries `source: ossf-malicious-packages-derived`
plus the upstream `osv_id`; hand-curated rows are preserved untouched.

### Manual curation (a PR)

Add a row to `known_typosquats.json` or `malicious-packages/<eco>.json`. Every
entry **requires at least one external reference** — CI rejects anonymous
"trust me" entries.

!!! tip "Highest-leverage move: grow the popular-packages list"
    The runtime net's coverage is *popular list × distance ≤ 2*. Adding a
    moderately-popular package to `popular-packages/<eco>.json` instantly makes
    every distance-1/2 misspelling of it detectable — and feeds the
    derive script, which anchors on this list. Widening the popular list widens
    the zero-day net for free.

### Contribute-back (field → upstream)

```bash
skills-check contribute add -p request -e pypi --reason "typosquat of requests"  # instant local block
skills-check contribute keygen
skills-check contribute submit --key ed25519.key                                 # signed candidate → PR
```

A maintainer runs `contribute verify`, promotes the candidate into the curated
DB, and it ships in the next signed release that every `skills-check update`
pulls. See [Contribute a Finding](../contribute.md).

---

## One-line summary

> Typosquat detection is database-bound on purpose: the **curated DB** gives a
> build-failing verdict on known squats, while the **edit-distance net** against
> *popular* names catches novel ones before any human files them. Neither alone
> is enough, so both run — and the only layer that is *not* database-bound is
> the generation-time skill that catches code flaws with no signature at all.
