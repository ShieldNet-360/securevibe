"""Demo config with hardcoded secrets — DO NOT copy this pattern.

These are fake, well-known *test* values (not live credentials). They exist so
the scanner has something obvious to catch. In a real project these would be a
breach waiting to happen — load secrets from the environment instead.
"""

# ❌ Hardcoded credentials — exactly what an AI assistant leaves behind.
GITHUB_TOKEN = "ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789"
STRIPE_KEY = "sk_live_4eC39HqLyjWDarjtT1zdp7dcAbCdEfGhItuv"

# ✅ The right way:
#   import os
#   GITHUB_TOKEN = os.environ["GITHUB_TOKEN"]
