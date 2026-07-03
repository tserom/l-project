## ADDED Requirements

### Requirement: Document draft and confirm lifecycle

The system SHALL store business documents in stock-manage with status `draft` or `confirmed`, and only `confirmed` documents trigger stock-center changes.

#### Scenario: Draft document editable

- **WHEN** document status is draft
- **THEN** user can update or delete the document without stock changes

#### Scenario: Confirmed document locked

- **WHEN** user confirms a document
- **THEN** stock-manage calls stock-center and sets status to confirmed; further edits are rejected

### Requirement: Inbound order

The system SHALL support inbound orders for purchase and manual finished-goods receipt with line items referencing materialId, batchId, warehouse, weightKg, and quantity.

#### Scenario: Confirm inbound order

- **WHEN** user confirms an inbound order with valid lines
- **THEN** stock-manage calls stock-center inbound for each line with refType `inbound` and the document docNo

### Requirement: Outbound order

The system SHALL support outbound orders for processing pick and other issues with the same line structure as inbound.

#### Scenario: Confirm outbound order

- **WHEN** user confirms an outbound order with sufficient stock
- **THEN** stock-manage calls stock-center outbound for each line with refType `outbound`

### Requirement: Sales order without stock deduction

The system SHALL create sales orders with customerName and lines without changing stock until shipment is confirmed.

#### Scenario: Confirm sales order only

- **WHEN** user confirms a sales order
- **THEN** document status becomes confirmed and stock balance is unchanged

### Requirement: Sales shipment from sales order

The system SHALL allow creating sales shipment from a confirmed sales order and deduct stock on shipment confirm.

#### Scenario: Confirm sales shipment

- **WHEN** user confirms a sales shipment linked to a sales order
- **THEN** stock-manage calls stock-center outbound with refType `sale` for each shipment line

### Requirement: Processing order with loss

The system SHALL support processing orders with pick lines (raw weightKg outbound) and finish lines (quantity/optional weightKg inbound) and store lossWeightKg as sum(pick weight) minus sum(finish weight).

#### Scenario: Confirm processing order

- **WHEN** user confirms a processing order with pick and finish lines
- **THEN** stock-manage outbound pick lines, inbound finish lines, and persists lossWeightKg on the order

#### Scenario: Finish line without weight

- **WHEN** finish lines have quantity but no weightKg
- **THEN** lossWeightKg equals total pick weightKg minus zero finish weight

### Requirement: Operation audit log

The system SHALL write operation logs for document create, confirm, and stock-center call failures with doc type, docNo, operator, and action.

#### Scenario: Log on confirm

- **WHEN** any document is confirmed successfully
- **THEN** stock-manage appends an operation log entry
