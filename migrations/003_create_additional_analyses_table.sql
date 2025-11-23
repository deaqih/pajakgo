-- Create additional_analyses table
CREATE TABLE IF NOT EXISTS additional_analyses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL,
    analysis_type VARCHAR(100) NOT NULL,
    analysis_title VARCHAR(255) NOT NULL,
    analysis_content TEXT NOT NULL,
    category VARCHAR(50) DEFAULT 'manual',
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'active',
    notes TEXT,
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    INDEX idx_account_id (account_id),
    INDEX idx_analysis_type (analysis_type),
    INDEX idx_category (category),
    INDEX idx_priority (priority),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);