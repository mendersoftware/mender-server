# mender-server

Mender Server monorepo

## Getting started

This tutorial teaches you how to setup up the Mender Server for evaluation purposes.
The tutorial will take you through starting the Mender Server and adding a virtual device to learn how Mender works.

> [!NOTE]
> This docker composition is only meant for evaluation purposes.
> Please use [the helm chart](https://github.com/mendersoftware/mender-helm) in production environments.

### Prerequisites

* Before you start the Mender Server, make sure you have [Docker Compose](https://docs.docker.com/compose/install/) installed.
Mender Server requires Docker Compose version `v2.23.1` or newer.

  ```bash
  docker compose version
  ```

* You will also need to update your local DNS or host rewrites, for example by appending to `/etc/hosts`:

  ```bash
  echo '127.0.0.1   docker.mender.io s3.docker.mender.io' | sudo tee -a /etc/hosts
  ```

### Starting the Mender Server

1. Start by cloning and entering this repository:

   ```bash
   git clone https://github.com/mendersoftware/mender-server && cd mender-server
   ```

2. To start evaluating Mender, you can bring up your Mender Server by running:

   ```bash
   docker compose up -d
   ```

3. Once all containers are running, you can create your initial user using the following command:

   ```bash
   MENDER_USERNAME=admin@docker.mender.io
   MENDER_PASSWORD=password123
   docker compose exec useradm useradm create-user --username "$MENDER_USERNAME" --password "$MENDER_PASSWORD"
   ```

4. Visit [https://docker.mender.io](https://docker.mender.io) and sign in using the credentials from the snippet above.

5. To add a virtual client with the composition, run the following command:

   ```bash
   docker compose run -d client
   ```

## Evaluating Mender Enterprise

> [!WARNING]
> We strongly advise using [the helm chart](https://github.com/mendersoftware/mender-helm) for production setups.
> This docker composition is only meant for evaluation purposes.

For paying customers with access to Mender Enterprise components, you can evaluate the Mender Server Enterprise by replacing the snippet in step 2 above to:

```bash
export COMPOSE_FILE="docker-compose.yml:compose/docker-compose.enterprise.yml"
docker compose up -d
```

> [!NOTE]
> This composition will not work without a paid license with access to registry.mender.io.

To initialize the admin user, use the following snippet in step 3:

```bash
MENDER_NAME=Admin
MENDER_USERNAME=admin@docker.mender.io
MENDER_PASSWORD=password123
docker compose exec tenantadm tenantadm create-org --username "$MENDER_USERNAME" --password "$MENDER_PASSWORD" --name "$MENDER_NAME"
```

Visit [https://localhost](https://localhost) and sign in using the credentials from the snippet above.

## Testing Gitlab build artifacts

To test/debug artifacts from GitLab CI, you can pull the image straight from the internal registry (after generating credentials).
For testing, you need to setup the following environment variables:

```bash
COMMIT_SHA="$(git rev-parse HEAD)"
export MENDER_IMAGE_REGISTRY=registry.gitlab.com
export MENDER_IMAGE_REPOSITORY=northern.tech/mender/mender-server
export MENDER_IMAGE_TAG=build-$COMMIT_SHA
export MENDER_IMAGE_TAG_TEST=test-$COMMIT_SHA
```
