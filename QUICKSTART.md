# Quick Start Guide

## Prerequisites Check

Before starting, you need Docker installed. Check if it's installed:

```bash
docker --version
```

If you see "command not found", you need to install Docker Desktop first.

## Step 1: Install Docker Desktop (if needed)

### macOS

**Easiest way (if you have Homebrew):**
```bash
brew install --cask docker
```

Then start Docker Desktop:
```bash
open -a Docker
```

**Or download manually:**
1. Go to https://www.docker.com/probrew install --cask dockerducts/docker-desktop/
2. Download Docker Desktop for Mac
3. Install and open Docker Desktop
4. Wait for Docker to start (you'll see a whale icon in the menu bar)

### Verify Docker is Running

```bash
docker ps
# Should show an empty list (no error)
```

## Step 2: Start the System

Once Docker is running:

```bash
make up
```

This will:
- Build all Docker images
- Start all services (Postgres, Redis, Redpanda, Prometheus, Grafana, Jaeger, API, Worker, Publisher)
- Take about 2-3 minutes the first time

## Step 3: Verify Services

Wait about 30 seconds for services to start, then check:

```bash
# Check API health
curl http://localhost:8080/health

# Check all services
docker ps
```

You should see 10+ containers running.

## Step 4: Access Services

- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

## Step 5: Test the API

```bash
# Create an account
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{"currency": "USD"}'

# Save the account_id from the response, then create a transaction
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID_HERE",
    "amount_cents": 10000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "test-123"
  }'
```

## Troubleshooting

### "docker: command not found"
- Docker Desktop is not installed or not running
- Install Docker Desktop (see Step 1)
- Make sure Docker Desktop is running (whale icon in menu bar)

### "Cannot connect to Docker daemon"
- Docker Desktop is not running
- Start Docker Desktop: `open -a Docker`
- Wait for it to fully start (whale icon should be steady, not animating)

### Port already in use
- Another service is using the port
- Stop the conflicting service or change ports in `infra/docker-compose.yml`

### Services won't start
- Check Docker has enough resources: Docker Desktop → Settings → Resources
- Recommended: 4GB RAM, 2 CPUs
- Check logs: `docker compose -f infra/docker-compose.yml logs`

## Stop the System

```bash
make down
```

## Clean Everything (including data)

```bash
make clean
```


