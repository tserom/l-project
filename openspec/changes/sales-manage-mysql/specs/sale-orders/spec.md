## ADDED Requirements

### Requirement: Persist sales orders in MySQL via sales-manage

The system SHALL persist sales-front sales orders in `sales-manage` MySQL as header and line tables, and expose nested JSON matching the existing frontend `SalesOrder` shape (string UUID `id`, nested `lines`, recomputed `totalQuantity` / `totalAmount`).

#### Scenario: Create sales order

- **WHEN** client `POST /api/v1/sale/orders` with valid customer name and at least one line
- **THEN** the service stores the order, assigns string ids, recomputes line amounts and totals, and returns the nested order

#### Scenario: Reject invalid create

- **WHEN** client creates an order with empty customer name, no lines, or negative quantity/unit price
- **THEN** the service rejects the request with a clear error (HTTP 4xx)

### Requirement: Sales order CRUD routes under /api/v1/sale/orders

The system SHALL provide `GET` list, `GET` by id, `POST` create, `PUT` update, and `DELETE` remove under `/api/v1/sale/orders`, and MUST NOT reuse stock-manage `/api/v1/sales-orders`.

#### Scenario: Get by id

- **WHEN** client `GET /api/v1/sale/orders/:id` for an existing id
- **THEN** the service returns the order with lines

#### Scenario: Missing order

- **WHEN** client gets or updates an unknown id
- **THEN** the service responds not found with message equivalent to 「销售单不存在」

### Requirement: Preserve invoiceDocId on update

When updating a sales order, the service SHALL preserve each line's existing `invoiceDocId` when the request omits it, so editing does not detach invoicing links.

#### Scenario: Update without invoiceDocId in payload

- **WHEN** client updates an order whose line already has `invoiceDocId` and the payload line omits that field
- **THEN** the stored line still has the previous `invoiceDocId`

### Requirement: Block delete when invoiced

The service SHALL refuse to delete a sales order if any line has a non-empty `invoiceDocId`.

#### Scenario: Delete order with invoiced line

- **WHEN** client deletes an order that has at least one line linked to an invoice doc
- **THEN** the service rejects the delete with a clear error and the order remains

#### Scenario: Delete order without invoice links

- **WHEN** client deletes an order with no line `invoiceDocId`
- **THEN** the order and its lines are removed

### Requirement: List sales orders with qp predicates and pagination

The service SHALL filter list results using whitelisted `qp-<field>-<operator>` query params and paginate with `page` / `pageSize` (not as `qp-*`). Unknown field or operator MUST return HTTP 400. Default sort MUST be `updatedAt` descending.

#### Scenario: Filter by customer and date range

- **WHEN** client lists with `qp-customerName-like` and `qp-orderDate-gte` / `qp-orderDate-lte`
- **THEN** only matching orders are returned with `list`, `total`, `page`, `pageSize`

#### Scenario: Illegal qp field

- **WHEN** client lists with an unknown `qp-*` field
- **THEN** the service responds HTTP 400
