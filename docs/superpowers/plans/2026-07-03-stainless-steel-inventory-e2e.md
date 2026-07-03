# Stainless Steel Inventory — E2E Verification

**Date:** 2026-07-03  
**Branch:** `feature/stainless-steel-inventory`  
**Worktree:** `.worktrees/stainless-steel-inventory`

## Automated verification (Task 13)

| Step | Command / check | Result | Notes |
|------|-----------------|--------|-------|
| Go tests — stock-center | `go test ./...` in `apps/stock-center` | **PASS** | `internal/service` tests (material duplicate, stock inbound/outbound) |
| Go tests — stock-manage | `go test ./...` in `apps/stock-manage` | **PASS** | inbound/outbound/sales/processing/docno service tests |
| Full build | `make build-all` at repo root | **PASS** | Requires Node ≥20 + pnpm (via corepack). Default shell Node v16 fails with `pnpm: command not found`. |
| Partial build (Go only) | `make build-center` | **PASS** | `bin/stock-center` produced |
| MySQL TCP | `127.0.0.1:3306` | **REACHABLE** | Port open |
| MySQL auth | `root` / `.env.example` password | **FAIL** | `Access denied for user 'root'@'localhost'` — no `.env` files present; example creds do not match local instance |
| Live HTTP flow | curl material → sales | **SKIPPED** | Blocked by MySQL credentials; services cannot migrate schema |

### Unit-test coverage for acceptance criteria 11.2

Automated (no MySQL):

- `TestOutbound_OverBalanceFails` — outbound over balance returns error
- `TestProcessingOrder_ConfirmOutboundThenInboundWithLoss` — `lossWeightKg = pick − finish`
- `TestSalesOrder_ConfirmDoesNotCallCenter` — sales confirm does not deduct stock
- `TestShipment_ConfirmCallsOutboundWithSaleRefType` — shipment confirm calls center outbound

---

## Prerequisites (manual live test)

1. **MySQL 8** running locally with credentials matching your `.env` files.
2. Copy env templates and adjust passwords:

   ```bash
   cp apps/stock-center/.env.example apps/stock-center/.env
   cp apps/stock-manage/.env.example apps/stock-manage/.env
   # Edit DB_PASSWORD (and user if needed)
   ```

3. Initialize databases:

   ```bash
   mysql -u root -p < scripts/mysql/init-databases.sql
   ```

4. Build (Node 20+ required for frontend):

   ```bash
   export PATH="$HOME/.nvm/versions/node/v20.19.5/bin:$PATH"  # adjust if needed
   make build-all
   ```

5. Start services:

   ```bash
   ./scripts/launcher/start.sh
   # stock-center → :8081, stock-manage (BFF + UI) → :8082
   ```

---

## Manual E2E checklist — 来料入库 → 加工 → 销售出库

Base URL: `http://localhost:8082/api/v1` (stock-manage BFF).  
Use `jq` for readability; all responses use `{ "code": 0, "data": … }` envelope.

### 1. Health

- [ ] `curl -sf http://localhost:8081/health`
- [ ] `curl -sf http://localhost:8082/health`

### 2. Create raw material (304 plate)

```bash
curl -s -X POST http://localhost:8082/api/v1/materials \
  -H 'Content-Type: application/json' \
  -d '{
    "materialCode": "304-PLATE-3MM",
    "grade": "304",
    "form": "plate",
    "spec": "3mm",
    "primaryUnit": "kg",
    "materialType": "raw"
  }' | jq .
```

- [ ] Record `RAW_MAT_ID` from response `data.id`

### 3. Create finished material (304 part)

```bash
curl -s -X POST http://localhost:8082/api/v1/materials \
  -H 'Content-Type: application/json' \
  -d '{
    "materialCode": "304-PART-A",
    "grade": "304",
    "form": "part",
    "spec": "cut A",
    "primaryUnit": "piece",
    "materialType": "finished"
  }' | jq .
```

- [ ] Record `FG_MAT_ID`

### 4. Create batches

```bash
# Raw batch
curl -s -X POST http://localhost:8082/api/v1/batches \
  -H 'Content-Type: application/json' \
  -d "{\"materialId\": $RAW_MAT_ID, \"heatNo\": \"H20260703001\"}" | jq .

# Finished batch
curl -s -X POST http://localhost:8082/api/v1/batches \
  -H 'Content-Type: application/json' \
  -d "{\"materialId\": $FG_MAT_ID, \"heatNo\": \"H20260703002\"}" | jq .
```

- [ ] Record `RAW_BATCH_ID`, `FG_BATCH_ID`

### 5. Inbound 100 kg raw material

