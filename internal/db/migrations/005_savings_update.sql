ALTER TABLE savings
    MODIFY COLUMN pembayaranbunga_kredit_tgltetap DOUBLE,
    MODIFY COLUMN pembayaranbunga_debit_tgltetap DOUBLE,
    MODIFY COLUMN produk_sukubungadebit DOUBLE,
    MODIFY COLUMN nilaibukablokir DOUBLE,
    MODIFY COLUMN saldoblokir DOUBLE,
    MODIFY COLUMN plafondlimit DOUBLE,
    MODIFY COLUMN total_so_aktif DOUBLE,
    DROP COLUMN datastandingorder;

CREATE TABLE IF NOT EXISTS savings_standing_order (
    row_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    saving_norekening VARCHAR(64) NOT NULL,
    tgltransaksi TEXT,
    norektujuan TEXT,
    namanasabah TEXT,
    nominal TEXT,
    frekuensi TEXT,
    xfrekuensi BIGINT,
    tglawal TEXT,
    tglakhir TEXT,
    tglberikutnya TEXT,
    status TEXT,
    keterangan TEXT,
    fee TEXT,
    KEY idx_savings_standing_order_norek (saving_norekening)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
