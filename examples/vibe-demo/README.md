# vibe-demo — feel SecureVibe in 2 minutes

> ⚠️ **This folder is intentionally insecure.** It contains a typosquatted
> dependency, hardcoded secrets, and an SSRF hole on purpose. Never copy this
> code into a real project. The fake credentials are well-known *test* values,
> not live secrets.

This is the fastest way to *feel* what SecureVibe does — for yourself, not in a
slide. It's the same workspace used to record the [hero demo GIF](../../docs/assets/demo.gif).

## The 30-second version

From the repo root:

```bash
make demo
```

That builds the CLI and runs the **gate** over this folder. Watch it block the
typosquat and the leaked secrets — **offline, no API key, before any commit.**

## The full story

For the 3-act, copy-paste walkthrough (including the "AI fixes its own code at
generation time" moment), see **[DEMO.md](./DEMO.md)**.

## What's in here

| File | The planted problem | Who catches it |
|------|---------------------|----------------|
| `requirements.txt` | `colourama` — a typosquat of `colorama` | `check-typosquat` / `scan-dependencies` / `gate` |
| `config.py` | Hardcoded GitHub + Stripe tokens | `scan-secrets` / `gate` |
| `app.py` | SSRF: user controls the fetched URL | the `ssrf-prevention` skill (at generation time) + secure-code review |
