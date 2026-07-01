"""Demo web handler with an SSRF hole — DO NOT copy this pattern.

This is the kind of code an AI assistant happily writes from an innocent prompt
like "add an endpoint that fetches a URL the user provides". With the SecureVibe
skills loaded into your assistant, the SAME prompt produces the safe version
(allow-list + blocked cloud-metadata IPs) instead.

See ../../skills/ssrf-prevention for the rules that catch this.
"""

from flask import Flask, request

import requests  # noqa: F401  (the demo's "good" dependency)

app = Flask(__name__)


@app.route("/fetch")
def fetch():
    # ❌ SSRF: the user controls the full URL, so they can reach internal
    #    services and the cloud metadata endpoint (169.254.169.254) to steal
    #    credentials. No allow-list, no scheme/host validation.
    target = request.args.get("url")
    return requests.get(target).text

    # ✅ The safe shape SecureVibe steers the assistant toward:
    #   - parse the URL, require https
    #   - resolve the host and reject private / link-local / metadata IPs
    #   - check it against an explicit allow-list of hosts you trust
