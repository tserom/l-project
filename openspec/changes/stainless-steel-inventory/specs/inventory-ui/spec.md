## ADDED Requirements

### Requirement: Browser only calls stock-manage

The frontend SHALL send all API requests to stock-manage base URL only.

#### Scenario: No direct center URL in frontend

- **WHEN** frontend code is built for production
- **THEN** no environment variable or constant points to stock-center port

### Requirement: Inventory UI pages

The frontend SHALL provide pages for material list, stock query, inbound orders, outbound orders, sales orders, and processing orders.

#### Scenario: Navigate to inbound list

- **WHEN** user opens inbound orders page
- **THEN** a paginated list of inbound documents is displayed from manage API

#### Scenario: Create and confirm inbound from UI

- **WHEN** user fills inbound form and clicks confirm
- **THEN** UI calls manage confirm API and shows success or validation error

### Requirement: Processing loss display

The processing order detail page SHALL display lossWeightKg after confirmation.

#### Scenario: View confirmed processing order

- **WHEN** user opens a confirmed processing order
- **THEN** lossWeightKg is visible on the detail view

### Requirement: Stock query filters

The stock query page SHALL filter by grade, form, and heatNo via manage-proxied qp parameters.

#### Scenario: Filter stock by heat number

- **WHEN** user enters a heatNo filter and searches
- **THEN** only matching stock rows are shown
