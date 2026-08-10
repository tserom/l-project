## ADDED Requirements

### Requirement: Api layer calls sales-manage over HTTP

`sales-front` SHALL keep exported functions `listSalesOrders`, `getSalesOrder`, `createSalesOrder`, `updateSalesOrder`, `removeSalesOrder`, `listInvoiceDocs`, `getInvoiceDoc`, `listInvoiceCandidates`, `createInvoiceDoc`, and `voidInvoiceDoc`, and MUST implement them by calling `sales-manage` HTTP APIs. Pages MUST NOT import Dexie or `src/storage/db`.

#### Scenario: Create order from UI path

- **WHEN** the order edit page saves a new order through `createSalesOrder`
- **THEN** the browser sends `POST /api/v1/sale/orders` (via configured API base / proxy) and does not write IndexedDB

#### Scenario: Create invoice from UI path

- **WHEN** the invoice create page submits through `createInvoiceDoc`
- **THEN** the browser sends a single `POST /api/v1/sale/invoice-docs` and does not open a Dexie multi-table transaction

### Requirement: Encode list filters as qp query params

Order list and invoice candidate UIs SHALL send filter fields as `qp-*` query parameters (and `page` / `pageSize` when listing), instead of downloading all orders and filtering only in memory for the primary list path.

#### Scenario: Customer name filter

- **WHEN** the user filters the order list by customer name
- **THEN** `listSalesOrders` requests include `qp-customerName-like` with that value

### Requirement: Remove IndexedDB as source of truth

After the cutover, `sales-front` MUST NOT use IndexedDB as the persistence source for sales orders or invoice docs. Dexie dependencies and `fake-indexeddb` test wiring tied to repository persistence SHALL be removed or replaced with HTTP-level tests / pure-function tests.

#### Scenario: No Dexie schema in runtime path

- **WHEN** the production/dev app bundle loads order and invoice flows
- **THEN** it does not open the `sales-front` IndexedDB database for CRUD

### Requirement: Document and configure API base

The project SHALL document how to run MySQL + `sales-manage` + `sales-front`, including API base URL or Vite proxy to port `8083`.

#### Scenario: Local dev README

- **WHEN** a developer follows `sales-front` / root README for sales tooling
- **THEN** steps include starting `sales-manage` and pointing the front end at it
