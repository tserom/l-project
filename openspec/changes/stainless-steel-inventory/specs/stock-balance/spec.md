## ADDED Requirements

### Requirement: Dual-measure stock balance

The system SHALL maintain stock balance per materialId, batchId, and warehouse with weightKg and quantity fields.

#### Scenario: Query balance by keys

- **WHEN** client queries stock by materialId, batchId, and warehouse
- **THEN** stock-center returns current weightKg and quantity or zero values if no record exists

### Requirement: Stock inbound

The system SHALL increase weightKg and/or quantity on inbound and append an immutable ledger entry.

#### Scenario: Inbound raw material by weight

- **WHEN** inbound request specifies weightKg greater than zero
- **THEN** balance weightKg increases and ledger records positive deltaWeightKg with refType and refNo

#### Scenario: Inbound finished goods by quantity

- **WHEN** inbound request specifies quantity greater than zero
- **THEN** balance quantity increases and ledger records positive deltaQuantity

### Requirement: Stock outbound validation

The system SHALL reject outbound when requested weight or quantity exceeds available balance.

#### Scenario: Insufficient weight

- **WHEN** outbound weightKg exceeds current weightKg
- **THEN** stock-center returns HTTP 400 and does not change balance

#### Scenario: Successful outbound

- **WHEN** outbound weightKg and quantity are within balance
- **THEN** balance decreases and ledger records negative deltas

### Requirement: Stock ledger traceability

The system SHALL record ledger entries with refType, refNo, materialId, batchId, warehouse, deltaWeightKg, deltaQuantity, and createdAt.

#### Scenario: Query ledger by document

- **WHEN** client filters ledger by refNo
- **THEN** all movements for that document are returned in chronological order
