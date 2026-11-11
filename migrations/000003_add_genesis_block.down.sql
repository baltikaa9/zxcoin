CREATE TEMP TABLE _irreversible_error (id INT);

CREATE TEMP TRIGGER _trg_irreversible
BEFORE INSERT ON _irreversible_error
BEGIN
	SELECT RAISE(ABORT, 'Irreversible migration: cannot delete genesis block');
END;

INSERT INTO _irreversible_error VALUES (1);
