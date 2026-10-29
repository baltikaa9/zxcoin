DROP TABLE utxos;

CREATE TABLE blocks (
    hash BLOB NOT NULL CHECK (length(hash) = 32),
    height INTEGER NOT NULL,
    prev_hash BLOB NOT NULL CHECK (length(prev_hash) = 32),
    root_hash BLOB NOT NULL CHECK (length(root_hash) = 32),
    nonce INTEGER NOT NULL,
    timestamp  INTEGER NOT NULL,
    PRIMARY KEY (hash)
);

CREATE INDEX idx_block_height ON blocks (height);

CREATE TABLE transactions (
    hash BLOB NOT NULL CHECK (length(hash) = 32),
    block_hash BLOB NOT NULL CHECK (length(block_hash) = 32) REFERENCES blocks (hash),
    position INTEGER NOT NULL,
    PRIMARY KEY (hash)
);

CREATE TABLE outputs (
    tx_id BLOB NOT NULL CHECK (length(tx_id) = 32) REFERENCES transactions (hash),
    out_index INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    owner_key BLOB NOT NULL CHECK (length(owner_key) = 65),
    PRIMARY KEY (tx_id, out_index)
);

CREATE TABLE inputs (
    tx_id BLOB NOT NULL CHECK (length(tx_id) = 32),
    out_index INTEGER NOT NULL,
    signature BLOB NOT NULL CHECK (length(signature) = 64),
    tx_hash BLOB NOT NULL REFERENCES transactions (hash),
    position INTEGER NOT NULL,
    PRIMARY KEY (tx_id, out_index),
    FOREIGN KEY (tx_id, out_index) REFERENCES outputs (tx_id, out_index)
);
