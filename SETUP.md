# Setup Instructions

## Initial Setup

1. **Generate go.sum:**
   ```bash
   go mod tidy
   ```

2. **Start all services:**
   ```bash
   make up
   ```

3. **Verify services are running:**
   - Check logs: `docker compose -f infra/docker-compose.yml logs -f`
   - Health checks:
     - API: `curl http://localhost:8080/health`
     - Worker: `curl http://localhost:8081/health`
     - Publisher: `curl http://localhost:8082/health`

## Troubleshooting

### Services won't start
- Ensure Docker has enough resources (4GB RAM, 2 CPUs recommended)
- Check if ports 8080, 5432, 6379, 9090, 3000, 16686 are available
- View logs: `docker compose -f infra/docker-compose.yml logs [service-name]`

### Database connection errors
- Wait for postgres to be fully ready (healthcheck passes)
- Check postgres logs: `docker compose -f infra/docker-compose.yml logs postgres`

### Kafka/Redpanda issues
- Wait for redpanda to be healthy before starting other services
- Check topics: `docker compose -f infra/docker-compose.yml exec redpanda rpk topic list --brokers localhost:19092`

### Build errors
- Ensure `go.sum` exists: run `go mod tidy`
- Clear Docker cache: `docker system prune -a`

## Development Workflow

1. Make code changes
2. Rebuild specific service: `docker compose -f infra/docker-compose.yml build [api|worker|publisher]`
3. Restart service: `docker compose -f infra/docker-compose.yml restart [api|worker|publisher]`
4. View logs: `docker compose -f infra/docker-compose.yml logs -f [service-name]`

## Testing

Run E2E tests (requires services to be running):
```bash
make e2e
```

Run all tests:
```bash
make test
```


