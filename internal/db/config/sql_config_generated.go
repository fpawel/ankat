package config

const SQLConfig = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_checked (
  work_order INTEGER NOT NULL PRIMARY KEY,
  checked    TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS section (
  section_name TEXT    NOT NULL PRIMARY KEY CHECK (section_name != ''),
  hint         TEXT    NOT NULL CHECK (hint != ''),
  sort_order   INTEGER NOT NULL UNIQUE CHECK (sort_order >= 0)
);

CREATE TABLE IF NOT EXISTS config (
  section_name  TEXT    NOT NULL CHECK (section_name != ''),
  property_name TEXT    NOT NULL CHECK (property_name != ''),
  hint          TEXT    NOT NULL CHECK (hint != ''),
  sort_order    INTEGER NOT NULL CHECK (sort_order >= 0),
  type          TEXT    NOT NULL CHECK (type in ('bool', 'integer', 'real', 'text', 'comport_name', 'baud')),
  default_value         NOT NULL,
  min,
  max,
  value,
  CONSTRAINT this_primary_key UNIQUE (property_name, section_name),
  FOREIGN KEY (section_name) REFERENCES section (section_name)
);


CREATE TABLE IF NOT EXISTS value_list (
  property_name NOT NULL CHECK (property_name IS NOT ''),
  value         NOT NULL,
  UNIQUE (property_name, value)
);

CREATE TRIGGER IF NOT EXISTS trigger_set_default_value
  AFTER INSERT
  ON config
  FOR EACH ROW
  WHEN (NEW.value IS NULL)
BEGIN
  UPDATE config
  SET value = new.default_value
  WHERE section_name = new.section_name
    AND property_name = new.property_name;
END;


INSERT
OR IGNORE INTO section (sort_order, section_name, hint)
VALUES (0, 'party', 'Параметры партии'),
       (1, 'comport_products', 'Связь с приборами'),
       (2, 'comport_gas', 'Пневмоблок'),
       (3, 'comport_temperature', 'Термокамера'),
       (4, 'automatic_work', 'Автоматическая настройка');

INSERT
OR IGNORE
    INTO config (sort_order, section_name, property_name, hint, type, min, max, default_value)
VALUES (0, 'automatic_work', 'delay_blow_nitrogen', 'Длит. продувки N2, мин.', 'integer', 1, 10, 3),
       (1, 'automatic_work', 'delay_blow_gas', 'Длит. продувки изм. газа, мин.', 'integer', 1, 10, 3),
       (2, 'automatic_work', 'delay_temperature', 'Длит. выдержки на температуре, часов', 'integer', 1, 5, 3),
       (3, 'automatic_work', 'delta_temperature', 'Погрешность установки температуры, "С', 'integer', 1, 5, 3),
       (4, 'automatic_work', 'timeout_temperature', 'Таймаут установки температуры, минут', 'integer', 5, 270, 120),
       (0, 'party', 'product_type_number', 'номер исполнения', 'integer', 1, NULL, 10),
       (1, 'party', 'sensors_count', 'количество каналов', 'integer', 1, 2, 1),
       (2, 'party', 'pressure_sensor', 'Датчик давления', 'bool', NULL, NULL, 0),
       (3, 'party', 'gas1', 'газ к.1', 'text', NULL, NULL, 'CH₄'),
       (4, 'party', 'gas2', 'газ к.2', 'text', NULL, NULL, 'CH₄'),
       (5, 'party', 'scale1', 'шкала к.1', 'real', 0, NULL, 2),
       (6, 'party', 'scale2', 'шкала к.2', 'real', 0, NULL, 2),
       (7, 'party', 'units1', 'ед.изм. к.1', 'text', NULL, NULL, '%, НКПР'),
       (8, 'party', 'units2', 'ед.изм. к.2', 'text', NULL, NULL, '%, НКПР'),
       (9, 'party', 'temperature_minus', 'T-,"С', 'real', NULL, NULL, -30),
       (10, 'party', 'temperature_plus', 'T+,"С', 'real', NULL, NULL, 45),
       (11, 'party', 'concentration_gas1', 'ПГС1 азот', 'real', 0, NULL, 0),
       (12, 'party', 'concentration_gas2', 'ПГС2 середина к.1', 'real', 0, NULL, 0.67),
       (13, 'party', 'concentration_gas3', 'ПГС3 середина доп.CO₂', 'real', 0, NULL, 1.33),
       (14, 'party', 'concentration_gas4', 'ПГС4 шкала к.1', 'real', 0, NULL, 2),
       (15, 'party', 'concentration_gas5', 'ПГС5 середина к.2', 'real', 0, NULL, 1.33),
       (16, 'party', 'concentration_gas6', 'ПГС6 шкала к.2', 'real', 0, NULL, 2);

INSERT
OR IGNORE INTO value_list (property_name, value)
VALUES ('units1', 'объемная доля, %'),
       ('units1', '%, НКПР'),
       ('units2', 'объемная доля, %'),
       ('units2', '%, НКПР'),
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
       ('sensors_count', 2);
`
