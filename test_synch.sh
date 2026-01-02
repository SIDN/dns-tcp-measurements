#!/usr/bin/env bash
set -euo pipefail

DNS_SERVER="127.0.0.1"   # we need to test whether the nameserver on our local machine is ready
PORT=4242
TEST_NAME="4o"               
SLEEP_SECONDS=1

date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] Waiting for DNS server at ${DNS_SERVER}:${PORT} to become ready..."

while true; do
  OUT=$(dig @"${DNS_SERVER}" -p "${PORT}" "${TEST_NAME}" 2>/dev/null || true)

  # Ready condition: status: NOERROR
  if echo "$OUT" | grep -q "status: NOERROR"; then
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] DNS query completed with status NOERROR. Treating DNS as ready."
    break
  fi

  sleep "${SLEEP_SECONDS}"
done

date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] DNS looks ready. Starting process B..."
