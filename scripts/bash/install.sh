#!/usr/bin/env bash
set -euo pipefail

OTELCOL_VERSION="${OTELCOL_VERSION:-0.96.0}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/otelcol"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) OTEL_ARCH=amd64 ;;
  aarch64|arm64) OTEL_ARCH=arm64 ;;
  *)
    echo "unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

if command -v otelcol >/dev/null 2>&1 && [ -f /etc/systemd/system/otelcol.service ]; then
  sudo systemctl is-active otelcol || sudo systemctl start otelcol
  otelcol --version
  echo "already installed"
  exit 0
fi

sudo mkdir -p "$CONFIG_DIR"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

TAR="otelcol_${OTELCOL_VERSION}_linux_${OTEL_ARCH}.tar.gz"
URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${OTELCOL_VERSION}/${TAR}"

curl -fsSL "$URL" -o "$TMP/$TAR"

declare -A OTELCOL_SHA256=(
  [amd64]="7f028065bbfce011dd1cedbbc5b30fdfe64d83f4e715597ef429fcebb49b1886"
  [arm64]="6a9c2aaf970bb09b410f14e22e6e09d888217142eed72fc2e1099f9bf48503b8"
)
EXPECTED_SHA256="${OTELCOL_SHA256[$OTEL_ARCH]}"
ACTUAL_SHA256="$(sha256sum "$TMP/$TAR" | awk '{print $1}')"
if [ "$ACTUAL_SHA256" != "$EXPECTED_SHA256" ]; then
  echo "checksum mismatch for $TAR: expected $EXPECTED_SHA256, got $ACTUAL_SHA256" >&2
  exit 1
fi

tar -xzf "$TMP/$TAR" -C "$TMP"
sudo install -m 755 "$TMP/otelcol" "${INSTALL_DIR}/otelcol"

if [ ! -f "$CONFIG_FILE" ]; then
  sudo tee "$CONFIG_FILE" > /dev/null <<'EOF'
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  debug:
    verbosity: basic

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [debug]
EOF
fi

sudo tee /etc/systemd/system/otelcol.service > /dev/null <<EOF
[Unit]
Description=OpenTelemetry Collector
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/otelcol --config=${CONFIG_FILE}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo otelcol validate --config="$CONFIG_FILE"
sudo systemctl daemon-reload
sudo systemctl enable otelcol
sudo systemctl restart otelcol
sudo systemctl is-active otelcol
otelcol --version
echo "installed"
