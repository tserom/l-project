## ADDED Requirements

### Requirement: Persist invoice docs in MySQL

The system SHALL persist invoice documents in `sales-manage` MySQL as header and line tables, returning nested JSON matching frontend `InvoiceDoc` (`status` `saved` | `voided`, snapshot fields on lines).

#### Scenario: List and get invoice docs

- **WHEN** client lists or gets invoice docs under `/api/v1/sale/invoice-docs`
- **THEN** the service returns nested documents with lines and supports `qp-status-eq` / `qp-invoiceNo-like` plus `page` / `pageSize`

### Requirement: Invoice candidates query

The system SHALL expose `GET /api/v1/sale/orders/invoice-candidates` (registered before `/:id`) that returns flattened candidate lines where the sales line has `needInvoice` true and no `invoiceDocId`, optionally filtered by the same sales-order `qp-*` whitelist (customer, dates, order numbers).

#### Scenario: Only open need-invoice lines

- **WHEN** client requests invoice candidates
- **THEN** lines already linked to an invoice doc or with `needInvoice` false are excluded

### Requirement: Create invoice in one transaction

The service SHALL create an invoice document and attach `invoiceDocId` on each referenced sales line in a single database transaction. It MUST reject if any target line is missing, not need-invoice, or already invoiced.

#### Scenario: Successful create

- **WHEN** client `POST /api/v1/sale/invoice-docs` with one or more valid candidate lines
- **THEN** the invoice is stored with status `saved`, an invoice number is assigned, and each sales line gains that invoice id

#### Scenario: Conflict on already invoiced line

- **WHEN** client creates an invoice referencing a line that already has `invoiceDocId`
- **THEN** the service rejects the request, no partial invoice is committed, and the error indicates reload (equivalent to 「明细已开票，请重新加载」)

### Requirement: Void invoice in one transaction

The service SHALL void an invoice via `POST /api/v1/sale/invoice-docs/:id/void` in one transaction: clear `invoiceDocId` on sales lines that still point to this invoice, set status `voided` and `voidedAt`. Already voided invoices MUST be rejected.

#### Scenario: Successful void

- **WHEN** client voids a `saved` invoice
- **THEN** status becomes `voided`, linked sales lines lose that `invoiceDocId`, and those lines can appear again as candidates

#### Scenario: Double void

- **WHEN** client voids an already `voided` invoice
- **THEN** the service rejects with message equivalent to 「开票单已作废」
