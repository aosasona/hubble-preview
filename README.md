# Hubble

> [!WARNING]
>
> This document is a work-in-progress, please create a new issue if you need help or have any questions.

Hubble is a lightweight, unified and intelligent knowledge base management software designed for user-extensibility and self-hosting.

# Pre-requisites

- [Mise](https://mise.jdx.dev/) (_at least_ version 2025.4.2)
  > You can also just use Docker and install all dependencies ([go](https://go.dev), [pnpm](https://pnpm.io/)) locally
- Optionally, [`direnv`](https://direnv.net/) (to automatically load environment variables)
- A `postgres` instance (with the [`pgvector`](https://github.com/pgvector/pgvector) extension installed)
- An S3-compatible object store (preferrably [Minio](https://min.io/))
- _(Optional)_ For LLM-powered features like semantic search, an OpenAI-compatible server like Ollama, LocalAI, OpenAI API, etc

> [!NOTE]
>
> `Mise` is recommended since it will handle other dependencies automatically for you without having to install them globally. It will also handle custom tasks defined in the `.mise.toml` cofiguration file.

# Running locally

- Setup the environment variables by copying `.envrc.example` to `.envrc` and updating the values
- Start the backend services in development using either of the following:
  - Mise
  ```sh
  mise run web:dev
  ```
  - Docker Compose
  ```sh
  docker compose -f docker-compose.dev.yml up -d
  ```
  > This will start a postgres server, a minio server and an Ollama server
- Start the frontend using either of the following methods:
  - Mise
  ```sh
  mise run web:ui:dev
  ```
  - pnpm
  ```sh
  pnpm install
  pnpm dev
  ```

# Running in production

> [!NOTE]
>
> For the required environment variables, see `.envrc.example` and set them accordingly for your preferred deployment method.
>
> You do need a `Postgres` and `Minio` (recommended) or any other S3-compatible object storage server.

## Fly

For deploying to [fly.io](https://fly.io), this repository already contains a `fly.toml` in `./web/fly.toml` that properly sets up mounts, ports, etc.

## Docker

For deploying the application (excluding the `postgres` database which should be deployed separately to any preferred host) using Docker, you can use the production-optimized Dockerfile provided in `web/deployment/prod.Dockerfile`.

## Bare metal

If you don't want to use Docker, the `.mise.toml` config file includes a build task to build both the API and frontend into a single binary.

```sh
mise run web:build
```

The result will be placed in `web/bin/web` and can be copied & executed on any server without additional libraries (dependencies).

> [!NOTE]
>
> All of these methods still require you to run a separate Postgres instance with `pgvector` installed and enabled. Hubble ships with support for two key-value stores by default:
>
> - [`badger`](https://docs.hypermode.com/badger/overview) (this is the default; it doesn't require a separate deployment since it is embedded, it also configurable via environment variables)
> - [`etcd`](https://etcd.io/) - recommended for multi-server deployments
>
> Support for other key-value stores like `Valkey` will be added in the future.

# Packages

- `web`: The main binary that will provide the web interface and API.

# Environment variables

Here are the environment variables you can provide to hubble, some of them are optional but will disable some functionalities if not set. See the `.envrc.example` file.

```sh
export HUBBLE_PORT=3288 # The port to run hubble on
export HUBBLE_APP_URL="http://localhost:5173" # The URL to the frontend (to deliver appropriate emails, for example)
export HUBBLE_ENVIRONMENT="development" # can also be production or staging

export HUBBLE_ENABLE_OPEN_TELEMETRY=false # not used at the moment
export HUBBLE_ENABLE_EMAIL_SERVICES=true # not entirely used at the moment
export HUBBLE_ENABLE_DEBUG=false # this is overriden and set to true if HUBBLE_ENVIRONMENT is development

export HUBBLE_POSTGRES_DSN="postgres://hubble-web:hubble-web-password@postgres:5432/hubble?sslmode=disable"
# These are all required for email functionality (and since there are quite a bit of them; they are required in general)
export HUBBLE_SMTP_HOST=
export HUBBLE_SMTP_USERNAME=
export HUBBLE_SMTP_PASSWORD=
export HUBBLE_SMTP_PORT=
export HUBBLE_SMTP_FROM_NAME=
export HUBBLE_SMTP_FROM_URL=

export HUBBLE_DRIVER_KV="badgerdb" # can also be etcd (other ones will be added in the future)
export HUBBLE_ETCD_ENDPOINTS="" # set this if you are using etcd
export HUBBLE_BADGER_DB_PATH="badger.db" # set this if you are using badger, ensure the path is writable

export HUBBLE_KEY_COOKIE_SECRET="" # required to sign cookies

# This is a comma separated list of TOTP secrets in the format "v1_<secret>,v2_<secret>,v3_<secret>"
# This is required for TOTP keys encryption
export HUBBLE_KEY_TOTP_SECRETS=""

# This is required for file storage (can also be any other Minio-go-compatible server)
export HUBBLE_MINIO_ENDPOINT="minio:9000"
export HUBBLE_MINIO_ACCESS_KEY="minioadmin"
export HUBBLE_MINIO_SECRET_KEY="minioadmin"
export HUBBLE_MINIO_USE_SSL=false

export HUBBLE_PLUGINS_DIR=".plugins" # This is the directory where the plugins-related files are stored; default is `${cwd}/.plugins`

# Configure the OpenAI-compatible server here - optional
export HUBBLE_LLM_BASE_URL=""
export HUBBLE_LLM_API_KEY=""
# The name of the embedding model to use, e.g. text-embedding-ada-002, nomic-embed-text, etc
export HUBBLE_LLM_EMBEDDING_MODEL="" # You need to set this if you want to enable semantic vector generation for entry chunks

# Search controls
export HUBBLE_SEARCH_MODE="threshold" # or "score", The default "threshold" mode will execlude less accurate (relative to the top-scored) results i.e. below HUBBLE_SEARCH_THRESHOLD
export HUBBLE_SEARCH_THRESHOLD=30.0 # if you are using the threshold, you need to set this or it will default to `30.0`
```
