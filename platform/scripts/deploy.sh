#!/usr/bin/env bash
# deploy.sh — Deploy the obs observability stack for a client project.
# Usage: deploy.sh --client <name> --project <path>

set -euo pipefail

OBS_PLATFORM_DIR="$(cd "$(dirname "$0")/.." && pwd -P)"
client=""
project_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --client)  client="$2";      shift 2 ;;
    --project) project_dir="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

[[ -z "$client" ]]      && { echo "Usage: deploy.sh --client <name> --project <path>" >&2; exit 1; }
[[ -z "$project_dir" ]] && { echo "Usage: deploy.sh --client <name> --project <path>" >&2; exit 1; }

OBS_CLIENT_DIR="$OBS_PLATFORM_DIR/clients/$client"
[[ -d "$OBS_CLIENT_DIR" ]] || { echo "Client directory not found: $OBS_CLIENT_DIR" >&2; exit 1; }

export OBS_PLATFORM_DIR OBS_CLIENT_DIR

# ── Step 1: Detect project ───────────────────────────────────────────────────
echo "-> Scanning project at $project_dir ..."
detection=$("$OBS_PLATFORM_DIR/scripts/detect.sh" "$project_dir")
echo "   Detected: $detection"

frameworks=$(echo "$detection" | jq -r '[.frameworks[]] | join(",")' )
databases=$(echo "$detection"  | jq -r '[.databases[]]  | join(",")' )
deployment_mode=$(echo "$detection" | jq -r '.deployment_mode')
app_network=$(echo "$detection" | jq -r '.app_network')

echo "   Frameworks:  ${frameworks:-none}"
echo "   Databases:   ${databases:-none}"
echo "   Mode:        $deployment_mode"
echo "   App network: $app_network"

# Allow deployment.yml to override detected mode
if [[ -f "$OBS_CLIENT_DIR/deployment.yml" ]]; then
  override_mode=$(yq e '.mode // ""' "$OBS_CLIENT_DIR/deployment.yml" 2>/dev/null || true)
  [[ -n "$override_mode" ]] && deployment_mode="$override_mode"
fi

# ── Step 2: Determine receivers ──────────────────────────────────────────────
_recv_tmp="$(mktemp)"
echo "hostmetrics" >> "$_recv_tmp"
echo "docker_stats" >> "$_recv_tmp"

echo "$detection" | jq -r '.databases[]' | while read -r db; do
  case "$db" in
    postgresql) echo "postgresql" >> "$_recv_tmp" ;;
    redis)      echo "redis"      >> "$_recv_tmp" ;;
    mongodb)    echo "mongodb"    >> "$_recv_tmp" ;;
  esac
done

receivers=()
while IFS= read -r r; do
  receivers+=("$r")
done < <(sort -u "$_recv_tmp")
rm -f "$_recv_tmp"

# Build OTel config args string
otel_config_args="--config /etc/otelcol/collector.base.yaml"
for r in "${receivers[@]}"; do
  otel_config_args="$otel_config_args --config /etc/otelcol/receivers/${r}.yaml"
done
receivers_str="${receivers[*]}"

# ── Step 3: Load config.env ──────────────────────────────────────────────────
echo "-> Loading client config ..."
set -a
source "$OBS_CLIENT_DIR/config.env"
CLIENT_APP_NETWORK="$app_network"
set +a
export CLIENT_APP_NETWORK

# ── Step 4: Validate ─────────────────────────────────────────────────────────
echo "-> Validating config ..."
"$OBS_PLATFORM_DIR/scripts/validate.sh" \
  --client-dir "$OBS_CLIENT_DIR" \
  --receivers "$receivers_str"

# ── Step 5: Generate Nginx config ────────────────────────────────────────────
echo "-> Generating Nginx config ..."
acl_rules=$(grep -v '^#' "$OBS_CLIENT_DIR/acl.conf" | grep -v '^$' || true)

sed \
  -e "s|OBS_CLIENT_DOMAIN|${CLIENT_DOMAIN}|g" \
  -e "s|OBS_INGEST_TOKEN_VALUE|${OBS_INGEST_TOKEN}|g" \
  "$OBS_PLATFORM_DIR/config/nginx/nginx.conf.template" \
  > "$OBS_PLATFORM_DIR/config/nginx/nginx.conf"

