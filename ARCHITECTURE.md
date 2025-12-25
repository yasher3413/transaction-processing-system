# Architecture Deep Dive

## System Design Principles

### 1. Event-Driven Architecture
- **Decoupling**: API service is decoupled from transaction processing
- **Scalability**: Workers can scale independently based on load
- **Resilience**: Failures in one component don't cascade immediately

### 2. Transactional Outbox Pattern
**Problem**: Ensuring events are published only after database transaction commits.

**Solution**: 
- Write transaction + outbox event in same DB transaction
- Publisher polls outbox table and publishes to Kafka
- Mark events as published only after successful Kafka delivery

**Tradeoffs**:
- ✅ Guarantees consistency
- ✅ No distributed transactions needed
- ⚠️ Small delay between DB write and event publish (polling interval)
- ⚠️ Publisher is single point of failure (mitigated by running multiple instances)

**Alternative Considered**: Change Data Capture (CDC) with Debezium
- ✅ Lower latency
- ✅ No polling needed
- ⚠️ More complex setup
- ⚠️ Requires additional infrastructure

### 3. Idempotency Strategy

#### API-Level Idempotency
- **Key**: `(account_id, idempotency_key)`
- **Mechanism**: Database unique constraint + SELECT before INSERT
- **Race Condition Handling**: 
  - Use `SERIALIZABLE` isolation level
  - Catch unique constraint violation and fetch existing record

#### Consumer-Level Idempotency
- **Key**: `event_id` (unique per event envelope)
- **Mechanism**: `processed_events` table with atomic INSERT
- **Pattern**: Insert-before-apply
  ```sql
  INSERT INTO processed_events (event_id, transaction_id)
  VALUES ($1, $2)
  ON CONFLICT (event_id) DO NOTHING
  ```
- **Guarantee**: Exactly-once effect (not exactly-once delivery)

### 4. Retry and DLQ Strategy

**Retry Policy**:
- Exponential backoff: `attempt * base_delay`
- Max attempts: 5
- Retryable errors: Database connection failures, transient network issues
- Non-retryable errors: Validation failures, insufficient balance

**DLQ Handling**:
- After max retries, message sent to `transactions.dlq` topic
- Original message headers preserved + error details
- Manual reprocessing required (future: admin endpoint)

### 5. Database Design

**Isolation Levels**:
- Transaction creation: `SERIALIZABLE` (prevents phantom reads in idempotency check)
- Transaction processing: `SERIALIZABLE` (ensures balance consistency)

**Locking Strategy**:
- Account row locked with `FOR UPDATE` during balance update
- Prevents concurrent balance modifications
- Ensures balance consistency

**Indexes**:
- `accounts.status` - for filtering active accounts
- `transactions.account_id` - for account transaction queries
- `transactions.status` - for status-based queries
- `transactions.idempotency_key` - for idempotency checks
- `outbox_events.status, created_at` - for publisher polling
- `processed_events.event_id` - for consumer idempotency

### 6. Message Broker Design

**Partitioning**:
- Key: `transaction_id` or `account_id`
- Ensures events for same account/transaction are ordered
- Tradeoff: May cause hot partitions if one account has high volume

**Topics**:
- `transactions`: Main event stream
- `transactions.dlq`: Dead-letter queue

**Consumer Groups**:
- Single consumer group for workers
- Enables horizontal scaling (multiple worker instances)

### 7. Observability Strategy

**Metrics** (Prometheus):
- Request/response metrics (latency, count, errors)
- Business metrics (transactions processed, balances updated)
- System metrics (retries, DLQ volume)

**Logging** (Structured JSON):
- Correlation IDs for request tracing
- Event IDs for event tracking
- Transaction/Account IDs for business context

**Tracing** (OpenTelemetry + Jaeger):
- Distributed tracing across services
- Trace propagation via message headers
- Performance bottleneck identification

## Failure Scenarios and Handling

### 1. API Service Fails During Transaction Creation
- **Scenario**: DB transaction committed but response not sent
- **Handling**: Client retries with same idempotency key → returns existing transaction

### 2. Publisher Fails After DB Commit
- **Scenario**: Outbox event created but not published
- **Handling**: Publisher retries on next poll (event remains PENDING)

### 3. Worker Fails During Processing
- **Scenario**: Event consumed but processing fails
- **Handling**: 
  - Offset not committed → Kafka redelivers
  - Retry with exponential backoff
  - After max retries → DLQ

### 4. Duplicate Event Delivery
- **Scenario**: Kafka delivers same event twice
- **Handling**: `processed_events` table prevents duplicate processing

### 5. Concurrent Transaction Processing
- **Scenario**: Multiple workers process events for same account
- **Handling**: 
  - Row-level locking (`FOR UPDATE`)
  - Serializable isolation level
  - Atomic balance updates

## Scalability Considerations

### Horizontal Scaling
- **API**: Stateless → scale horizontally
- **Publisher**: Can run multiple instances (idempotent publishing)
- **Worker**: Consumer group enables multiple instances

### Database Scaling
- **Read Replicas**: For read-heavy workloads (account queries)
- **Partitioning**: Partition transactions table by account_id or date
- **Connection Pooling**: Tune pool sizes based on load

### Kafka Scaling
- **Partitions**: Increase partitions for higher throughput
- **Replication**: Multi-broker setup for production
- **Consumer Lag**: Monitor and scale workers based on lag

## Security Considerations

### Current Implementation
- API key authentication (simple, for demo)
- Input validation
- SQL injection prevention (parameterized queries)

### Production Improvements Needed
- TLS/SSL for all connections
- OAuth2/JWT for authentication
- Role-based access control (RBAC)
- Secrets management (Vault, AWS Secrets Manager)
- Rate limiting per client
- Audit logging for sensitive operations

## Performance Optimizations

### Current
- Database connection pooling
- Batch outbox publishing
- Efficient indexing

### Future Improvements
- Redis caching for account lookups
- Batch transaction processing
- Async API responses (fire-and-forget pattern)
- Database query optimization
- Message compression

## Monitoring and Alerting

### Key Metrics to Alert On
- High error rate (> 1% of requests)
- High latency (p95 > 1s)
- DLQ message volume (indicates systemic issues)
- Consumer lag (workers falling behind)
- Database connection pool exhaustion

### Dashboards
- Real-time system health
- Business metrics (transaction volume, success rate)
- Error analysis (by type, by service)
- Performance trends

## Testing Strategy

### Unit Tests
- Idempotency logic
- Retry classifier
- Message serialization

### Integration Tests
- API → DB (transaction + outbox)
- Publisher → Kafka
- Worker → DB (balance update)

### E2E Tests
- Full transaction flow
- Idempotency verification
- Failure scenarios (insufficient balance → DLQ)

## Deployment Strategy

### Current (Development)
- Docker Compose for local development
- Single-region deployment

### Production Recommendations
- Kubernetes for orchestration
- Multi-region deployment (active-active or active-passive)
- Blue-green deployments for zero downtime
- Feature flags for gradual rollouts
- Canary deployments for risk mitigation




