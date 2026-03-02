#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 EMAIL PASSWORD [BASE_URL] [COOKIE_JAR]" >&2
  echo "Default BASE_URL is http://localhost:8000" >&2
  echo "Default COOKIE_JAR is .auth_cookies.txt in repo root" >&2
}

if [[ "${1-}" == "-h" || "${1-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 2 ]]; then
  echo "ERROR: Missing required arguments." >&2
  usage
  exit 1
fi

EMAIL="$1"
PASSWORD="$2"
BASE_URL="${3:-http://localhost:8000}"
COOKIE_JAR="${4:-.auth_cookies.txt}"

BODY_JSON=$(cat <<EOF
{
  "email": "$EMAIL",
  "password": "$PASSWORD"
}
EOF
)

echo "Logging in user at $BASE_URL/auth/login ..."
echo "Cookies (if any) will be stored in $COOKIE_JAR"

TMP_BODY="$(mktemp)"
TMP_HEADERS="$(mktemp)"
HTTP_STATUS="$(curl -sS -o "$TMP_BODY" -D "$TMP_HEADERS" -w "%{http_code}" \
  -c "$COOKIE_JAR" \
  -H "Content-Type: application/json" \
  -X POST "${BASE_URL}/auth/login" \
  -d "$BODY_JSON")"

RESPONSE_BODY="$(cat "$TMP_BODY")"

# Extract JWT token from Set-Cookie header (auth_token)
AUTH_TOKEN="$(grep -i 'Set-Cookie: auth_token=' "$TMP_HEADERS" | head -n1 | sed -E 's/.*auth_token=([^;]+).*/\1/')"

rm -f "$TMP_BODY" "$TMP_HEADERS"

echo "HTTP status: $HTTP_STATUS"

if [[ "$HTTP_STATUS" == 200 ]]; then
  echo "SUCCESS: Login succeeded."
  if command -v jq >/dev/null 2>&1; then
    echo "$RESPONSE_BODY" | jq .
  else
    echo "$RESPONSE_BODY"
  fi
  echo "You can now use this cookie jar with other scripts, e.g.:"
  echo "  curl -H \"Authorization: Bearer $AUTH_TOKEN\" ${BASE_URL}/me/profile"
else
  echo "FAILED: Login request was not successful." >&2
  if command -v jq >/dev/null 2>&1; then
    echo "$RESPONSE_BODY" | jq . >&2 || echo "$RESPONSE_BODY" >&2
  else
    echo "$RESPONSE_BODY" >&2
  fi
  exit 1
fi

echo
echo "=== Verifying authenticated endpoints with stored cookies ==="

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

  # expected can be a single code or "code1,code2"
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
  "bio": "Created from login.sh verification flow.",
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

