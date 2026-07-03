## ADDED Requirements

### Requirement: Local MySQL dual database

The system SHALL use MySQL 8 with separate databases stock_center and stock_manage on localhost for phase 1.

#### Scenario: Auto migrate on startup

- **WHEN** stock-center or stock-manage starts with valid DB credentials
- **THEN** each service runs GORM AutoMigrate for its models

### Requirement: Launcher one-click start

The system SHALL provide a launcher script that starts stock-center, waits for health, starts stock-manage, waits for health, and opens the browser at manage port.

#### Scenario: Successful launch

- **WHEN** user runs the launcher with MySQL running and binaries present
- **THEN** both /health endpoints return OK and browser opens http://localhost:8082

#### Scenario: MySQL not available

- **WHEN** MySQL is not reachable
- **THEN** launcher prints a clear error and exits without leaving orphan processes

### Requirement: Embedded frontend in manage binary

The system SHALL serve stock-front static assets from stock-manage using go embed for production builds.

#### Scenario: SPA route fallback

- **WHEN** user navigates to a frontend route directly
- **THEN** stock-manage returns index.html for unknown non-API paths

### Requirement: Build artifacts

The system SHALL document building stock-center and stock-manage binaries and optional Windows batch or macOS command launcher.

#### Scenario: Production build

- **WHEN** developer runs documented build commands
- **THEN** dist binaries are produced and manage binary includes frontend assets
