# SecureVibe — the 2-minute demo

Three acts. Copy-paste each block. By the end you've *felt* the whole pitch:
your AI assistant writes secure code at generation time, and a deterministic
gate catches anything that slips — **offline, no API key.**

> Run everything from the **repo root**. The deterministic parts (Act 1 & 3)
> need no AI account at all.

---

## Act 1 — The pain (10s)

This folder is what an AI assistant cheerfully produces from innocent prompts
("add an endpoint that fetches a user-supplied URL", "wire up the deps"):

```bash
cat examples/vibe-demo/requirements.txt   # colourama → typosquat of colorama
cat examples/vibe-demo/config.py          # hardcoded GitHub + Stripe tokens
cat examples/vibe-demo/app.py             # SSRF: user controls the fetched URL
```

😬 A typosquatted dependency, two leaked secrets, and an SSRF hole — and it all
*looks* fine at a glance. This is the everyday reality of "vibe coding."

---

## Act 2 — The magic: secure at generation time (30s)

> Needs an AI assistant with the SecureVibe skills loaded (Claude Code shown).

```bash
npx @shieldnet360/secure-code-skill init     # writes 29 skills to ./.claude/skills
```

Now ask your assistant the *same* innocent prompt:

> "Add an endpoint that fetches whatever URL the user passes, and add a colored
> terminal output library to requirements."

With the skills in context, the assistant now:

- validates the URL and **blocks cloud-metadata / private IPs** instead of a raw
  `requests.get(user_url)`,
- reaches for **`colorama`**, and warns that **`colourama` is a known typosquat**,
- loads secrets from the environment instead of hardcoding them.

**Same prompt. Secure output.** That's the part a post-hoc scanner can't do — it
never touches generation.

---

## Act 3 — The proof: a deterministic gate (20s)

"But what if something still slips through?" Run the offline backstop:

```bash
make demo
```

Under the hood that's the same `gate` your CI and pre-commit hook run:

```text
=== gate examples/vibe-demo/requirements.txt ===
Verdict:        FAIL
Scanner used:   scan_dependencies
Findings: 2   (critical: 1, medium: 1)        ← colourama typosquat

=== gate examples/vibe-demo/config.py ===
Verdict:        FAIL
Scanner used:   scan_secrets
Findings: 2   (critical: 2)                    ← GitHub + Stripe tokens

error: gate: 4 finding(s) at or above high     ← exit 1 → CI build fails
```

Prefer the individual tools?

```bash
skills-check check-typosquat -p colourama     # curated: target=colorama squat=colourama
skills-check scan-secrets   examples/vibe-demo # GitHub PAT + Stripe Live key
skills-check scan-dependencies examples/vibe-demo
```

---

## The one number to remember

In SecureVibe's own evals, putting the skills in the model's context drops the
**insecure-output rate from 17.5% → 7.5%** — prevention a scanner structurally
cannot produce, because it only ever runs *after* the code exists.

→ Wire the same gate into [CI and pre-commit](../../README.md#gate-in-pre-commit-and-ci).
