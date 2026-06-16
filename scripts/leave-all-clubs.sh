#!/usr/bin/env bash
# Remove a user from all clubs (memberships + trainer links + primary_club_id).
# Does NOT delete clubs the user owns — prints a warning if any.
#
# Usage:
#   ./scripts/leave-all-clubs.sh --email user@example.com
#   ./scripts/leave-all-clubs.sh --id <uuid>
#   ./scripts/leave-all-clubs.sh --email user@example.com --yes

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose -f compose.yml -f compose.prod.yml)

EMAIL=""
USER_ID=""
YES=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --email) EMAIL="$2"; shift 2 ;;
    --id)    USER_ID="$2"; shift 2 ;;
    --yes)   YES=true; shift ;;
    -h|--help)
      cat <<EOF
Usage: $0 --email USER@example.com [--yes]
       $0 --id UUID [--yes]

Removes user from:
  - club_members
  - club_trainers
  - trainer_students (as trainer or dancer)
  - clears profiles.primary_club_id

Clubs owned by the user (clubs.owner_user_id) are kept; reassign manually if needed.
EOF
      exit 0
      ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$EMAIL" && -z "$USER_ID" ]]; then
  echo "Provide --email or --id" >&2
  exit 1
fi

dotenv_get() {
  grep -E "^${1}=" .env 2>/dev/null | head -1 | cut -d= -f2- \
    | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | tr -d '"'"'"
}

PGUSER="$(dotenv_get POSTGRES_USER)"
PGDB="$(dotenv_get POSTGRES_DB)"

psql_exec() {
  "${COMPOSE[@]}" exec -T db psql -U "$PGUSER" -d "$PGDB" -v ON_ERROR_STOP=1 "$@"
}

if [[ -n "$EMAIL" ]]; then
  safe_email="${EMAIL//\'/\'\'}"
  USER_ID="$(psql_exec -tA -c "SELECT id::text FROM users WHERE email = '${safe_email}'")"
else
  if ! [[ "$USER_ID" =~ ^[0-9a-fA-F-]{36}$ ]]; then
    echo "Invalid UUID" >&2
    exit 1
  fi
  USER_ID="$(psql_exec -tA -c "SELECT id::text FROM users WHERE id = '$USER_ID'::uuid")"
fi

if [[ -z "$USER_ID" ]]; then
  echo "User not found" >&2
  exit 1
fi

echo "User ID: $USER_ID"

owned="$(psql_exec -tA -c "SELECT coalesce(string_agg(name, ', '), '') FROM clubs WHERE owner_user_id = '$USER_ID'::uuid AND is_active = true")"
members="$(psql_exec -tA -c "SELECT count(*) FROM club_members WHERE user_id = '$USER_ID'::uuid")"
trainers="$(psql_exec -tA -c "SELECT count(*) FROM club_trainers WHERE trainer_user_id = '$USER_ID'::uuid")"

echo "Club memberships:  $members"
echo "Trainer links:     $trainers"
if [[ -n "$owned" ]]; then
  echo "Owned clubs (kept): $owned"
fi

if [[ "$members" == "0" && "$trainers" == "0" ]]; then
  echo "Nothing to remove."
  exit 0
fi

if ! $YES; then
  read -r -p "Remove all club links for this user? [y/N] " ans
  [[ "$ans" == "y" || "$ans" == "Y" ]] || { echo "Aborted."; exit 0; }
fi

psql_exec <<SQL
BEGIN;

DELETE FROM club_members WHERE user_id = '$USER_ID'::uuid;
DELETE FROM club_trainers WHERE trainer_user_id = '$USER_ID'::uuid;
DELETE FROM trainer_students WHERE trainer_user_id = '$USER_ID'::uuid OR dancer_user_id = '$USER_ID'::uuid;
UPDATE profiles SET primary_club_id = NULL WHERE user_id = '$USER_ID'::uuid;

COMMIT;
SQL

echo "Done. User removed from all clubs."
if [[ -n "$owned" ]]; then
  echo "Note: user still owns club(s): $owned"
fi
