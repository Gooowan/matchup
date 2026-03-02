#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 EMAIL PASSWORD INVITER_ID [BASE_URL]" >&2
  echo "Default BASE_URL is http://localhost:8000" >&2
}

if [[ "${1-}" == "-h" || "${1-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 3 ]]; then
  echo "ERROR: Missing required arguments." >&2
  usage
  exit 1
fi

EMAIL="$1"
PASSWORD="$2"
INVITER_ID="$3"
BASE_URL="${4:-http://localhost:8000}"

BODY_JSON=$(cat <<EOF
{
  "email": "$EMAIL",
  "password": "$PASSWORD",
  "inviter_id": "$INVITER_ID",
  "profile_data": {
    "name": "CLI Test User",
    "locale": "en"
  },
  "metadata": {
    "source": "reg.sh"
  }
}
EOF
)

echo "Registering user at $BASE_URL/auth/register ..."

TMP_BODY="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -H "Content-Type: application/json" \
  -X POST "${BASE_URL}/auth/register" \
  -d "$BODY_JSON" || true)"
RESPONSE_BODY="$(cat "$TMP_BODY")"
rm -f "$TMP_BODY"

echo "HTTP status: $HTTP_STATUS"

if [[ "$HTTP_STATUS" == 201 || "$HTTP_STATUS" == 200 ]]; then
  echo "SUCCESS: User registered."
  if command -v jq >/dev/null 2>&1; then
    echo "$RESPONSE_BODY" | jq .
  else
    echo "$RESPONSE_BODY"
  fi
else
  echo "FAILED: Registration request was not successful." >&2
  if command -v jq >/dev/null 2>&1; then
    echo "$RESPONSE_BODY" | jq . >&2 || echo "$RESPONSE_BODY" >&2
  else
    echo "$RESPONSE_BODY" >&2
  fi
  exit 1
fi

echo
echo "Attempting to login newly registered user and verify endpoints ..."

LOGIN_BODY=$(cat <<EOF
{
  "email": "$EMAIL",
  "password": "$PASSWORD"
}
EOF
)

TMP_BODY="$(mktemp)"
TMP_HEADERS="$(mktemp)"
LOGIN_STATUS="$(curl -sS -o "$TMP_BODY" -D "$TMP_HEADERS" -w "%{http_code}" \
  -H "Content-Type: application/json" \
  -X POST "${BASE_URL}/auth/login" \
  -d "$LOGIN_BODY" || true)"
LOGIN_RESP="$(cat "$TMP_BODY")"

# Extract JWT token from Set-Cookie header (auth_token)
AUTH_TOKEN="$(grep -i 'Set-Cookie: auth_token=' "$TMP_HEADERS" | head -n1 | sed -E 's/.*auth_token=([^;]+).*/\1/')"

rm -f "$TMP_BODY" "$TMP_HEADERS"

echo "Login after registration HTTP status: $LOGIN_STATUS"
if [[ "$LOGIN_STATUS" != 200 ]]; then
  echo "ERROR: Could not login newly registered user; skipping further verification." >&2
  if command -v jq >/dev/null 2>&1; then
    echo "$LOGIN_RESP" | jq . >&2 || echo "$LOGIN_RESP" >&2
  else
    echo "$LOGIN_RESP" >&2
  fi
  exit 1
fi

if command -v jq >/dev/null 2>&1; then
  echo "$LOGIN_RESP" | jq .
else
  echo "$LOGIN_RESP"
fi

FAILED=0

curl_json() {
  local method="$1"
  local path="$2"
  local expected="$3"
  local body="${4-}"

  local tmp status resp
  tmp="$(mktemp)"

  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$tmp" -w "%{http_code}" \
      -H "Authorization: Bearer $AUTH_TOKEN" \
      -H "Content-Type: application/json" \
      -X "$method" "${BASE_URL}${path}" \
      -d "$body" || true)"
  else
    status="$(curl -sS -o "$tmp" -w "%{http_code}" \
      -H "Authorization: Bearer $AUTH_TOKEN" \
      -H "Accept: application/json" \
      -X "$method" "${BASE_URL}${path}" || true)"
  fi

  resp="$(cat "$tmp")"
  rm -f "$tmp"

  echo
  echo "$method $path -> HTTP $status (expected: $expected)"
  if command -v jq >/dev/null 2>&1; then
    echo "$resp" | jq . 2>/dev/null || echo "$resp"
  else
    echo "$resp"
  fi

  IFS=',' read -r -a exp_codes <<< "$expected"
  local ok=1
  for code in "${exp_codes[@]}"; do
    if [[ "$status" == "$code" ]]; then
      ok=0
      break
    fi
  done

  if [[ $ok -ne 0 ]]; then
    echo "!! Endpoint ${method} ${path} FAILED (got $status, expected $expected)" >&2
    FAILED=1
  fi
}

echo
echo "=== Verifying authenticated endpoints for newly registered user ==="

# 1) Core user profile and inviter
curl_json "GET" "/user/profile" "200"
curl_json "GET" "/user/inviter" "200"

# 2) Update locale
curl_json "POST" "/user/locale" "200" '{"locale":"en"}'

# 3) Update basic profile data
curl_json "POST" "/user/profile/update" "200" '{"first_name":"CLI","last_name":"Tester"}'

# 4) Recommendation profile lifecycle
curl_json "GET" "/me/profile" "200,404"
curl_json "PUT" "/me/profile" "200,201" '{
  "dance_styles": ["salsa","bachata"],
  "latitude": 40.7128,
  "longitude": -74.0060,
  "visible": true,
  "dance_role": "LEAD",
  "dance_level": "INTERMEDIATE",
  "height_cm": 180,
  "bio": "Created from reg.sh verification flow.",
  "birth_date": "1990-01-01",
  "gender": "other",
  "city": "New York"
}'
curl_json "GET" "/me/profile" "200"

# 5) Preferences lifecycle
curl_json "GET" "/me/preferences" "200,404"
curl_json "PUT" "/me/preferences" "200" '{
  "max_distance_km": 50,
  "age_range": [25, 40],
  "preferred_roles": ["FOLLOW","LEAD"],
  "preferred_styles": ["salsa","bachata"]
}'
curl_json "GET" "/me/preferences" "200"

# 6) Matchup feed
curl_json "GET" "/matchup/feed?limit=5" "200"

echo
if [[ "$FAILED" -eq 0 ]]; then
  echo "All verification endpoint calls succeeded with expected statuses."
else
  echo "Some verification endpoint calls FAILED. See logs above." >&2
fi

exit "$FAILED"

