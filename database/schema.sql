-- Accounting Web Database Schema
-- MySQL 8.0+

CREATE DATABASE IF NOT EXISTS accounting_db
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE accounting_db;

-- ===================================
-- USERS TABLE
-- ===================================
CREATE TABLE users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('admin', 'user') NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_role (role),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- ACCOUNTS TABLE
-- ===================================
CREATE TABLE accounts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account_code VARCHAR(50) NOT NULL UNIQUE,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(100),
    nature VARCHAR(50) COMMENT 'Asset, Liability, Equity, Revenue, Expense',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_account_code (account_code),
    INDEX idx_nature (nature),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- UPLOAD SESSIONS TABLE
-- ===================================
CREATE TABLE upload_sessions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_code VARCHAR(50) NOT NULL UNIQUE,
    user_id INT UNSIGNED NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500),
    total_rows INT UNSIGNED NOT NULL DEFAULT 0,
    processed_rows INT UNSIGNED NOT NULL DEFAULT 0,
    failed_rows INT UNSIGNED NOT NULL DEFAULT 0,
    status ENUM('uploaded', 'processing', 'completed', 'failed') NOT NULL DEFAULT 'uploaded',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_session_code (session_code),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- TRANSACTION DATA TABLE
-- ===================================
CREATE TABLE transaction_data (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id INT UNSIGNED NOT NULL,

    -- Input Fields
    document_type VARCHAR(100),
    document_number VARCHAR(100),
    posting_date DATE,
    account VARCHAR(50),
    account_name VARCHAR(255),
    keterangan TEXT,
    debet DECIMAL(20, 2) DEFAULT 0.00,
    credit DECIMAL(20, 2) DEFAULT 0.00,
    net DECIMAL(20, 2) DEFAULT 0.00,

    -- Output Fields (Auto-filled)
    analisa_nature_akun VARCHAR(100),
    analisa_koreksi_obyek VARCHAR(255),
    koreksi VARCHAR(255),
    obyek VARCHAR(255),
    um_pajak_db DECIMAL(20, 2) DEFAULT 0.00,
    pm_db DECIMAL(20, 2) DEFAULT 0.00,
    wth_21_cr DECIMAL(20, 2) DEFAULT 0.00,
    wth_23_cr DECIMAL(20, 2) DEFAULT 0.00,
    wth_26_cr DECIMAL(20, 2) DEFAULT 0.00,
    wth_4_2_cr DECIMAL(20, 2) DEFAULT 0.00,
    wth_15_cr DECIMAL(20, 2) DEFAULT 0.00,
    pk_cr DECIMAL(20, 2) DEFAULT 0.00,
    analisa_tambahan TEXT,

    -- Processing status
    is_processed BOOLEAN NOT NULL DEFAULT FALSE,
    processing_error TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (session_id) REFERENCES upload_sessions(id) ON DELETE CASCADE,
    INDEX idx_session_id (session_id),
    INDEX idx_account (account),
    INDEX idx_posting_date (posting_date),
    INDEX idx_is_processed (is_processed)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- KOREKSI RULES TABLE
-- ===================================
CREATE TABLE koreksi_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    priority INT NOT NULL DEFAULT 0 COMMENT 'Higher priority = checked first',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_keyword (keyword),
    INDEX idx_priority (priority DESC),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- OBYEK RULES TABLE
-- ===================================
CREATE TABLE obyek_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    priority INT NOT NULL DEFAULT 0 COMMENT 'Higher priority = checked first',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_keyword (keyword),
    INDEX idx_priority (priority DESC),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- WITHHOLDING TAX RULES TABLE
-- ===================================
CREATE TABLE withholding_tax_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    tax_type ENUM('wth_21', 'wth_23', 'wth_26', 'wth_4_2', 'wth_15') NOT NULL,
    tax_rate DECIMAL(5, 4) NOT NULL COMMENT 'Example: 0.0200 for 2%',
    priority INT NOT NULL DEFAULT 0 COMMENT 'Higher priority = checked first',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_keyword (keyword),
    INDEX idx_tax_type (tax_type),
    INDEX idx_priority (priority DESC),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- TAX KEYWORDS TABLE
-- ===================================
CREATE TABLE tax_keywords (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    tax_category ENUM('input_tax', 'output_tax') NOT NULL,
    priority INT NOT NULL DEFAULT 0 COMMENT 'Higher priority = checked first',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_keyword (keyword),
    INDEX idx_tax_category (tax_category),
    INDEX idx_priority (priority DESC),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===================================
-- SEED DATA
-- ===================================

-- Default admin user (password: admin123)
INSERT INTO users (username, email, password_hash, role) VALUES
('admin', 'admin@accounting.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYSqVzGZE9G', 'admin');

-- Sample accounts
INSERT INTO accounts (account_code, account_name, account_type, nature) VALUES
('1010101', 'Cash on Hand', 'Current Asset', 'Asset'),
('1010201', 'Bank Account', 'Current Asset', 'Asset'),
('2010101', 'Accounts Payable', 'Current Liability', 'Liability'),
('4010101', 'Sales Revenue', 'Revenue', 'Revenue'),
('5010101', 'Cost of Goods Sold', 'Expense', 'Expense'),
('5020101', 'Entertainment Expense', 'Operating Expense', 'Expense');

-- Sample koreksi rules
INSERT INTO koreksi_rules (keyword, value, priority) VALUES
('entertain', 'Biaya Entertainment', 100),
('entertainment', 'Biaya Entertainment', 100),
('sewa', 'Biaya Sewa', 90),
('gaji', 'Biaya Gaji', 90),
('listrik', 'Biaya Listrik', 80),
('telepon', 'Biaya Telepon', 80);

-- Sample obyek rules
INSERT INTO obyek_rules (keyword, value, priority) VALUES
('hotel', 'Hotel', 100),
('restaurant', 'Restaurant', 100),
('restoran', 'Restaurant', 100),
('office', 'Office', 90),
('kantor', 'Office', 90),
('gedung', 'Gedung', 90);

-- Sample withholding tax rules
INSERT INTO withholding_tax_rules (keyword, tax_type, tax_rate, priority) VALUES
('PPh 21', 'wth_21', 0.0500, 100),
('PPH 21', 'wth_21', 0.0500, 100),
('PPh 23', 'wth_23', 0.0200, 100),
('PPH 23', 'wth_23', 0.0200, 100),
('PPh 26', 'wth_26', 0.2000, 100),
('PPH 26', 'wth_26', 0.2000, 100),
('PPh 4(2)', 'wth_4_2', 0.1000, 100),
('PPH 4(2)', 'wth_4_2', 0.1000, 100),
('PPh 15', 'wth_15', 0.0200, 100),
('PPH 15', 'wth_15', 0.0200, 100);

-- Sample tax keywords
INSERT INTO tax_keywords (keyword, tax_category, priority) VALUES
('Input Tax', 'input_tax', 100),
('INPUT TAX', 'input_tax', 100),
('PPN Masukan', 'input_tax', 100),
('Output Tax', 'output_tax', 100),
('OUTPUT TAX', 'output_tax', 100),
('PPN Keluaran', 'output_tax', 100);
