package dataconfig

const SQLConfig = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_checked (
  work_order INTEGER PRIMARY KEY,
  checked    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS section (
  section_name TEXT PRIMARY KEY CHECK (section_name != ''),
  hint         TEXT    NOT NULL CHECK (hint != ''),
  sort_order   INTEGER NOT NULL UNIQUE CHECK (sort_order >= 0)
);

CREATE TABLE IF NOT EXISTS config (
  section_name  TEXT    NOT NULL CHECK (section_name != ''),
  property_name TEXT CHECK (property_name != ''),
  hint          TEXT    NOT NULL CHECK (hint != ''),
  sort_order    INTEGER NOT NULL CHECK (sort_order >= 0),
  type          TEXT    NOT NULL CHECK (type in ('bool', 'integer', 'real', 'text', 'comport_name', 'baud')),
  default_value         NOT NULL,
  min,
  max,
  value                 NOT NULL,
  CONSTRAINT this_primary_key UNIQUE (property_name, section_name),
  FOREIGN KEY (section_name) REFERENCES section (section_name)
);


CREATE TABLE IF NOT EXISTS value_list (
  property_name NOT NULL CHECK (property_name IS NOT ''),
  value         NOT NULL,
  UNIQUE (property_name, value)
);

INSERT
OR IGNORE INTO section (sort_order, section_name, hint)
VALUES (0, 'comport_products', 'Связь с приборами'),
       (1, 'comport_gas', 'Пневмоблок'),
       (2, 'comport_temperature', 'Термокамера'),
       (3, 'automatic_work', 'Автоматическая настройка');

INSERT
OR IGNORE INTO config (sort_order, section_name, property_name, hint, type, min, max, default_value, value)
VALUES (0, 'automatic_work', 'delay_blow_nitrogen', 'Длит. продувки N2, мин.', 'integer', 1, 10, 3, 3),
       (1, 'automatic_work', 'delay_blow_gas', 'Длит. продувки изм. газа, мин.', 'integer', 1, 10, 3, 3),
       (2, 'automatic_work', 'delay_temperature', 'Длит. выдержки на температуре, часов', 'integer', 1, 5, 3, 3),
       (3, 'automatic_work', 'delta_temperature', 'Погрешность установки температуры, "С', 'integer', 1, 5, 3, 3),
       (4,
        'automatic_work',
        'timeout_temperature',
        'Таймаут установки температуры, минут',
        'integer',
        5,
        270,
        120,
        120);

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
