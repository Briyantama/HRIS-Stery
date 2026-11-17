# Makefile reference

Run from the repository root. Needs GNU Make and Bash (Git Bash on Windows) and a `.env` file.

| Target   | What it does                                                                                   |
| -------- | ---------------------------------------------------------------------------------------------- |
| `up`     | Starts `deploy/docker-compose.yml`, runs `go run ./cmd/migrate up`, starts the API in the background (pid in `.run/api.pid`, log in `.run/api.log`) |
| `down`   | Stops the API and the containers                                                                |
| `status` | Shows whether the API pid is alive and calls `/healthz`                                         |
| `build`  | Builds the API binary to `.run/api`                                                             |
| `check`  | Runs `scripts/check.sh` (format, vet, tests, web checks), the same gate as CI                  |
