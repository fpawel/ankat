package dataproducts

import (
	"os"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func MustOpen(dbFilename string) (x DB) {

	if _, err := os.Stat(dbFilename); os.IsNotExist(err) {
		file, err := os.Create(dbFilename)
		if err == nil {
			err = file.Close()
		}
		if err != nil {
			panic(err)
		}
	}
	x.DB = sqlx.MustConnect("sqlite3", dbFilename )

	x.MustExec(`
PRAGMA foreign_keys = ON; 
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS parties (
  party_id INTEGER PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

CREATE VIEW IF NOT EXISTS parties_years AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INT) as year
  FROM parties
  ORDER BY year;

CREATE VIEW IF NOT EXISTS parties_years_months AS
  SELECT DISTINCT
    cast(strftime('%Y', created_at) AS INT) as year,
    cast(strftime('%m', created_at) AS INT) as month
  FROM parties
  ORDER BY created_at;


CREATE VIEW IF NOT EXISTS parties_years_months_days AS
  SELECT DISTINCT
    cast(strftime('%Y', created_at) AS INT) as year,
    cast(strftime('%m', created_at) AS INT) as month,
    cast(strftime('%d', created_at) AS INT) as day
  FROM parties
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS parties_years_months_days2 AS
  SELECT
    p.party_id,
    cast(strftime('%Y', created_at) AS INT) as year,
    cast(strftime('%m', created_at) AS INT) as month,
    cast(strftime('%d', created_at) AS INT) as day,
    created_at,
    value AS product_type FROM parties p
    INNER JOIN party_var_values ON p.party_id = party_var_values.party_id
  WHERE var = 'product_type'
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS party_products_count AS 
  SELECT count(distinct products.party_id) FROM products 
	INNER JOIN parties p on products.party_id = p.party_id;


CREATE VIEW IF NOT EXISTS current_party AS
  SELECT * FROM parties
  ORDER BY created_at DESC LIMIT 1;

CREATE VIEW IF NOT EXISTS current_party_id AS
  SELECT party_id FROM current_party;

CREATE VIEW IF NOT EXISTS current_products AS 
	SELECT serial FROM products WHERE party_id IN current_party_id;

CREATE TABLE IF NOT EXISTS products (
  party_id INTEGER NOT NULL,
  serial INTEGER NOT NULL,
  CONSTRAINT unique_serial UNIQUE (party_id, serial),
  CONSTRAINT positive_serial CHECK (serial > 0),
  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS party_vars (
  var NOT NULL PRIMARY KEY CHECK ( var != ''),
  name TEXT NOT NULL CHECK ( name != ''),
  sort_order INTEGER NOT NULL DEFAULT 0,
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text')), 
  min, max, def_val
);

CREATE TABLE IF NOT EXISTS product_vars_sections (
  section TEXT NOT NULL PRIMARY KEY CHECK ( section != ''),
  name TEXT NOT NULL CHECK ( name != ''),
  points_count INTEGER NOT NULL CHECK ( points_count > 0)	
);

CREATE TABLE IF NOT EXISTS product_vars (
  var NOT NULL PRIMARY KEY CHECK ( var != ''),
  name TEXT NOT NULL CHECK ( name != ''),
  sort_order INTEGER NOT NULL DEFAULT 0,
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text')), 
  min, max, def_val
);

CREATE TABLE IF NOT EXISTS var_value_list(
  var NOT NULL CHECK ( var IS NOT ''),
  value NOT NULL,
  UNIQUE (var, value)   
);


CREATE TABLE IF NOT EXISTS party_var_values (
  var NOT NULL,
  party_id INTEGER NOT NULL,
  value NOT NULL,
  UNIQUE (party_id,  var),
  FOREIGN KEY(var) REFERENCES party_vars(var) ON DELETE CASCADE,  
  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS party_var_values2 AS
  SELECT party_id, name, value FROM party_var_values
    INNER JOIN party_vars ON party_var_values.var = party_vars.var;

CREATE TABLE IF NOT EXISTS product_var_values (
  var NOT NULL,
  party_id INTEGER NOT NULL,
  serial INTEGER NOT NULL,
  value NOT NULL,
  UNIQUE (party_id, serial, var),
  FOREIGN KEY(var) REFERENCES product_vars(var) ON DELETE CASCADE,  
  FOREIGN KEY(party_id, serial) REFERENCES products(party_id, serial) ON DELETE CASCADE
);

INSERT OR IGNORE INTO parties(party_id) VALUES (1);
INSERT OR IGNORE INTO products(party_id, serial) VALUES (1,1);

CREATE TABLE IF NOT EXISTS app_config (  
  var NOT NULL PRIMARY KEY CHECK ( var != ''),
  value NOT NULL,
  name TEXT NOT NULL CHECK ( name != ''),
  section TEXT NOT NULL CHECK ( name != ''),
  sort_order INTEGER NOT NULL DEFAULT 0,
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text', 'comport_name', 'baud')), 
  min, max 
);

INSERT OR IGNORE INTO app_config (section, sort_order, var, name, type, min, max, value) VALUES
  ('СОМ порт приборов', 0, 'comport_products', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('СОМ порт приборов', 1, 'comport_products_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('СОМ порт приборов', 2, 'comport_products_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('СОМ порт приборов', 3, 'comport_products_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('СОМ порт приборов', 4, 'comport_products_repeat_count', 'количество повторов', 'integer', 0, 10, 0 );
`)
	return

}