# Inject ACL rules — use a temp file to avoid sed delimiter collisions
_acl_tmp="$(mktemp)"
echo "$acl_rules" > "$_acl_tmp"
sed -e "/OBS_ACL_RULES_PLACEHOLDER/r $_acl_tmp" \
    -e "/OBS_ACL_RULES_PLACEHOLDER/d" \
    "$OBS_PLATFORM_DIR/config/nginx/snippets/ip-acl.conf.template" \
    > "$OBS_PLATFORM_DIR/config/nginx/snippets/ip-acl.conf"
rm -f "$_acl_tmp"

# ── Step 6: Deploy ───────────────────────────────────────────────────────────
if [[ "$deployment_mode" == "compose" ]]; then
  echo "-> Deploying via Docker Compose ..."

  # Determine stack variant
  dbs_sorted=$(echo "$databases" | tr ',' '\n' | grep -v '^$' | sort | tr '\n' '-' | sed 's/-$//')
  fw_sorted=$(echo "$frameworks"  | tr ',' '\n' | grep -v '^$' | sort | tr '\n' '-' | sed 's/-$//')

  # Join non-empty parts with -
  variant_parts=()
  [[ -n "$fw_sorted" ]]  && variant_parts+=("$fw_sorted")
  [[ -n "$dbs_sorted" ]] && variant_parts+=("$dbs_sorted")
  variant=$(IFS='-'; echo "${variant_parts[*]}")
  variant_file="$OBS_PLATFORM_DIR/stacks/compose/variants/${variant}.yml"

  compose_args=(
    -f "$OBS_PLATFORM_DIR/stacks/compose/base/signoz.yml"
    -f "$OBS_PLATFORM_DIR/stacks/compose/base/posthog.yml"
    -f "$OBS_PLATFORM_DIR/stacks/compose/base/otel-collector.yml"
    -f "$OBS_PLATFORM_DIR/stacks/compose/base/nginx.yml"
  )

  if [[ -f "$variant_file" ]]; then
    compose_args+=(-f "$variant_file")
  else
    echo "   Warning: no variant file for '$variant' — OTel Collector won't bridge to app network." >&2
  fi

  OTEL_CONFIG_ARGS="$otel_config_args" \
  docker compose \
    "${compose_args[@]}" \
    --env-file "$OBS_CLIENT_DIR/config.env" \
    up -d \
    --remove-orphans

elif [[ "$deployment_mode" == "helm" ]]; then
  echo "-> Deploying via Helm ..."
  helm repo update
  helm upgrade --install "obs-${client}" "$OBS_PLATFORM_DIR/stacks/helm" \
    -f "$OBS_PLATFORM_DIR/stacks/helm/signoz/values.base.yaml" \
    -f "$OBS_PLATFORM_DIR/stacks/helm/posthog/values.base.yaml" \
    --namespace "obs-${client}" \
    --create-namespace \
    --set "global.clientDomain=${CLIENT_DOMAIN}"

else
  echo "Deployment mode '$deployment_mode' is not yet supported by deploy.sh" >&2
  exit 1
fi

# ── Step 7: Summary ──────────────────────────────────────────────────────────
echo ""
echo "Observability stack deployed for client: $client"
echo ""
echo "Dashboards (requires Osillation employee TLS cert + allowlisted IP):"
echo "  SigNoz:  https://dashboard.${CLIENT_DOMAIN}/signoz"
echo "  PostHog: https://dashboard.${CLIENT_DOMAIN}/posthog"
echo ""
echo "Ingest endpoint:"
echo "  https://ingest.${CLIENT_DOMAIN}"
echo ""
echo "Next step — instrument the client application:"
echo "  Copy files from $OBS_PLATFORM_DIR/instrumentation/<framework>/ into each service."
echo "  Set in each service env:"
echo "    OTEL_EXPORTER_OTLP_ENDPOINT=https://ingest.${CLIENT_DOMAIN}/v1"
echo "    OBS_INGEST_TOKEN=<from config.env>"
echo "    OTEL_SERVICE_NAME=<service-name>"
echo "    OTEL_SERVICE_NAMESPACE=${client}"
