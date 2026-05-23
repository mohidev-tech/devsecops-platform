#!/usr/bin/env bash
# End-to-end smoke test:
#   1. wait for api ready
#   2. port-forward api:80 -> localhost:8080
#   3. POST 5 jobs
#   4. wait for worker to drain
#   5. assert api_jobs_done >= 5 in /metrics
set -euo pipefail

NS="${NS:-app}"
RELEASE="${RELEASE:-app}"
JOBS="${JOBS:-5}"
TIMEOUT="${TIMEOUT:-60}"

echo "[smoke] waiting for api rollout"
kubectl -n "$NS" rollout status deploy/"$RELEASE"-api --timeout=120s

echo "[smoke] port-forwarding"
kubectl -n "$NS" port-forward svc/"$RELEASE"-api 18080:80 > /tmp/pf.log 2>&1 &
PF_PID=$!
trap 'kill $PF_PID 2>/dev/null || true' EXIT
sleep 2

echo "[smoke] posting $JOBS jobs"
for i in $(seq 1 "$JOBS"); do
  curl -sf -X POST -H "Content-Type: application/json" \
    -d "{\"n\":$i}" http://localhost:18080/api/v1/jobs > /dev/null
done

echo "[smoke] waiting up to ${TIMEOUT}s for worker to drain"
for _ in $(seq 1 "$TIMEOUT"); do
  done_n=$(curl -sf http://localhost:18080/metrics | awk '/^api_jobs_done /{print $2}')
  if [[ "${done_n:-0}" -ge "$JOBS" ]]; then
    echo "[smoke] OK: api_jobs_done=$done_n"
    exit 0
  fi
  sleep 1
done

echo "[smoke] FAIL: drain did not complete within ${TIMEOUT}s"
curl -s http://localhost:18080/metrics | grep ^api_
exit 1
