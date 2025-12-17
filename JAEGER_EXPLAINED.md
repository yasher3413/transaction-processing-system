# What is Jaeger? (Distributed Tracing Explained)

## 🎯 What It Does

Jaeger shows you **the complete journey of a request** as it travels through all your services. It's like a GPS tracker for your transactions!

## 🔍 Think of It Like This:

Imagine you order food delivery:
- **Without Jaeger**: You know food arrived, but not what happened in between
- **With Jaeger**: You see the complete path:
  - Order placed → Restaurant received → Cook started → Food ready → Driver picked up → On the way → Delivered

## 🏗️ In Your Transaction System:

When you create a transaction, Jaeger shows:

```
1. API Service receives request
   └─> Validates input
   └─> Checks idempotency
   └─> Writes to database
   └─> Creates outbox event
   └─> Returns response

2. Publisher Service
   └─> Polls outbox table
   └─> Publishes to Kafka
   └─> Marks as published

3. Worker Service
   └─> Consumes from Kafka
   └─> Checks idempotency
   └─> Updates balance
   └─> Marks transaction processed
```

## 🎨 What You See in Jaeger

### Main View (http://localhost:16686)

1. **Search Page**: Find traces by:
   - Service name (api, worker, publisher)
   - Operation (create_transaction, process_event)
   - Time range
   - Tags (account_id, transaction_id)

2. **Trace View**: When you click a trace, you see:
   - **Timeline**: How long each step took
   - **Service Map**: Visual diagram of service interactions
   - **Spans**: Each operation (database query, Kafka publish, etc.)
   - **Timing**: Which parts are slow

## 💡 Why It's Useful

### 1. **Debugging Performance Issues**
- See which service is slow
- Find slow database queries
- Identify bottlenecks

### 2. **Understanding Flow**
- See exactly how a transaction flows through the system
- Understand the order of operations
- Verify everything happened correctly

### 3. **Debugging Errors**
- See where errors occurred
- Understand the context when something fails
- Track down issues across services

### 4. **Production Monitoring**
- Monitor request latency
- See error rates
- Understand system behavior under load

## 🎯 How to Use It

### Step 1: Create a Transaction

```bash
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID",
    "amount_cents": 10000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "trace-test-123"
  }'
```

### Step 2: View in Jaeger

1. Go to http://localhost:16686
2. Click **Search** (left sidebar)
3. Select:
   - **Service**: `api-service` (or `worker-service`, `publisher-service`)
   - **Operation**: `POST /v1/transactions` (or any operation)
   - **Lookback**: Last 1 hour
4. Click **Find Traces**

### Step 3: Explore a Trace

1. Click on any trace in the results
2. You'll see:
   - **Timeline view**: Horizontal bars showing duration
   - **Service breakdown**: Which services were involved
   - **Span details**: Click any span to see:
     - Start/end time
     - Duration
     - Tags (account_id, transaction_id, etc.)
     - Logs (if any)

## 📊 What Each Service Shows

### API Service Traces
- HTTP request handling
- Database queries (idempotency check, insert)
- Outbox event creation
- Response time

### Publisher Service Traces
- Outbox polling
- Kafka message publishing
- Database updates (marking as published)

### Worker Service Traces
- Kafka message consumption
- Idempotency check
- Balance update
- Transaction status update

## 🎨 Example: Following a Transaction

1. **Create transaction** → See trace in `api-service`
2. **Wait a few seconds** → See trace in `publisher-service` (publishing event)
3. **Wait a few more seconds** → See trace in `worker-service` (processing)

All connected by the same `trace_id`!

## 🔍 Search Tips

### Find a Specific Transaction
- Use **Tags** search
- Add tag: `transaction_id=YOUR_TRANSACTION_ID`
- Or: `account_id=YOUR_ACCOUNT_ID`

### Find Slow Requests
- Sort by **Duration**
- Look for traces with long bars (red/orange)

### Find Errors
- Filter by **Error=true**
- See which service failed

## 🆚 Jaeger vs Grafana

| Jaeger (Tracing) | Grafana (Metrics) |
|-----------------|------------------|
| Shows **individual requests** | Shows **aggregate statistics** |
| "This specific transaction" | "Average latency" |
| Complete request journey | Overall system health |
| Debugging specific issues | Monitoring trends |

**Think of it this way:**
- **Grafana**: "How many transactions per second?"
- **Jaeger**: "Why did THIS transaction take so long?"

## 🎯 When to Use Each

### Use Jaeger When:
- Debugging a specific transaction
- Understanding why something is slow
- Seeing the complete flow
- Finding where errors occur

### Use Grafana When:
- Monitoring overall system health
- Seeing trends over time
- Alerting on thresholds
- Business metrics (total balance, transaction volume)

## 💡 Real-World Example

**Problem**: "Transactions are slow today"

**With Grafana**: 
- See average latency increased
- But don't know WHY

**With Jaeger**:
- Find a slow trace
- See it's stuck in database query
- Identify the specific query
- Fix the issue!

## 🎓 Summary

Jaeger is like a **flight tracker for your transactions**:
- See where they go
- See how long each step takes
- See if they get stuck anywhere
- Debug problems easily

It's especially useful when you have multiple services (API → Publisher → Worker) and need to understand how they work together!

