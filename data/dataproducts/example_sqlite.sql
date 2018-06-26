PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';
CREATE TABLE IF NOT EXISTS product_types ( product_type TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS parties (
  party_id INTEGER PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  product_type TEXT NOT NULL,
  FOREIGN KEY(product_type) REFERENCES product_types(product_type) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS current_party AS
  SELECT * FROM parties
  ORDER BY created_at DESC LIMIT 1;

CREATE VIEW IF NOT EXISTS current_party_id AS
  SELECT party_id FROM current_party;

CREATE TABLE IF NOT EXISTS products (
  party_id INTEGER NOT NULL,
  serial INTEGER NOT NULL DEFAULT 1,
  CONSTRAINT unique_serial UNIQUE (party_id, serial),
  CONSTRAINT positive_serial CHECK (serial > 0),
  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS party_param (
  param NOT NULL PRIMARY KEY,
  min_value,
  max_value,
  CHECK ( param IS NOT '' AND param >= 0)
);

CREATE TABLE IF NOT EXISTS party_value (
  party_id INTEGER NOT NULL,
  param NOT NULL,
  value NOT NULL,
  CONSTRAINT unique_param UNIQUE (party_id, param),

  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE,
  FOREIGN KEY(param) REFERENCES party_param(param) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_param (
  param NOT NULL PRIMARY KEY,
  min_value,
  max_value,
  CHECK ( param IS NOT '' AND param >= 0)
);

CREATE TABLE IF NOT EXISTS product_value (
  party_id INTEGER NOT NULL,
  serial INTEGER NOT NULL,
  param NOT NULL,
  value NOT NULL,
  CONSTRAINT unique_param UNIQUE (party_id, param),
  FOREIGN KEY(party_id, serial) REFERENCES products(party_id,serial) ON DELETE CASCADE,
  FOREIGN KEY(param) REFERENCES party_param(param) ON DELETE CASCADE
);

drop table party_value;
drop table party_param;

INSERT INTO party_param (param) VALUES ('gas1');

INSERT OR REPLACe INTO party_param (param, min_value) VALUES (0, 0);
INSERT OR REPLACe INTO party_param (param, min_value, max_value) VALUES ('22', '00', '99');
select * from party_param;

INSERT INTO product_types (product_type) VALUES ('035');
INSERT INTO parties (product_type) VALUES ('035');
INSERT OR REPLACE INTO party_value(party_id, param, value) VALUES (1, '22', 12);
INSERT OR REPLACE INTO party_value(party_id, param, value) VALUES (1, 0, -1);

delete from party_value where param = 1;

select * from party_value;

CREATE TRIGGER IF NOT EXISTS trigger_validate_party_value
  BEFORE INSERT ON party_value
  WHEN
    exists(
        SELECT * FROM party_param
        WHERE party_param.param = new.param AND
              (party_param.min_value IS NOT NULL) AND
              ( typeof(new.value) IS NOT typeof(party_param.min_value) OR
                new.value < party_param.min_value )
              OR
              party_param.param = new.param AND
              (party_param.max_value IS NOT NULL) AND
              ( typeof(new.value) IS NOT typeof(party_param.min_value) OR
                new.value > party_param.max_value
              )
    )
BEGIN
  SELECT RAISE(ABORT,'party value out of range');
END;

CREATE TRIGGER IF NOT EXISTS trigger_validate_product_value
  BEFORE INSERT ON product_value
  WHEN
    exists(
        SELECT * FROM product_param
        WHERE product_param.param = new.param AND
              (product_param.min_value IS NOT NULL) AND
              ( typeof(new.value) IS NOT typeof(product_param.min_value) OR
                new.value < product_param.min_value )
              OR
              product_param.param = new.param AND
              (product_param.max_value IS NOT NULL) AND
              ( typeof(new.value) IS NOT typeof(product_param.min_value) OR
                new.value > product_param.max_value
              )
    )
BEGIN
  SELECT RAISE(ABORT,'product value out of range');
END;


drop trigger trigger_validate_party_value;