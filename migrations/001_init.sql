-- skema Sistem Pengadaan, Penerimaan, Penjualan + Kartu Stok
-- Semua nilai uang = BIGINT (rupiah, tanpa desimal)
-- Persen margin = INT

CREATE TABLE IF NOT EXISTS roles (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL UNIQUE            -- admin, gudang, kasir
);

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    role_id       INT NOT NULL REFERENCES roles(id),
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS satuan (
    id          SERIAL PRIMARY KEY,
    nama_satuan VARCHAR(30) NOT NULL UNIQUE,    -- pcs, box, kg
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS barang (
    id         SERIAL PRIMARY KEY,
    kode       VARCHAR(30)  NOT NULL UNIQUE,
    nama       VARCHAR(100) NOT NULL,
    satuan_id  INT NOT NULL REFERENCES satuan(id),
    harga      BIGINT NOT NULL CHECK (harga >= 0),  -- harga beli
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vendor (
    id         SERIAL PRIMARY KEY,
    nama       VARCHAR(100) NOT NULL,
    alamat     TEXT,
    telepon    VARCHAR(20),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS margin_penjualan (
    id         SERIAL PRIMARY KEY,
    persen     INT NOT NULL CHECK (persen >= 0 AND persen <= 1000),
    is_active  BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- hanya 1 margin aktif
CREATE UNIQUE INDEX IF NOT EXISTS uq_margin_aktif ON margin_penjualan (is_active) WHERE is_active = TRUE;

-- PENGADAAN
CREATE TABLE IF NOT EXISTS pengadaan (
    id         SERIAL PRIMARY KEY,
    vendor_id  INT NOT NULL REFERENCES vendor(id),
    user_id    INT NOT NULL REFERENCES users(id),
    tanggal    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status     VARCHAR(20) NOT NULL DEFAULT 'diproses' CHECK (status IN ('diproses','selesai')),
    subtotal   BIGINT NOT NULL DEFAULT 0,
    ppn        BIGINT NOT NULL DEFAULT 0,       -- 10% dari subtotal
    total      BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS detail_pengadaan (
    id           SERIAL PRIMARY KEY,
    pengadaan_id INT NOT NULL REFERENCES pengadaan(id) ON DELETE CASCADE,
    barang_id    INT NOT NULL REFERENCES barang(id),
    jumlah       INT NOT NULL CHECK (jumlah > 0),
    harga        BIGINT NOT NULL CHECK (harga >= 0),  -- snapshot barang.harga
    subtotal     BIGINT NOT NULL,
    UNIQUE (pengadaan_id, barang_id)
);

-- PENERIMAAN
CREATE TABLE IF NOT EXISTS penerimaan (
    id           SERIAL PRIMARY KEY,
    pengadaan_id INT NOT NULL REFERENCES pengadaan(id),
    user_id      INT NOT NULL REFERENCES users(id),
    tanggal      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS detail_penerimaan (
    id            SERIAL PRIMARY KEY,
    penerimaan_id INT NOT NULL REFERENCES penerimaan(id) ON DELETE CASCADE,
    barang_id     INT NOT NULL REFERENCES barang(id),
    jumlah_terima INT NOT NULL CHECK (jumlah_terima > 0),
    harga         BIGINT NOT NULL CHECK (harga >= 0)
);

-- PENJUALAN
CREATE TABLE IF NOT EXISTS penjualan (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id),           -- kasir
    margin_id  INT NOT NULL REFERENCES margin_penjualan(id),
    tanggal    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status     VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','final')),
    total      BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS detail_penjualan (
    id           SERIAL PRIMARY KEY,
    penjualan_id INT NOT NULL REFERENCES penjualan(id) ON DELETE CASCADE,
    barang_id    INT NOT NULL REFERENCES barang(id),
    jumlah       INT NOT NULL CHECK (jumlah > 0),
    harga_jual   BIGINT NOT NULL CHECK (harga_jual >= 0),   -- harga + margin
    subtotal     BIGINT NOT NULL,
    UNIQUE (penjualan_id, barang_id)
);

-- KARTU STOK
CREATE TABLE IF NOT EXISTS kartu_stok (
    id         SERIAL PRIMARY KEY,
    barang_id  INT NOT NULL REFERENCES barang(id),
    jenis      VARCHAR(10) NOT NULL CHECK (jenis IN ('masuk','keluar')),
    jumlah     INT NOT NULL CHECK (jumlah > 0),
    saldo      INT NOT NULL CHECK (saldo >= 0),             -- saldo SETELAH mutasi
    ref_tipe   VARCHAR(20) NOT NULL CHECK (ref_tipe IN ('penerimaan','penjualan')),
    ref_id     INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_kartu_stok_barang ON kartu_stok (barang_id, id DESC);

-- VIEW: saldo terakhir per barang
CREATE OR REPLACE VIEW v_stok_barang AS
SELECT b.id AS barang_id, b.kode, b.nama, s.nama_satuan, b.harga,
       COALESCE((SELECT k.saldo FROM kartu_stok k
                 WHERE k.barang_id = b.id ORDER BY k.id DESC LIMIT 1), 0) AS stok
FROM barang b JOIN satuan s ON s.id = b.satuan_id;

-- STORED PROCEDURE: finalisasi penjualan
CREATE OR REPLACE PROCEDURE sp_finalisasi_penjualan(p_penjualan_id INT)
LANGUAGE plpgsql AS $$
DECLARE
    v_status VARCHAR(20);
    r        RECORD;
    v_saldo  INT;
BEGIN
    SELECT status INTO v_status FROM penjualan WHERE id = p_penjualan_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'penjualan tidak ditemukan';
    END IF;
    IF v_status <> 'draft' THEN
        RAISE EXCEPTION 'penjualan sudah final';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM detail_penjualan WHERE penjualan_id = p_penjualan_id) THEN
        RAISE EXCEPTION 'penjualan belum memiliki detail';
    END IF;

    FOR r IN SELECT barang_id, jumlah FROM detail_penjualan
             WHERE penjualan_id = p_penjualan_id ORDER BY barang_id LOOP
        PERFORM 1 FROM barang WHERE id = r.barang_id FOR UPDATE;   -- kunci barang
        SELECT COALESCE((SELECT saldo FROM kartu_stok WHERE barang_id = r.barang_id
                         ORDER BY id DESC LIMIT 1), 0) INTO v_saldo;
        IF v_saldo < r.jumlah THEN
            RAISE EXCEPTION 'stok barang id % tidak cukup (sisa %)', r.barang_id, v_saldo;
        END IF;
        INSERT INTO kartu_stok (barang_id, jenis, jumlah, saldo, ref_tipe, ref_id)
        VALUES (r.barang_id, 'keluar', r.jumlah, v_saldo - r.jumlah, 'penjualan', p_penjualan_id);
    END LOOP;

    UPDATE penjualan SET status = 'final' WHERE id = p_penjualan_id;
END;
$$;

-- SEED (user admin dibuat lewat Go saat startup)
INSERT INTO roles (name) VALUES ('admin'), ('gudang'), ('kasir') ON CONFLICT DO NOTHING;
INSERT INTO satuan (nama_satuan) VALUES ('pcs'), ('box'), ('kg') ON CONFLICT DO NOTHING;
