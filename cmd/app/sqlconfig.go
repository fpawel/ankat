package main

const SQLConfigDB = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_checked (
  work_order INTEGER PRIMARY KEY,
  checked TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS section(
  name TEXT PRIMARY KEY CHECK ( name != ''),
  sort_order INTEGER NOT NULL UNIQUE CHECK (sort_order >= 0)
);

CREATE TABLE IF NOT EXISTS config (
  var NOT NULL PRIMARY KEY CHECK ( var != ''),
  value NOT NULL,
  name TEXT NOT NULL CHECK ( name != ''),
  section_name TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text', 'comport_name', 'baud')),
  min, max,
  FOREIGN KEY(section_name) REFERENCES section(name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS value_list(
  var NOT NULL CHECK ( var IS NOT ''),
  value NOT NULL,
  UNIQUE (var, value)
);

INSERT OR IGNORE INTO section(name, sort_order) VALUES 
  ('Связь с приборами', 0),
  ('Пневмоблок', 1),
  ('Термокамера', 2),
  ('Автоматическая настройка', 3)
  ;

INSERT OR IGNORE INTO config (section_name, sort_order, var, name, type, min, max, value) VALUES
  ('Связь с приборами', 0, 'comport_products', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Связь с приборами', 1, 'comport_products_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Связь с приборами', 2, 'comport_products_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Связь с приборами', 3, 'comport_products_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Связь с приборами', 4, 'comport_products_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Связь с приборами', 5, 'comport_products_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Связь с приборами', 6, 'comport_products_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 );

INSERT OR IGNORE INTO config (section_name, sort_order, var, name, type, min, max, value) VALUES
  ('Пневмоблок', 0, 'comport_gas', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Пневмоблок', 1, 'comport_gas_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Пневмоблок', 2, 'comport_gas_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Пневмоблок', 3, 'comport_gas_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Пневмоблок', 4, 'comport_gas_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Пневмоблок', 5, 'comport_gas_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Пневмоблок', 6, 'comport_gas_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 );

INSERT OR IGNORE INTO config (section_name, sort_order, var, name, type, min, max, value) VALUES
  ('Термокамера', 0, 'comport_temp', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Термокамера', 1, 'comport_temp_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Термокамера', 2, 'comport_temp_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Термокамера', 3, 'comport_temp_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Термокамера', 4, 'comport_temp_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Термокамера', 5, 'comport_temp_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Термокамера', 6, 'comport_temp_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 );

INSERT OR IGNORE INTO config (section_name, sort_order, var, name, type, min, max, value) VALUES
  ('Автоматическая настройка', 0, 'delay_blow_nitrogen', 'Длительность продувки азота, минут', 'integer', 1, 10, 3 ),
  ('Автоматическая настройка', 1, 'delay_blow_gas', 'Длительность продувки измеряемого газа, минут', 'integer', 1, 10, 3 )
;

INSERT OR IGNORE INTO value_list(var, value) VALUES
  ('gas1', 'CH₄'),
  ('gas1', 'C₃H₈'),
  ('gas1', '∑CH'),
  ('gas1', 'CO₂'),
  ('gas2', 'CH₄'),
  ('gas2', 'C₃H₈'),
  ('gas2', '∑CH'),
  ('gas2', 'CO₂'),
  ('scale1', 2),
  ('scale1', 5),
  ('scale1', 10),
  ('scale1', 100),
  ('scale2', 2),
  ('scale2', 5),
  ('scale2', 10),
  ('scale2', 100),
  ('sensors_count', 1),
  ('sensors_count', 2) ;
`
