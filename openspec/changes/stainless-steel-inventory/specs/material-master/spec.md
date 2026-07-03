## ADDED Requirements

### Requirement: Material master record

The system SHALL store material master data in stock-center with fields: materialCode, grade, form, spec, primaryUnit, materialType, status, and optional orgId.

#### Scenario: Create raw material

- **WHEN** user creates a material with materialType `raw` and primaryUnit `kg`
- **THEN** stock-center persists the record and returns a unique id

#### Scenario: Duplicate material code rejected

- **WHEN** user creates a material with an existing materialCode
- **THEN** stock-center returns HTTP 400

### Requirement: Heat number batch

The system SHALL store batches linked to a material with heatNo unique per materialId within orgId.

#### Scenario: Create batch for material

- **WHEN** user creates a batch with materialId and heatNo
- **THEN** stock-center persists the batch and returns batch id

#### Scenario: Duplicate heat number rejected

- **WHEN** user creates a batch with duplicate materialId and heatNo
- **THEN** stock-center returns HTTP 400

### Requirement: Material list filtering

The system SHALL support list filtering via `qp-<field>-<operator>` for grade, form, materialType, and status.

#### Scenario: Filter by grade

- **WHEN** client requests `GET /api/v1/materials?qp-grade-eq=304`
- **THEN** only materials with grade 304 are returned
