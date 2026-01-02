#!/usr/bin/env bash
set -euo pipefail

DNS_SERVER="127.0.0.1"   # or the IP
PORT=4242
TEST_NAME="4o"                # your test name
SLEEP_SECONDS=1

date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] Waiting for DNS server at ${DNS_SERVER}:${PORT} to become ready..."

while true; do
  # Capture stdout only; ignore stderr (connection errors etc.)
  OUT=$(dig @"${DNS_SERVER}" -p "${PORT}" "${TEST_NAME}" 2>/dev/null || true)

  # Uncomment this if you want to see the raw output for debugging:
  # echo "---- DIG OUTPUT START ----"
  # echo "$OUT"
  # echo "---- DIG OUTPUT END ----"

  # Ready condition: status: NOERROR
  if echo "$OUT" | grep -q "status: NOERROR"; then
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] DNS query completed with status NOERROR. Treating DNS as ready."
    break
  fi

  # echo "[B] DNS not ready yet (no NOERROR), sleeping ${SLEEP_SECONDS}s..."
  sleep "${SLEEP_SECONDS}"
done

date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] DNS looks ready. Starting process B..."
