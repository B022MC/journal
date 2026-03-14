# Remote Journal Config Overlay

Date: 2026-03-14

Purpose: close `JRV-020` by providing a reusable way to render remote
single-db config files for all seven services without editing the default local
development YAMLs or committing live credentials.

## What This Overlay Does

- Renders a dedicated remote config for `journal-api`, `admin-api`,
  `user.rpc`, `paper.rpc`, `rating.rpc`, `news.rpc`, and `admin.rpc`.
- Replaces every service `DB.DataSource` with the injected remote journal DSN.
- Forces `DB.ReadWriteSplit: false` in every rendered file.
- Removes replica-only keys from the rendered config: `Policy` and `Replicas`.
- Rewrites `journal-api` and `admin-api` RPC client blocks to direct localhost
  endpoints (`127.0.0.1:9001-9005`) so validation traffic stays inside the
  local smoke process tree.
- Removes top-level RPC `Etcd` blocks from the rendered service YAMLs, so the
  validation services do not register into the live discovery namespace.
- Rewrites both `Redis` and `CacheRedis` passwords when a remote validation
  Redis password is provided.

## Render Command

```bash
python3 backend/scripts/render_remote_journal_configs.py \
  --dsn "$REMOTE_JOURNAL_DSN" \
  --output-dir /tmp/journal-remote-validation
```

The script writes:

- `/tmp/journal-remote-validation/api/etc/journal-api.remote.yaml`
- `/tmp/journal-remote-validation/admin-api/etc/admin-api.remote.yaml`
- `/tmp/journal-remote-validation/rpc/user/etc/user.remote.yaml`
- `/tmp/journal-remote-validation/rpc/paper/etc/paper.remote.yaml`
- `/tmp/journal-remote-validation/rpc/rating/etc/rating.remote.yaml`
- `/tmp/journal-remote-validation/rpc/news/etc/news.remote.yaml`
- `/tmp/journal-remote-validation/rpc/admin/etc/admin.remote.yaml`

Before each render, the script deletes only the previously generated
`*.remote.yaml` targets under the chosen output directory. This keeps repeated
runs idempotent without touching unrelated files in the same folder.

## Activation Examples

- `go run backend/api/journal.go -f /tmp/journal-remote-validation/api/etc/journal-api.remote.yaml`
- `go run backend/admin-api/admin.go -f /tmp/journal-remote-validation/admin-api/etc/admin-api.remote.yaml`
- `go run backend/rpc/user/user.go -f /tmp/journal-remote-validation/rpc/user/etc/user.remote.yaml`
- `go run backend/rpc/paper/paper.go -f /tmp/journal-remote-validation/rpc/paper/etc/paper.remote.yaml`
- `go run backend/rpc/rating/rating.go -f /tmp/journal-remote-validation/rpc/rating/etc/rating.remote.yaml`
- `go run backend/rpc/news/news.go -f /tmp/journal-remote-validation/rpc/news/etc/news.remote.yaml`
- `go run backend/rpc/admin/admin.go -f /tmp/journal-remote-validation/rpc/admin/etc/admin.remote.yaml`

## Review Checklist

After rendering, the following grep must return no matches:

```bash
rg -n "ReadWriteSplit: true|127\\.0\\.0\\.1:13306|127\\.0\\.0\\.1:13307|127\\.0\\.0\\.1:13308|Policy:|Replicas:" \
  /tmp/journal-remote-validation
```

The script-level regression check is:

```bash
python3 -m unittest discover -s backend/scripts -p 'test_*.py'
```

The rendered files are valid for this issue when:

- all seven files exist
- every `DataSource` equals the injected remote DSN
- no rendered file contains replica addresses or `ReadWriteSplit: true`
- API configs resolve RPCs through direct localhost endpoints
- RPC configs omit the top-level `Etcd` block
- the default tracked YAMLs in `backend/*/etc/*.yaml` remain unchanged

## Current Limitation

- This issue validates the config overlay generation locally, not a live remote
  startup. Actual DSN usage in service logs is deferred to `JRV-040`.
