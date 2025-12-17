# Getting Started Guide

## 🎉 Your System is Running!

All services are up and running. Here's how to use the system:

## Quick Test - Create Your First Transaction

### Step 1: Create an Account

```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{"currency": "USD"}'
```

**Save the `id` from the response!** You'll need it for the next step.

### Step 2: Create a Transaction

Replace `YOUR_ACCOUNT_ID` with the ID from Step 1:

```bash
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID",
    "amount_cents": 10000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "my-first-transaction-123"
  }'
```

### Step 3: Check Transaction Status

Use the `id` from the transaction response:

```bash
curl http://localhost:8080/v1/transactions/TRANSACTION_ID \
  -H "X-API-Key: demo-api-key-12345"
```

The status will change from `PENDING` → `PROCESSING` → `PROCESSED` (takes a few seconds).

### Step 4: Check Account Balance

```bash
curl http://localhost:8080/v1/accounts/ACCOUNT_ID \
  -H "X-API-Key: demo-api-key-12345"
```

You should see the balance updated to `10000` cents ($100.00)!

## 🧪 Test Idempotency

Send the **same transaction request twice** with the same `idempotency_key`:

```bash
# First request
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID",
    "amount_cents": 5000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "same-key-456"
  }'

# Second request (same idempotency_key)
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID",
    "amount_cents": 5000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "same-key-456"
  }'
```

Both requests return the **same transaction ID** - no duplicate created!

## 📊 View Dashboards

### Grafana (Metrics & Dashboards)
- **URL**: http://localhost:3000
- **Username**: `admin`
- **Password**: `admin`
- View API request rates, latency, worker throughput, DLQ messages

### Prometheus (Raw Metrics)
- **URL**: http://localhost:9090
- Query metrics directly

### Jaeger (Distributed Tracing)
- **URL**: http://localhost:16686
- See how requests flow through the system

## 🔍 Monitor What's Happening

### View Service Logs

```bash
# API logs
docker logs transactions-api -f

# Worker logs (processing transactions)
docker logs transactions-worker -f

# Publisher logs (publishing events)
docker logs transactions-publisher -f
```

### Check Service Status

```bash
# All services
docker ps

# Specific service health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

## 🎯 How It Works (Simple Explanation)

1. **You create a transaction** → API saves it to database (status: PENDING)
2. **Publisher service** → Reads from database and publishes event to Kafka
3. **Worker service** → Consumes event, updates account balance, marks transaction as PROCESSED
4. **You check the account** → Balance is updated!

All of this happens **asynchronously** - the API responds immediately, processing happens in the background.

## 🧹 Clean Up

When you're done:

```bash
make down
```

To remove all data too:

```bash
make clean
```

## 🆘 Troubleshooting

### Services not responding?
```bash
docker ps  # Check if containers are running
docker logs transactions-api  # Check for errors
```

### Transaction stuck in PENDING?
- Check worker logs: `docker logs transactions-worker`
- Check if Kafka is working: `docker logs transactions-redpanda`

### Want to see all transactions?
```bash
curl http://localhost:8080/v1/transactions \
  -H "X-API-Key: demo-api-key-12345"
```

## 📚 Next Steps

- Try creating multiple accounts
- Create DEBIT transactions (subtract from balance)
- Try creating a transaction that would cause insufficient balance (should fail)
- Explore the Grafana dashboards to see metrics
- Check Jaeger to see distributed traces

