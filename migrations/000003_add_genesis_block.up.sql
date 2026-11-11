INSERT INTO blocks (hash, height, prev_hash, root_hash, nonce, timestamp) VALUES (X'5fa15835633afa56c4f4e17bfc9c4379bc4028d93d75bbda11d941a09553509a', 0, X'0000000000000000000000000000000000000000000000000000000000000000', X'89e3c48071f1f881ea5a5a6c0c80a2e7e43d70ac3b9be082fcab2f77c21eba0d', 0, 0);

INSERT INTO transactions (hash, block_hash, coinbase_height, position) VALUES (X'89e3c48071f1f881ea5a5a6c0c80a2e7e43d70ac3b9be082fcab2f77c21eba0d', X'5fa15835633afa56c4f4e17bfc9c4379bc4028d93d75bbda11d941a09553509a', 0, 0);

INSERT INTO outputs (tx_id, out_index, amount, owner_key) VALUES (X'89e3c48071f1f881ea5a5a6c0c80a2e7e43d70ac3b9be082fcab2f77c21eba0d', 0, 42, X'04599f000e9651decab680bde41fe47c30c9c68c81c2258be6fddd04a1fb2f0a16bd57d3bdeecd63c3b0c5adae96470628015c0fc5f0c6ab8471be304f0ac58ee0');
