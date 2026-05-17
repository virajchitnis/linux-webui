#!/usr/bin/env bash
set -euo pipefail

WORKDIR=/tmp/lwui-e2e
mkdir -p "$WORKDIR/tls"

cp /app/e2e/fixtures/server.toml "$WORKDIR/config.toml"

echo "Starting linux-webui server..."
cd "$WORKDIR"
/app/linux-webui --config "$WORKDIR/config.toml" > "$WORKDIR/server.log" 2>&1 &
SERVER_PID=$!

echo "Waiting for server to be ready..."
for i in $(seq 1 30); do
  if curl -sk https://127.0.0.1:8443/health > /dev/null 2>&1; then
    echo "Server ready after ${i}s"
    break
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "Server process died. Logs:"
    cat "$WORKDIR/server.log"
    exit 1
  fi
  sleep 1
done

if ! curl -sk https://127.0.0.1:8443/health > /dev/null 2>&1; then
  echo "Server did not become ready in 30s. Logs:"
  cat "$WORKDIR/server.log"
  exit 1
fi

cd /app/e2e
npx playwright test
EXIT=$?

echo ""
echo "=== Server logs ==="
cat "$WORKDIR/server.log"

kill "$SERVER_PID" 2>/dev/null || true
exit $EXIT
