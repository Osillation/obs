#!/usr/bin/env bash
# gen-certs.sh — Manage mTLS client certificates for Osillation employees.
# Usage: OBS_CLIENT_DIR=clients/mari gen-certs.sh <command> [args]

set -euo pipefail
umask 077

: "${OBS_CLIENT_DIR:?OBS_CLIENT_DIR must be set (e.g. clients/mari)}"

TLS_DIR="${OBS_CLIENT_DIR}/tls"
CA_KEY="$TLS_DIR/ca.key"
CA_CERT="$TLS_DIR/ca.crt"
CERT_VALIDITY_DAYS=365

# Use system openssl (LibreSSL on macOS) to avoid missing config issues
# with Homebrew openssl installations that lack openssl.cnf
OPENSSL="${OPENSSL:-/usr/bin/openssl}"
if ! "$OPENSSL" version &>/dev/null; then
  OPENSSL="openssl"
fi

cmd="${1:-}"
shift || true

case "$cmd" in
  init-ca)
    mkdir -p "$TLS_DIR"
    if [[ -f "$CA_KEY" ]]; then
      echo "CA already exists at $CA_KEY — skipping." >&2
      exit 0
    fi
    "$OPENSSL" genrsa -out "$CA_KEY" 4096 2>/dev/null
    "$OPENSSL" req -new -x509 -days 3650 \
      -key "$CA_KEY" \
      -out "$CA_CERT" \
      -subj "/CN=Osillation-CA/O=Osillation/C=PK" 2>/dev/null
    chmod 600 "$CA_KEY"
    echo "CA created: $CA_CERT"
    ;;

  add-employee)
    name="${1:?Usage: gen-certs.sh add-employee <name>}"
    [[ "$name" =~ ^[a-zA-Z0-9_-]+$ ]] || { echo "Invalid name '$name': use only letters, digits, hyphens, underscores" >&2; exit 1; }
    [[ ! -f "$CA_KEY" ]] && { echo "Run init-ca first." >&2; exit 1; }
    "$OPENSSL" genrsa -out "$TLS_DIR/${name}.key" 2048 2>/dev/null
    "$OPENSSL" req -new \
      -key "$TLS_DIR/${name}.key" \
      -out "$TLS_DIR/${name}.csr" \
      -subj "/CN=${name}/O=Osillation/C=PK" 2>/dev/null
    "$OPENSSL" x509 -req \
      -days "$CERT_VALIDITY_DAYS" \
      -in "$TLS_DIR/${name}.csr" \
      -CA "$CA_CERT" \
      -CAkey "$CA_KEY" \
      -CAcreateserial \
      -out "$TLS_DIR/${name}.crt" 2>/dev/null
    rm "$TLS_DIR/${name}.csr"
    chmod 600 "$TLS_DIR/${name}.key"
    echo "Certificate issued: $TLS_DIR/${name}.crt"
    echo "Key (keep private): $TLS_DIR/${name}.key"
    echo ""
    echo "Employee install instructions:"
    echo "  macOS: double-click ${name}.crt, add to Keychain, then import ${name}.key"
    echo "  Chrome/Firefox: Settings -> Certificates -> Import -> select ${name}.crt + ${name}.key"
    ;;

  revoke)
    name="${1:?Usage: gen-certs.sh revoke <name>}"
    rm -f "$TLS_DIR/${name}.crt" "$TLS_DIR/${name}.key"
    echo "Revoked: $name — reload nginx to apply"
    ;;

  list)
    echo "Active employee certificates in $TLS_DIR:"
    found=0
    for cert in "$TLS_DIR"/*.crt; do
      [[ -f "$cert" ]] || continue
      [[ "$cert" == *"/ca.crt" ]] && continue
      [[ "$cert" == *"/server.crt" ]] && continue
      name=$(basename "$cert" .crt)
      expiry=$("$OPENSSL" x509 -noout -enddate -in "$cert" 2>/dev/null | cut -d= -f2)
      echo "  $name — expires: $expiry"
      found=1
    done
    if [[ "$found" -eq 0 ]]; then echo "  (none)"; fi
    ;;

  *)
    echo "Usage: gen-certs.sh <init-ca|add-employee <name>|revoke <name>|list>" >&2
    exit 1
    ;;
esac
