PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_checked (
  work_order INTEGER PRIMARY KEY,
  checked TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS section(
  section_name TEXT PRIMARY KEY CHECK ( section_name != ''),
  hint TEXT NOT NULL CHECK ( hint != ''),
  sort_order INTEGER NOT NULL UNIQUE CHECK (sort_order >= 0)
);

CREATE TABLE IF NOT EXISTS property(
  property_name TEXT PRIMARY KEY CHECK ( property_name != ''),
  hint TEXT NOT NULL CHECK ( hint != ''),
  sort_order INTEGER NOT NULL CHECK (sort_order >= 0),
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text', 'comport_name', 'baud')),
  default_value NOT NULL , min, max
);

CREATE TABLE IF NOT EXISTS config (
  property_name NOT NULL CHECK ( property_name != ''),
  section_name TEXT NOT NULL CHECK ( section_name != ''),
  value NOT NULL,

  CONSTRAINT this_primary_key UNIQUE (property_name, section_name),
  FOREIGN KEY(property_name) REFERENCES property(property_name),
  FOREIGN KEY(section_name) REFERENCES section(section_name)
);

CREATE TABLE IF NOT EXISTS value_list(
  property_name NOT NULL CHECK ( property_name IS NOT ''),
  value NOT NULL,
  UNIQUE (property_name, value)
);

INSERT OR IGNORE INTO section(sort_order, section_name, hint)
VALUES
 (0, 'comport_products', 'Связь с приборами'),
 (1, 'comport_gas', 'Пневмоблок'),
 (2, 'comport_temperature','Термокамера'),
 (3, 'automatic_work', 'Автоматическая настройка') ;

INSERT OR IGNORE INTO property ( sort_order, property_name, hint, type, min, max, default_value)
VALUES
 (0, 'port', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
 (1, 'baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
 (2, 'timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
 (3, 'byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
 (4, 'repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
 (5, 'bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),

 (0, 'delay_blow_nitrogen', 'Длит. продувки N2, мин.', 'integer', 1, 10, 3 ),
 (1, 'delay_blow_gas', 'Длит. продувки изм. газа, мин.', 'integer', 1, 10, 3 ),
 (2, 'delay_temperature', 'Длит. выдержки на температуре, часов', 'integer', 1, 5, 3 ),
 (3, 'delta_temperature', 'Погрешность установки температуры, "С', 'integer', 1, 5, 3 ),
 (4, 'timeout_temperature', 'Таймаут установки температуры, минут', 'integer', 5, 270, 120 );

INSERT OR IGNORE INTO value_list(property_name, value) VALUES
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
