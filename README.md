# Event-Driven Ledger with Zero-Knowledge Proofs (Go)

## 📌 Overview

This project is a **mini event-driven ledger system** ("bank core ledger") built in **Go**, with a **Zero-Knowledge Proof (ZKP) add-on** that allows users to prove account ownership without revealing their secrets.

It demonstrates:

* ✅ Event-driven architecture using **NATS** (publish/subscribe)
* ✅ Strong consistency via **idempotency keys**
* ✅ **Balance enforcement** (no overdrafts)
* ✅ **Zero-Knowledge ownership proofs** using **gnark (Groth16 + MiMC)**
* ✅ A modular, production-style Go project layout

This is similar to how modern fintech systems (like Monzo, Revolut) handle ledgers, but extended with a cryptographic twist.

---

## 🏗️ Architecture

```
                        ┌────────────────────┐
                        │   ZKP Service      │
                        │  (cmd/zkp)         │
                        │ - /commit          │
                        │ - /prove           │
                        │ - /verify          │
                        └─────────┬──────────┘
                                  │
                                  ▼
                          (public_hash)

┌──────────────┐    publishes    ┌──────────────┐   writes to   ┌───────────────┐
│   API        │ ──────────────► │   Worker     │ ────────────► │   SQLite DB   │
│  (cmd/api)   │                 │  (cmd/worker)│               │  ledger.db    │
│ - /accounts  │                 │              │               │               │
│ - /txns      │                 │              │               │ accounts      │
│ - /balances  │                 │              │               │ ledger_entries│
└──────────────┘                 └──────────────┘               └───────────────┘
       ▲
       │
       │ REST calls
       │
    [Client]
```

---

## ✨ Features

* **Event-driven ledger**: API publishes `tx.created` events, Worker consumes and applies them.
* **Database migrations**: auto-applied on startup (SQLite file `ledger.db`).
* **Balance safety**: prevents overdrafts by validating before applying debit.
* **Idempotency**: retries won’t double-apply a transaction.
* **Zero-Knowledge Proofs**:

  * `/commit` → generates a public hash from a secret.
  * `/prove` → creates a ZK proof that you know the secret.
  * `/verify` → verifies the proof against the stored hash.

---

## 🛠️ Tech Stack

* **Go 1.22+**
* **NATS** (event bus, lightweight Kafka alternative)
* **SQLite (modernc.org/sqlite)** for storage
* **Fiber** (fast HTTP framework)
* **gnark + gnark-crypto** (Zero-Knowledge Proofs, Groth16, MiMC)

---

## ⚡ Quickstart

### 1. Clone repo

```bash
git clone https://github.com/ruona125/ledger-zkp.git
cd ledger-zkp
go mod tidy
```

### 2. Start NATS broker

```bash
docker compose up -d
```

This runs NATS on:

* `4222` (client port)
* `8222` (management UI → [http://localhost:8222](http://localhost:8222))

---

### 3. Run the Worker (ledger writer)

```bash
go run ./cmd/worker
```

This listens to NATS for new transactions and applies them into `ledger.db`.

---

### 4. Run the API (ledger API)

In a second terminal:

```bash
go run ./cmd/api
```

Exposes endpoints:

* `POST /accounts` → create account
* `POST /transactions` → publish transaction
* `GET /balances/:id` → query account balance

---

### 5. Run the ZKP Service

In another terminal:

```bash
go run ./cmd/zkp
```

Exposes endpoints:

* `POST /commit`
* `POST /prove`
* `POST /verify`

---

## 🧪 Usage Examples

### Step 1. Generate a commitment

```bash
curl -X POST localhost:8082/commit \
  -H "Content-Type: application/json" \
  -d '{"secret":"123456"}'
```

Response:

```json
{"public_hash":"19183260921837..."}
```

Save the `public_hash`.

---

### Step 2. Create an account

```bash
curl -X POST localhost:8081/accounts \
  -H "Content-Type: application/json" \
  -d '{"name":"Ruona","public_hash":"19183260921837..."}'
```

Response:

```json
{"id":"8a2f9c4b2f134eddbf6e08b18302db9b","public_hash":"19183260921837..."}
```

---

### Step 3. Add money

```bash
curl -X POST localhost:8081/transactions \
  -H "Content-Type: application/json" \
  -d '{"account_id":"8a2f9c4b2f134eddbf6e08b18302db9b","amount":1000,"idempotency_key":"k1"}'
```

---

### Step 4. Check balance

```bash
curl localhost:8081/balances/8a2f9c4b2f134eddbf6e08b18302db9b
```

Response:

```json
{"account_id":"8a2f9c4b2f134eddbf6e08b18302db9b","balance":1000}
```

---

### Step 5. Generate Proof

```bash
curl -X POST localhost:8082/prove \
  -H "Content-Type: application/json" \
  -d '{"secret":"123456","public_hash":"19183260921837..."}'
```

Response:

```json
{"proof_b64":"...","vk_b64":"...","public":"19183260921837..."}
```

---

### Step 6. Verify Proof

```bash
curl -X POST localhost:8082/verify \
  -H "Content-Type: application/json" \
  -d '{"proof_b64":"...","vk_b64":"...","public_hash":"19183260921837..."}'
```

Response:

```json
{"valid":true}
```

---

## ⚠️ Notes

* The ledger DB is stored in `ledger.db` (auto-created). Delete the file to reset.
* NATS must be running for API/Worker to communicate.
* `secret` should be a decimal integer < BN254 field size (safe if < 2^200).
