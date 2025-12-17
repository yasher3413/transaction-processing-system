# How to See Account Balance in Grafana

## Option 1: Using Prometheus Metrics (Recommended)

I've added metrics that export account balances. After rebuilding the API service, you'll see balances in Grafana.

### Step 1: Rebuild API Service

```bash
cd infra && docker compose build api && docker compose up -d api
```

### Step 2: Access Grafana

1. Go to http://localhost:3000
2. Login: `admin` / `admin`
3. The "Account Balance Dashboard" should appear automatically

### Step 3: View Balance

- Go to **Dashboards** → **Account Balance Dashboard**
- You'll see:
  - Account balances over time
  - Total balance by currency
  - Number of accounts

## Option 2: Query Directly in Grafana

### Add PostgreSQL Data Source

1. Go to **Configuration** (gear icon) → **Data Sources**
2. Click **Add data source**
3. Select **PostgreSQL**
4. Configure:
   - **Host**: `postgres:5432`
   - **Database**: `transactions`
   - **User**: `postgres`
   - **Password**: `postgres`
   - **SSL Mode**: `disable`
5. Click **Save & Test**

### Create a Query

1. Go to **Explore** (compass icon)
2. Select **PostgreSQL** data source
3. Run this query:

```sql
SELECT 
  id,
  currency,
  balance_cents,
  status,
  updated_at
FROM accounts
ORDER BY updated_at DESC
```

### Create a Dashboard Panel

1. Go to **Dashboards** → **New Dashboard**
2. Click **Add visualization**
3. Select **PostgreSQL** data source
4. Use this query:

```sql
SELECT 
  currency,
  SUM(balance_cents) as total_balance
FROM accounts
WHERE status = 'ACTIVE'
GROUP BY currency
```

5. Choose visualization type (e.g., **Stat** or **Bar chart**)
6. Save the dashboard

## Option 3: Use the API (Simplest)

Just query the API directly:

```bash
# Get account balance
curl http://localhost:8080/v1/accounts/ACCOUNT_ID \
  -H "X-API-Key: demo-api-key-12345"
```

The `balance_cents` field shows the current balance.

## Quick Test

After rebuilding, create a transaction and watch the balance update in Grafana:

1. Create account and transaction (balance will be exported as metric)
2. Go to Grafana → Account Balance Dashboard
3. You should see the balance appear!

## Troubleshooting

**Don't see the dashboard?**
- Make sure you rebuilt the API: `cd infra && docker compose build api`
- Check if metrics are being exported: http://localhost:8080/metrics (search for `account_balance_cents`)

**Metrics not updating?**
- The metrics update when you GET an account (they're lazy-loaded)
- Or they update when accounts are created
- For real-time updates, you'd need to add metrics to the worker when it updates balances

**Want to see all accounts?**
- Use the PostgreSQL data source option (Option 2) for direct database queries

