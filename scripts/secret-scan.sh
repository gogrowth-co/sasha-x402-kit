#!/usr/bin/env bash
# secret-scan.sh — pre-commit gate. Fails the commit if secret-shaped material is staged.
# Defense-in-depth on top of .gitignore. Tuned to NOT false-positive on the public hex
# (Casper account/public keys, package hashes, tx hashes) this repo legitimately contains.
#
# Heuristics ALWAYS run (filename + private-key block + known token shapes). gitleaks, if
# installed, runs additionally as a content scanner — it is NOT a substitute for the
# filename rail (gitleaks passes fake/low-entropy key bodies; the filename block does not).
set -uo pipefail

fail() { echo "🚫 secret-scan BLOCK: $1" >&2; exit 1; }

staged_files=$(git diff --cached --name-only --diff-filter=ACM || true)
diff=$(git diff --cached -U0 --diff-filter=ACM || true)

# 1) never allow key/secret files by NAME (the robust rail — independent of content)
if printf '%s\n' "$staged_files" | grep -qiE '(\.pem$|\.key$|(^|/)\.env($|\.)|(^|/)secret_key|(^|/)keys/)'; then
  fail "a .pem/.key/.env/secret_key/keys file is staged"
fi

# 2) private-key blocks in content
if printf '%s' "$diff" | grep -qE -- '-----BEGIN ([A-Z ]+ )?PRIVATE KEY-----'; then
  fail "PEM private-key block in staged content"
fi

# 3) known high-confidence credential token shapes (NOT plain hex)
if printf '%s' "$diff" | grep -qE '(sk-[A-Za-z0-9]{20,}|ghp_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{30,}|xox[baprs]-[A-Za-z0-9-]{10,}|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{30,})'; then
  fail "a known credential token shape is staged"
fi

# 4) gitleaks (additional content scan) if available
if command -v gitleaks >/dev/null 2>&1; then
  gitleaks protect --staged --redact --no-banner >/dev/null 2>&1 || fail "gitleaks flagged staged content"
fi

echo "✅ secret-scan: clean"
exit 0
