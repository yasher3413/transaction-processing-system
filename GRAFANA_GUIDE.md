# How to Use Grafana (The Dashboard)

## 🚪 How to Log In

1. **Open your web browser** (Chrome, Safari, Firefox, etc.)

2. **Go to this address:**
   ```
   http://localhost:3000
   ```

3. **Login with:**
   - **Username**: `admin`
   - **Password**: `admin`

4. **Click "Log in"**

5. **If it asks you to change password:**
   - You can click "Skip" for now (or set a new one if you want)

## 🎨 What You'll See

Once you're in, you'll see:
- A dashboard with graphs
- Metrics about your system (requests, processing time, etc.)
- You can explore different views

## 📊 What the Graphs Show

- **API Requests**: How many requests came in
- **Processing Time**: How fast transactions are being processed
- **Errors**: If anything went wrong
- **Worker Activity**: How busy the worker is

## 💡 Important Notes

- **Grafana is just for viewing** - it doesn't control anything
- **You don't need Grafana** for the system to work
- It's like a "monitor" or "TV screen" showing you stats
- If you can't access it, that's okay - the system still works!

## 🔧 If Grafana Won't Load

1. Make sure Docker is running: `docker ps`
2. Check if Grafana container is running: `docker logs transactions-grafana`
3. Try refreshing the page
4. If it still doesn't work, you can skip it - it's optional!

## 🎯 What to Do Instead

If Grafana is confusing, just use the API directly:

```bash
# Create an account
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{"currency": "USD"}'

# Create a transaction
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo-api-key-12345" \
  -d '{
    "account_id": "YOUR_ACCOUNT_ID",
    "amount_cents": 10000,
    "currency": "USD",
    "type": "CREDIT",
    "idempotency_key": "test-123"
  }'

# Check balance
curl http://localhost:8080/v1/accounts/YOUR_ACCOUNT_ID \
  -H "X-API-Key: demo-api-key-12345"
```

That's all you really need! Grafana is just bonus visualization.

