# MatchUp

## Development

### Seeding test profiles

The seeder creates 60-80 realistic dummy profiles with photos uploaded to MinIO.

```bash
# Seed 70 balanced-gender profiles, assign them to existing clubs
go run ./cmd/seed-profiles \
  --count 70 \
  --gender-split balanced \
  --assign-clubs

# Seed only female profiles
go run ./cmd/seed-profiles --count 35 --gender-split female

# Skip MinIO upload (use external placeholder URLs instead)
go run ./cmd/seed-profiles --count 70 --no-minio
```

The command reads the same env vars as the main API:

| Variable | Default |
|---|---|
| `DATABASE_URL` | built from `POSTGRES_*` vars |
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_ACCESS_KEY` | `minioadmin` |
| `MINIO_SECRET_KEY` | `minioadmin` |
| `MINIO_PUBLIC_ENDPOINT` | `http://localhost:9000` |

The seeder is **idempotent**: profiles whose email (`seed-NNN@matchup.local`) already exist are skipped.

**Verification after seeding:**
- Open the feed → confirm photos render.
- Open the map → confirm pins are spread across Ukrainian cities.
- Apply each filter dimension → confirm result sets shrink as expected.
