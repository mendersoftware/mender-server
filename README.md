# mender-server
Mender Server monorepo

## Getting started

Before you start the Mender server, make sure you have [Docker Compose](https://docs.docker.com/compose/install/) installed.
Mender server requires Docker Compose version `v2.21.0` or newer.
```
docker compose version
```

You will also need to update your local DNS or host rewrites, for example by appending to `/etc/hosts`:
```bash
echo '127.0.0.1   docker.mender.io s3.docker.mender.io' | sudo tee -a /etc/hosts
```

> [!NOTE]
> This docker composition is only meant for evaluation purposes.
> Please use [the helm chart](https://github.com/mendersoftware/mender-helm) in production environments.

To start evaluating Mender, you can bring up your Mender server by running:
```
docker compose up -d
```

Once all containers are running, you can create your initial user using the following command:
```
MENDER_USERNAME=admin@docker.mender.io
MENDER_PASSWORD=password123
docker compose exec useradm useradm create-user --username "$MENDER_USERNAME" --password "$MENDER_PASSWORD"
```

Visit [https://localhost](https://localhost) and sign in using the credentials from the snippet above.

### Adding a virtual client

To add a virtual client with the composition, run the following command:
```bash
docker compose run -d client
```

### Evaluating Mender enterprise

> [!WARNING]
> We strongly advise using [the helm chart](https://github.com/mendersoftware/mender-helm) for production setups.
> This docker composition is only meant for evaluation purposes.

For paying customers with access to private enterprise components, you can evaluate the Mender Server enterprise by replacing the above snippets from the following section.
To start the server run:
```
export COMPOSE_FILE="docker-compose.yml:compose/docker-compose.enterprise.yml"
docker compose up -d
```

> [!IMPORTANT]
> This composition will not work without a paid license with access to registry.mender.io.

To initialize the admin user, use the following snippet:
```
MENDER_NAME=Admin
MENDER_USERNAME=admin@docker.mender.io
MENDER_PASSWORD=password123
docker compose exec tenantadm tenantadm create-org --username "$MENDER_USERNAME" --password "$MENDER_PASSWORD" --name "$MENDER_NAME"
```

Visit [https://localhost](https://localhost) and sign in using the credentials from the snippet above.
