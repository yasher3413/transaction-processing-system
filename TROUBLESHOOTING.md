# Troubleshooting Guide

## Docker Not Found Error

### Symptom
```
/bin/sh: docker: command not found
make: *** [up] Error 127
```

### Solution

The Makefile has been updated to automatically add Docker to PATH, so `make up` should work now.

**For a permanent fix**, add Docker to your shell PATH:

**For zsh (default on macOS):**
```bash
echo 'export PATH="/Applications/Docker.app/Contents/Resources/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Or run the helper script:**
```bash
./scripts/fix-docker-path.sh
source ~/.zshrc
```

**To fix the broken symlink (optional):**
```bash
sudo rm /usr/local/bin/docker
sudo ln -s /Applications/Docker.app/Contents/Resources/bin/docker /usr/local/bin/docker
```

### Verify Docker Works
```bash
docker --version
docker compose version
```

## Other Common Issues

### Port Already in Use
**Error:** `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution:**
- Find what's using the port: `lsof -i :8080`
- Stop the conflicting service
- Or change the port in `infra/docker-compose.yml`

### Docker Desktop Not Running
**Error:** `Cannot connect to Docker daemon`

**Solution:**
- Open Docker Desktop: `open -a Docker`
- Wait for it to fully start (whale icon should be steady)
- Verify: `docker ps`

### Services Won't Start
**Error:** Containers exit immediately

**Solution:**
1. Check logs: `docker compose -f infra/docker-compose.yml logs [service-name]`
2. Check Docker resources: Docker Desktop → Settings → Resources
   - Recommended: 4GB RAM, 2 CPUs
3. Check disk space: `df -h`

### Database Connection Errors
**Error:** `connection refused` or `dial tcp: lookup postgres`

**Solution:**
- Wait for postgres to be healthy: `docker compose -f infra/docker-compose.yml ps`
- Check postgres logs: `docker compose -f infra/docker-compose.yml logs postgres`
- Restart postgres: `docker compose -f infra/docker-compose.yml restart postgres`

### Kafka/Redpanda Issues
**Error:** Topics not created or connection refused

**Solution:**
- Wait for redpanda to be healthy
- Check topics: `docker compose -f infra/docker-compose.yml exec redpanda rpk topic list --brokers localhost:9092`
- Manually create topics if needed:
  ```bash
  docker compose -f infra/docker-compose.yml exec redpanda rpk topic create transactions --brokers redpanda:9092
  docker compose -f infra/docker-compose.yml exec redpanda rpk topic create transactions.dlq --brokers redpanda:9092
  ```

### Build Errors
**Error:** `go: cannot find module` or build failures

**Solution:**
- Ensure `go.sum` exists (optional - Docker will download deps)
- Clear Docker cache: `docker system prune -a`
- Rebuild: `docker compose -f infra/docker-compose.yml build --no-cache`

### Permission Errors (Linux)
**Error:** `permission denied` when running docker commands

**Solution:**
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

## Getting Help

1. Check service logs: `docker compose -f infra/docker-compose.yml logs [service]`
2. Check service status: `docker compose -f infra/docker-compose.yml ps`
3. View all logs: `docker compose -f infra/docker-compose.yml logs -f`