```bash
curl -s -X POST http://localhost:8082/api/v1/inbound-orders \
  -H 'Content-Type: application/json' \
  -d "{
    \"operator\": \"e2e\",
    \"lines\": [{
      \"materialId\": $RAW_MAT_ID,
      \"batchId\": $RAW_BATCH_ID,
      \"warehouse\": \"WH-RM\",
      \"weightKg\": \"100\",
      \"quantity\": \"0\"
    }]
  }" | jq .
```

- [ ] Record `INBOUND_ID`; confirm:

```bash
curl -s -X POST "http://localhost:8082/api/v1/inbound-orders/$INBOUND_ID/confirm" | jq .
```

- [ ] Verify stock: `curl -s "http://localhost:8082/api/v1/stocks/query?materialId=$RAW_MAT_ID&batchId=$RAW_BATCH_ID&warehouse=WH-RM" | jq .` → weight ≈ 100 kg
- [ ] Verify ledger: `curl -s "http://localhost:8082/api/v1/ledger?materialId=$RAW_MAT_ID" | jq .`

### 6. Processing — pick 50 kg, finish 42 kg (loss 8 kg)

```bash
curl -s -X POST http://localhost:8082/api/v1/processing-orders \
  -H 'Content-Type: application/json' \
  -d "{
    \"operator\": \"e2e\",
    \"pickLines\": [{
      \"materialId\": $RAW_MAT_ID,
      \"batchId\": $RAW_BATCH_ID,
      \"warehouse\": \"WH-RM\",
      \"weightKg\": \"50\"
    }],
    \"finishLines\": [{
      \"materialId\": $FG_MAT_ID,
      \"batchId\": $FG_BATCH_ID,
      \"warehouse\": \"WH-FG\",
      \"weightKg\": \"42\",
      \"quantity\": \"10\"
    }]
  }" | jq .
```

- [ ] Record `PROC_ID`; confirm:

```bash
curl -s -X POST "http://localhost:8082/api/v1/processing-orders/$PROC_ID/confirm" | jq .
```

- [ ] Response `data.lossWeightKg` = `"8"`
- [ ] Raw stock WH-RM ≈ 50 kg; finished WH-FG ≈ 42 kg / 10 pcs

### 7. Sales order + shipment outbound

```bash
curl -s -X POST http://localhost:8082/api/v1/sales-orders \
  -H 'Content-Type: application/json' \
  -d "{
    \"customerName\": \"E2E Customer\",
    \"operator\": \"e2e\",
    \"lines\": [{
      \"materialId\": $FG_MAT_ID,
      \"batchId\": $FG_BATCH_ID,
      \"weightKg\": \"20\",
      \"quantity\": \"5\"
    }]
  }" | jq .
```

- [ ] Record `SO_ID`; confirm sales (no stock change):

```bash
curl -s -X POST "http://localhost:8082/api/v1/sales-orders/$SO_ID/confirm" | jq .
```

- [ ] Create shipment from order:

```bash
curl -s -X POST "http://localhost:8082/api/v1/sales-orders/$SO_ID/shipments" \
  -H 'Content-Type: application/json' \
  -d '{
    "operator": "e2e",
    "warehouse": "WH-FG",
    "lines": []
  }' | jq .
```

  (Empty `lines` copies all order lines; or pass explicit lines.)

- [ ] Record `SHIP_ID`; confirm shipment:

```bash
curl -s -X POST "http://localhost:8082/api/v1/sales-shipments/$SHIP_ID/confirm" | jq .
```

- [ ] Finished stock WH-FG ≈ 22 kg remaining

### 8. Negative test — outbound over balance

```bash
curl -s -X POST http://localhost:8082/api/v1/outbound-orders \
  -H 'Content-Type: application/json' \
  -d "{
    \"operator\": \"e2e\",
    \"lines\": [{
      \"materialId\": $FG_MAT_ID,
      \"batchId\": $FG_BATCH_ID,
      \"warehouse\": \"WH-FG\",
      \"weightKg\": \"9999\",
      \"quantity\": \"0\"
    }]
  }" | jq .
```

- [ ] Confirm returns HTTP 400 / business error (insufficient balance)

### 9. UI smoke (optional)

- [ ] Open `http://localhost:8082` — embedded Vue app loads
- [ ] Navigate: 物料档案 → 库存查询 → 入库单 → 出库单 → 销售单 → 加工单
- [ ] Dev proxy (separate terminal): `cd apps/stock-front && pnpm dev` → `http://localhost:8104`

### 10. Launcher

- [ ] `./scripts/launcher/start.sh` waits for both health endpoints and opens browser
- [ ] Ctrl+C stops both processes cleanly

---

## Sign-off

| Criterion | Automated | Manual |
|-----------|-----------|--------|
| 11.1 Full flow stock + ledger consistent | — | Pending |
| 11.2 lossWeightKg + over-stock error | **PASS** (unit tests) | Optional re-check live |
| 11.3 Launcher one-click start | — | Pending |

**Verified by:** Agent Task 13 (2026-07-03)
