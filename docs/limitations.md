# Known limitations

- Exits if Postgres is unreachable at startup (no retry/back-off in-process).
- Health is static; it doesn't reflect DB availability.
- Cache is per-pod and only refreshed by `SIGUSR1`/restart.
- The reload flag is shared between goroutines without synchronisation (technically a data race; benign in practice).
- `amd64`-only image that runs as root unless the pod overrides it.
