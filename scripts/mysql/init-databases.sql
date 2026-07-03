-- Initialize databases for stainless steel inventory services.
-- Run locally (requires MySQL client; not required in CI):
--   mysql -u root -p < scripts/mysql/init-databases.sql

CREATE DATABASE IF NOT EXISTS stock_center DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS stock_manage DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
