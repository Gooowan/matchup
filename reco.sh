#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat >&2 <<EOF
Usage: $0 INVITER_ID [BASE_URL]

This script simulates the recommendation flow:
  1) Creates a main test user and logs in.
  2) Creates 10 mock users with different coordinates.
  3) For each mock user, logs in and sets /me/profile with specific lat/lon.
  4) Fetches /matchup/feed for the main user and prints candidates.

Requirements:
  - Backend running (default at http://localhost:8000).
  - INVITER_ID must be an existing user UUID (registration requires it).
  - jq is optional but recommended for readable JSON output.
EOF
}

if [[ "${1-}" == "-h" || "${1-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 1 ]]; then
  echo "ERROR: INVITER_ID is required." >&2
  usage
  exit 1
fi

INVITER_ID="$1"
BASE_URL="${2:-http://localhost:8000}"

MAIN_EMAIL="reco-tester+main@example.com"
MAIN_PASSWORD="RecoTest123!"
MAIN_COOKIE_JAR=".reco_main_cookies.txt"
MAIN_AUTH_TOKEN=""

MOCK_PASSWORD="RecoMock123!"

declare -a MOCK_COORDS_LAT=(
  40.7128 40.7130 40.7132 40.7134 40.7136
  40.7138 40.7140 40.7142 40.7144 40.7146
)
declare -a MOCK_COORDS_LON=(
  -74.0060 -74.0058 -74.0056 -74.0054 -74.0052
  -74.0050 -74.0048 -74.0046 -74.0044 -74.0042
)

echo "=== Step 1: Create / login main test user ==="

MAIN_BODY_JSON=$(cat <<EOF
{
  "email": "$MAIN_EMAIL",
  "password": "$MAIN_PASSWORD",
  "inviter_id": "$INVITER_ID",
  "profile_data": {
    "name": "Reco Tester Main",
    "locale": "en"
  },
  "metadata": {
    "source": "reco.sh-main"
  }
}
EOF
)

TMP_BODY="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -H "Content-Type: application/json" \
  -X POST "${BASE_URL}/auth/register" \
  -d "$MAIN_BODY_JSON" || true)"
MAIN_REGISTER_BODY="$(cat "$TMP_BODY")"
rm -f "$TMP_BODY"

if [[ "$HTTP_STATUS" == 201 || "$HTTP_STATUS" == 200 ]]; then
  echo "Main user registered (status $HTTP_STATUS)."
else
  echo "Main user registration returned status $HTTP_STATUS (may already exist). Continuing..." >&2
  if [[ -n "$MAIN_REGISTER_BODY" ]]; then
    echo "$MAIN_REGISTER_BODY" >&2
  fi
fi

echo "Logging in main user $MAIN_EMAIL ..."
TMP_BODY="$(mktemp)"
TMP_HEADERS="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -D "$TMP_HEADERS" -w "%{http_code}" \
  -c "$MAIN_COOKIE_JAR" \
  -H "Content-Type: application/json" \
  -X POST "${BASE_URL}/auth/login" \
  -d "{\"email\":\"$MAIN_EMAIL\",\"password\":\"$MAIN_PASSWORD\"}" || true)"
MAIN_LOGIN_BODY="$(cat "$TMP_BODY")"

MAIN_AUTH_TOKEN="$(grep -i 'Set-Cookie: auth_token=' "$TMP_HEADERS" | head -n1 | sed -E 's/.*auth_token=([^;]+).*/\1/')"

rm -f "$TMP_BODY" "$TMP_HEADERS"

echo "Main login HTTP status: $HTTP_STATUS"
if [[ "$HTTP_STATUS" != 200 ]]; then
  echo "ERROR: Failed to login main user; cannot continue." >&2
  echo "$MAIN_LOGIN_BODY" >&2
  exit 1
fi
if [[ -z "$MAIN_AUTH_TOKEN" ]]; then
  echo "ERROR: Login succeeded but auth token was not found in Set-Cookie header." >&2
  exit 1
fi

if command -v jq >/dev/null 2>&1; then
  echo "$MAIN_LOGIN_BODY" | jq .
else
  echo "$MAIN_LOGIN_BODY"
fi

echo "Setting profile for main user ..."
MAIN_PROFILE_BODY=$(cat <<EOF
{
  "dance_styles": ["salsa", "bachata"],
  "latitude": 40.7135,
  "longitude": -74.0055,
  "visible": true,
  "dance_role": "LEAD",
  "dance_level": "INTERMEDIATE",
  "height_cm": 175,
  "bio": "Main test user for recommendation flow.",
  "birth_date": "1990-01-01",
  "gender": "other",
  "city": "New York"
}
EOF
)

TMP_BODY="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -H "Authorization: Bearer $MAIN_AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -X PUT "${BASE_URL}/me/profile" \
  -d "$MAIN_PROFILE_BODY" || true)"
MAIN_PROFILE_RESP="$(cat "$TMP_BODY")"
rm -f "$TMP_BODY"

echo "Main profile HTTP status: $HTTP_STATUS"
if [[ "$HTTP_STATUS" != 200 && "$HTTP_STATUS" != 201 ]]; then
  echo "ERROR: Failed to set main profile." >&2
  echo "$MAIN_PROFILE_RESP" >&2
  exit 1
fi

echo "=== Step 2: Create 10 mock users with different coordinates ==="

for i in $(seq 1 10); do
  EMAIL="reco-tester+mock${i}@example.com"
  LAT="${MOCK_COORDS_LAT[$((i-1))]}"
  LON="${MOCK_COORDS_LON[$((i-1))]}"
  COOKIE_JAR=".reco_mock_${i}_cookies.txt"
  AUTH_TOKEN=""

  echo "--- Mock user #$i: $EMAIL at ($LAT, $LON) ---"

  BODY_JSON=$(cat <<EOF
{
  "email": "$EMAIL",
  "password": "$MOCK_PASSWORD",
  "inviter_id": "$INVITER_ID",
  "profile_data": {
    "name": "Reco Mock $i",
    "locale": "en"
  },
  "metadata": {
    "source": "reco.sh-mock",
    "index": $i
  }
}
EOF
)

  TMP_BODY="$(mktemp)"
  HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -X POST "${BASE_URL}/auth/register" \
    -d "$BODY_JSON" || true)"
  REGISTER_BODY="$(cat "$TMP_BODY")"
  rm -f "$TMP_BODY"

  if [[ "$HTTP_STATUS" == 201 || "$HTTP_STATUS" == 200 ]]; then
    echo "Registered mock user #$i (status $HTTP_STATUS)."
  else
    echo "Registration for mock user #$i returned status $HTTP_STATUS (may already exist)." >&2
    if [[ -n "$REGISTER_BODY" ]]; then
      echo "$REGISTER_BODY" >&2
    fi
  fi

  echo "Logging in mock user #$i ..."
  TMP_BODY="$(mktemp)"
  TMP_HEADERS="$(mktemp)"
  HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -D "$TMP_HEADERS" -w "%{http_code}" \
    -c "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    -X POST "${BASE_URL}/auth/login" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$MOCK_PASSWORD\"}" || true)"
  LOGIN_BODY="$(cat "$TMP_BODY")"

  AUTH_TOKEN="$(grep -i 'Set-Cookie: auth_token=' "$TMP_HEADERS" | head -n1 | sed -E 's/.*auth_token=([^;]+).*/\1/')"

  rm -f "$TMP_BODY" "$TMP_HEADERS"

  if [[ "$HTTP_STATUS" != 200 ]]; then
    echo "ERROR: Failed to login mock user #$i (status $HTTP_STATUS)." >&2
    echo "$LOGIN_BODY" >&2
    continue
  fi
  if [[ -z "$AUTH_TOKEN" ]]; then
    echo "ERROR: Login succeeded but auth token was not found for mock user #$i." >&2
    continue
  fi

  PROFILE_BODY=$(cat <<EOF
{
  "dance_styles": ["salsa"],
  "latitude": $LAT,
  "longitude": $LON,
  "visible": true,
  "dance_role": "FOLLOW",
  "dance_level": "BEGINNER",
  "height_cm": 165,
  "bio": "Mock user $i for recommendation algorithm.",
  "birth_date": "1995-01-01",
  "gender": "female",
  "city": "New York"
}
EOF
)

  TMP_BODY="$(mktemp)"
  HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -X PUT "${BASE_URL}/me/profile" \
    -d "$PROFILE_BODY" || true)"
  PROFILE_RESP="$(cat "$TMP_BODY")"
  rm -f "$TMP_BODY"

  if [[ "$HTTP_STATUS" == 200 || "$HTTP_STATUS" == 201 ]]; then
    echo "Set profile for mock user #$i (status $HTTP_STATUS)."
  else
    echo "ERROR: Failed to set profile for mock user #$i (status $HTTP_STATUS)." >&2
    echo "$PROFILE_RESP" >&2
  fi
done

echo "=== Step 3: Fetch recommendation feed for main user ==="

TMP_BODY="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -H "Authorization: Bearer $MAIN_AUTH_TOKEN" \
  -H "Accept: application/json" \
  "${BASE_URL}/matchup/feed?limit=10" || true)"
FEED_BODY="$(cat "$TMP_BODY")"
rm -f "$TMP_BODY"

echo "Feed HTTP status: $HTTP_STATUS"

if [[ "$HTTP_STATUS" != 200 ]]; then
  echo "ERROR: Failed to fetch feed for main user." >&2
  if command -v jq >/dev/null 2>&1; then
    echo "$FEED_BODY" | jq . >&2 || echo "$FEED_BODY" >&2
  else
    echo "$FEED_BODY" >&2
  fi
  exit 1
fi

echo "SUCCESS: Retrieved recommendation feed for main user."
if command -v jq >/dev/null 2>&1; then
  echo "$FEED_BODY" | jq .
else
  echo "$FEED_BODY"
fi

echo "You should see candidates reflecting the 10 mock users with different coordinates."

