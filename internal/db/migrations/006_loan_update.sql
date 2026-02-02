ALTER TABLE loans
    MODIFY COLUMN tempatpenyimpanan TEXT,
    MODIFY COLUMN channeling TEXT;

CREATE TABLE IF NOT EXISTS loan_data_jaminan_lainnya (
    row_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    loan_id VARCHAR(64) NOT NULL,
    nm_jaminanlainnya TEXT,
    nomor_jaminan TEXT,
    lainnya_nilaijaminanreal DOUBLE,
    lainnya_nilaijaminan DOUBLE,
    KEY idx_loan_data_jaminan_lainnya_loan_id (loan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
