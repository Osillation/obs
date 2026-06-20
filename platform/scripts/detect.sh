#!/usr/bin/env bash
# detect.sh — Scans a project directory and outputs a JSON detection result.
# Usage: detect.sh [project-dir]
# Output: JSON to stdout
# Requires: jq, yq v4

set -euo pipefail

for _tool in jq yq; do
  command -v "$_tool" >/dev/null 2>&1 || { echo "detect.sh: '$_tool' required but not found" >&2; exit 1; }
done

PROJECT_DIR="$(cd "${1:-.}" && pwd -P)" || { echo "detect.sh: directory not found: ${1:-.}" >&2; exit 1; }

# Use temp files to collect results (bash 3.2 compatible, no mapfile/associative arrays)
TMPWORK="$(mktemp -d)"
trap 'rm -rf "$TMPWORK"' EXIT

: > "$TMPWORK/frameworks"
: > "$TMPWORK/databases"

deployment_mode="compose"
compose_file=""
credentials="{}"
app_network="$(basename "$PROJECT_DIR" | tr '[:upper:]' '[:lower:]' | tr -cd '[:alnum:]-')_default"

# ── Frameworks ──────────────────────────────────────────────────────────────

while IFS= read -r pkg; do
  [[ -f "$pkg" ]] || continue
  if jq -e '(.dependencies.next // .devDependencies.next) != null' "$pkg" >/dev/null 2>&1; then
    echo "nextjs" >> "$TMPWORK/frameworks"
  fi
  if jq -e '(.dependencies["@nestjs/core"] // .devDependencies["@nestjs/core"]) != null' "$pkg" >/dev/null 2>&1; then
    echo "nestjs" >> "$TMPWORK/frameworks"
  fi
done < <(find "$PROJECT_DIR" -name "package.json" \
  -not -path "*/node_modules/*" \
  -not -path "*/.next/*" \
  -maxdepth 4 2>/dev/null)

while IFS= read -r req; do
  [[ -f "$req" ]] || continue
  case "$(basename "$req")" in
    pyproject.toml)
      if grep -qiE '"fastapi' "$req" 2>/dev/null; then
        echo "fastapi" >> "$TMPWORK/frameworks"
      fi
      if grep -qiE '"flask' "$req" 2>/dev/null; then
        echo "flask" >> "$TMPWORK/frameworks"
      fi
      ;;
    requirements.txt)
      if grep -qiE "^fastapi" "$req" 2>/dev/null; then
        echo "fastapi" >> "$TMPWORK/frameworks"
      fi
      if grep -qiE "^flask" "$req" 2>/dev/null; then
        echo "flask" >> "$TMPWORK/frameworks"
      fi
      ;;
  esac
done < <(find "$PROJECT_DIR" \( -name "requirements.txt" -o -name "pyproject.toml" \) -maxdepth 4 2>/dev/null)

# ── Databases ────────────────────────────────────────────────────────────────

for compose_candidate in "$PROJECT_DIR/docker-compose.yml" "$PROJECT_DIR/docker-compose.yaml"; do
  [[ -f "$compose_candidate" ]] || continue
  compose_file="$compose_candidate"

  while IFS= read -r image; do
    [[ -z "$image" || "$image" == "null" ]] && continue
    case "$image" in
      postgres:*|postgresql:*)
        echo "postgresql" >> "$TMPWORK/databases"
        # Try map format first, then fall back to list format
        pg_db="$(yq e '.services[] | select(.image | test("^postgres")) | .environment.POSTGRES_DB // ""' "$compose_file" 2>/dev/null | head -1 || true)"
        if [[ -z "$pg_db" || "$pg_db" == "null" ]]; then
          pg_db="$(yq e '.services[] | select(.image | test("^postgres")) | .environment[] | select(. | test("^POSTGRES_DB=")) | sub("^POSTGRES_DB=", "")' "$compose_file" 2>/dev/null | head -1 || true)"
        fi
        pg_password="$(yq e '.services[] | select(.image | test("^postgres")) | .environment.POSTGRES_PASSWORD // ""' "$compose_file" 2>/dev/null | head -1 || true)"
        if [[ -z "$pg_password" || "$pg_password" == "null" ]]; then
          pg_password="$(yq e '.services[] | select(.image | test("^postgres")) | .environment[] | select(. | test("^POSTGRES_PASSWORD=")) | sub("^POSTGRES_PASSWORD=", "")' "$compose_file" 2>/dev/null | head -1 || true)"
        fi
        pg_host="$(yq e '.services[] | select(.image | test("^postgres")) | .container_name // ""' "$compose_file" 2>/dev/null | head -1)"
        [[ -z "$pg_host" || "$pg_host" == "null" ]] && pg_host="postgres"
        credentials="$(echo "$credentials" | jq \
          --arg host "$pg_host" --arg db "$pg_db" --arg password "$pg_password" \
          '. + {"postgresql": {"host": $host, "db": $db, "password": $password, "port": 5432}}')"
        ;;
      redis:*)
        echo "redis" >> "$TMPWORK/databases"
        redis_host="$(yq e '.services[] | select(.image | test("^redis")) | .container_name // ""' "$compose_file" 2>/dev/null | head -1)"
        [[ -z "$redis_host" || "$redis_host" == "null" ]] && redis_host="redis"
        credentials="$(echo "$credentials" | jq \
          --arg host "$redis_host" \
          '. + {"redis": {"host": $host, "port": 6379}}')"
        ;;
      mongo:*)
        echo "mongodb" >> "$TMPWORK/databases"
        mongo_host="$(yq e '.services[] | select(.image | test("^mongo")) | .container_name // ""' "$compose_file" 2>/dev/null | head -1)"
        [[ -z "$mongo_host" || "$mongo_host" == "null" ]] && mongo_host="mongo"
        credentials="$(echo "$credentials" | jq \
          --arg host "$mongo_host" \
          '. + {"mongodb": {"host": $host, "port": 27017}}')"
        ;;
    esac
  done < <(yq e '.services[].image // ""' "$compose_file" 2>/dev/null)

  break
done

# ── Deployment mode ──────────────────────────────────────────────────────────

if [[ -f "$PROJECT_DIR/Chart.yaml" ]] || [[ -f "$PROJECT_DIR/Helmfile.yaml" ]] || [[ -f "$PROJECT_DIR/helmfile.yaml" ]]; then
  deployment_mode="helm"
elif [[ -d "$PROJECT_DIR/k8s" ]] || [[ -d "$PROJECT_DIR/kubernetes" ]]; then
  deployment_mode="kubernetes"
elif find "$PROJECT_DIR" -maxdepth 3 -name "*.yaml" -not -path "*/node_modules/*" -exec grep -l "kind: Deployment" {} \; 2>/dev/null | grep -q .; then
  deployment_mode="kubernetes"
fi

# ── Build JSON ────────────────────────────────────────────────────────────────

# Deduplicate and build JSON arrays
frameworks_json="$(sort -u "$TMPWORK/frameworks" | jq -R . | jq -s .)"
databases_json="$(sort -u "$TMPWORK/databases" | jq -R . | jq -s .)"

jq -n \
  --argjson frameworks "$frameworks_json" \
  --argjson databases "$databases_json" \
  --arg deployment_mode "$deployment_mode" \
  --arg compose_file "$compose_file" \
  --arg app_network "$app_network" \
  --argjson credentials "$credentials" \
  '{
    frameworks: $frameworks,
    databases: $databases,
    deployment_mode: $deployment_mode,
    compose_file: $compose_file,
    app_network: $app_network,
    credentials: $credentials
  }'
