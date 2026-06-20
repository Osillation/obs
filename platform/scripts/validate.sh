#!/usr/bin/env bash
# validate.sh — Pre-deploy config validation.
# Usage: validate.sh --client-dir <path> --receivers "postgresql redis"

set -euo pipefail

OBS_PLATFORM_DIR="${OBS_PLATFORM_DIR:-$(cd "$(dirname "$0")/.." && pwd -P)}"
client_dir=""
receivers=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --client-dir) client_dir="$2"; shift 2 ;;
    --receivers)  receivers="$2";  shift 2 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

[[ -z "$client_dir" ]] && { echo "Usage: validate.sh --client-dir <path> --receivers \"postgresql redis\"" >&2; exit 1; }

errors=()

# Load config.env
config_env="$client_dir/config.env"
[[ -f "$config_env" ]] || { echo "FAIL: $config_env not found" >&2; exit 1; }
set -a; source "$config_env"; set +a

# Required vars
for var in CLIENT_DOMAIN OBS_INGEST_TOKEN; do
  if [[ -z "${!var:-}" ]]; then
    errors+=("MISSING: $var is not set in config.env")
  fi
done

# Token length
token_val="${OBS_INGEST_TOKEN:-}"
token_len="${#token_val}"
if [[ "$token_len" -lt 32 ]]; then
  errors+=("INVALID: OBS_INGEST_TOKEN must be at least 32 characters (got $token_len) — generate with: openssl rand -hex 32")
fi

# CA cert
if [[ ! -f "$client_dir/tls/ca.crt" ]]; then
  errors+=("MISSING: ca.crt not found in $client_dir/tls/ — run gen-certs.sh init-ca")
fi

# Employee certs
employee_cert_count=0
for cert in "$client_dir/tls/"*.crt; do
  [[ -f "$cert" ]] || continue
  [[ "$cert" == *"/ca.crt" ]]     && continue
  [[ "$cert" == *"/server.crt" ]] && continue
  employee_cert_count=$((employee_cert_count + 1))
done

if [[ "$employee_cert_count" -eq 0 ]]; then
  errors+=("MISSING: no employee cert found in $client_dir/tls/ — run gen-certs.sh add-employee <name>")
fi

# Receiver files
for receiver in $receivers; do
  [[ -z "$receiver" ]] && continue
  receiver_file="$OBS_PLATFORM_DIR/config/otel/receivers/${receiver}.yaml"
  if [[ ! -f "$receiver_file" ]]; then
    errors+=("MISSING: receiver config not found for '$receiver': $receiver_file")
  fi
done

# Report
if [[ "${#errors[@]}" -gt 0 ]]; then
  echo "Validation failed:" >&2
  for err in "${errors[@]}"; do
    echo "  x $err" >&2
  done
  exit 1
fi

echo "Validation passed."
