CREATE TABLE utxos (
    tx_id BLOB NOT NULL CHECK (length(tx_id) = 32),
    out_index INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    owner_key BLOB NOT NULL,

    PRIMARY KEY (tx_id, out_index)
);
