PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party (
  party_id            INTEGER          NOT NULL  PRIMARY KEY,
  created_at          TIMESTAMP UNIQUE NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  product_type_number INTEGER          NOT NULL DEFAULT 22 CHECK (product_type_number > 0),
  sensors_count       INTEGER          NOT NULL DEFAULT 1 CHECK (sensors_count IN (1, 2)),
  pressure_sensor     INTEGER          NOT NULL DEFAULT 1 CHECK (pressure_sensor IN (0, 1)),
  concentration_gas1  REAL             NOT NULL DEFAULT 0,
  concentration_gas2  REAL             NOT NULL DEFAULT 50,
  concentration_gas3  REAL             NOT NULL DEFAULT 70,
  concentration_gas4  REAL             NOT NULL DEFAULT 100,
  concentration_gas5  REAL             NOT NULL DEFAULT 0.67,
  concentration_gas6  REAL             NOT NULL DEFAULT 2,
  temperature_minus   REAL             NOT NULL DEFAULT -30.,
  temperature_plus    REAL             NOT NULL DEFAULT 45.,
  gas1                TEXT             NOT NULL DEFAULT 'CH₄',
  gas2                TEXT             NOT NULL DEFAULT 'CH₄',
  scale1              REAL             NOT NULL DEFAULT 100,
  scale2              REAL             NOT NULL DEFAULT 100,
  units1              TEXT             NOT NULL DEFAULT '%, НКПР',
  units2              TEXT             NOT NULL DEFAULT '%, НКПР'

);

CREATE VIEW IF NOT EXISTS party_year AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year
  FROM party
  ORDER BY year;

CREATE VIEW IF NOT EXISTS party_year_month AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year,
                  cast(strftime('%m', created_at) AS INTEGER) AS month
  FROM party
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS party_year_month_day AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year,
                  cast(strftime('%m', created_at) AS INTEGER) AS month,
                  cast(strftime('%d', created_at) AS INTEGER) AS day
  FROM party
  ORDER BY created_at;

CREATE TABLE IF NOT EXISTS product (
  party_id       INTEGER NOT NULL,
  product_serial INTEGER NOT NULL,
  CONSTRAINT unique_serial UNIQUE (party_id, product_serial),
  CONSTRAINT positive_serial CHECK (product_serial > 0),
  FOREIGN KEY (party_id) REFERENCES party (party_id)
    ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS current_party AS
  SELECT *
  FROM party
  ORDER BY created_at DESC
  LIMIT 1;

CREATE VIEW IF NOT EXISTS current_party_products AS
  SELECT product_serial
  FROM product
  WHERE party_id IN (SELECT party_id FROM current_party);

CREATE VIEW IF NOT EXISTS current_party_products_enumerated AS
  SELECT count(*) - 1 AS ordinal, cur.product_serial AS product_serial
  FROM current_party_products AS cur
         LEFT JOIN current_party_products AS oth
  WHERE cur.product_serial >= oth.product_serial
  GROUP BY cur.product_serial;

CREATE TABLE IF NOT EXISTS product_config (
  ordinal INTEGER PRIMARY KEY CHECK (ordinal >= 0 AND typeof(ordinal) = 'integer'),
  checked INTEGER NOT NULL DEFAULT 1 CHECK (checked IN (0, 1)),
  comport TEXT    NOT NULL DEFAULT 'COM1'
);

CREATE VIEW IF NOT EXISTS current_party_products_config AS
  SELECT a.ordinal, a.product_serial, IFNULL(b.checked, 1) AS checked, IFNULL(b.comport, 'COM1') AS comport
  FROM current_party_products_enumerated a
         LEFT JOIN product_config b ON a.ordinal = b.ordinal;

CREATE TABLE IF NOT EXISTS read_var (
  var         INTEGER NOT NULL PRIMARY KEY CHECK (typeof(var) = 'integer' AND var >= 0),
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL                           DEFAULT '',
  checked     INTEGER NOT NULL CHECK (checked IN (0, 1)) DEFAULT 1
);

CREATE VIEW IF NOT EXISTS read_var_enumerated AS
  SELECT count(*) - 1 AS ordinal, a.var AS var, a.checked AS checked, a.name AS name, a.description AS description
  FROM read_var AS a
         LEFT JOIN read_var AS b
  WHERE a.var >= b.var
  GROUP BY a.var;

CREATE TABLE IF NOT EXISTS product_value (
  party_id       INTEGER NOT NULL,
  product_serial REAL    NOT NULL,
  section        TEXT    NOT NULL,
  point          INTEGER NOT NULL,
  var            INTEGER NOT NULL,
  value          REAL    NOT NULL,

  UNIQUE (party_id, product_serial, section, point, var),

  FOREIGN KEY (var) REFERENCES read_var (var),
  FOREIGN KEY (party_id, product_serial)
  REFERENCES product (party_id, product_serial)
    ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS party_info AS
  SELECT *, cast(strftime('%Y', created_at) AS INTEGER)                                  AS year,
            cast(strftime('%m', created_at) AS INTEGER)                                  AS month,
            cast(strftime('%d', created_at) AS INTEGER)                                  AS day,
            p.product_type_number || ' ' || p.gas1 || ' ' || cast(p.scale1 AS INTEGER)
              || (CASE p.sensors_count
                    WHEN 1 THEN ''
                    ELSE ' ' || p.gas2 || ' ' || cast(p.scale2 AS INTEGER)
                END)                                                                     AS what,
            (SELECT exists(SELECT * FROM last_work_log w WHERE w.party_id = p.party_id)) AS has_log
  FROM party p;

CREATE TABLE IF NOT EXISTS main_error_source (
  party_id       INTEGER NOT NULL,
  product_serial INTEGER NOT NULL,
  sensor         INTEGER NOT NULL CHECK (sensor IN (1, 2)),
  scale          INTEGER NOT NULL CHECK (scale IN ('SCALE_BEGIN', 'SCALE_MIDDLE', 'SCALE_END')),
  temp           TEXT    NOT NULL CHECK (temp IN ('T_LOW', 'T_NORM', 'T_HIGH')),
  value          REAL    NOT NULL CHECK (typeof(value) IN ('real', 'integer')),

  UNIQUE (party_id, product_serial, sensor, scale, temp),

  FOREIGN KEY (party_id, product_serial)
  REFERENCES product (party_id, product_serial)
    ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS main_error1 AS
  SELECT *, (CASE a.sensor
               WHEN 1 THEN CASE a.scale
                             WHEN 'SCALE_BEGIN' THEN p.concentration_gas1
                             WHEN 'SCALE_MIDDLE' THEN p.concentration_gas2
                             WHEN 'SCALE_END' THEN p.concentration_gas4 END
               WHEN 2 THEN CASE a.scale
                             WHEN 'SCALE_BEGIN' THEN p.concentration_gas1
                             WHEN 'SCALE_MIDDLE' THEN p.concentration_gas5
                             WHEN 'SCALE_END' THEN p.concentration_gas6 END END) nominal,
            (CASE a.sensor
               WHEN 1 THEN p.units1
               WHEN 2 THEN p.units2 END) AS                                      units,
            (CASE a.sensor
               WHEN 1 THEN p.gas1
               WHEN 2 THEN p.gas2 END)   AS                                      gas,
            (CASE a.sensor
               WHEN 1 THEN p.scale1
               WHEN 2 THEN p.scale2 END) AS                                      scale_value
  FROM main_error_source a
         INNER JOIN party p ON p.party_id = a.party_id;

CREATE VIEW IF NOT EXISTS main_error2 AS
  SELECT *, (
                CASE units
                  WHEN '%, НКПР' THEN 5
                  WHEN 'объемная доля, %' THEN CASE gas
                                                 WHEN 'CH₄' THEN 0.22
                                                 WHEN 'C₃H₈' THEN 0.05
                                                 WHEN 'CO₂' THEN CASE scale_value
                                                                   WHEN 2. THEN 0.1
                                                                   WHEN 5. THEN 0.25
                                                                   WHEN 10. THEN 0.5
                        END
                    END
                    END) AS absolute_error_limit
  FROM main_error1 a;

CREATE VIEW IF NOT EXISTS main_error3 AS
  SELECT *, (value - nominal) AS abolute_error
  FROM main_error2;

CREATE VIEW IF NOT EXISTS main_error AS
  SELECT *, abs(abolute_error) < absolute_error_limit       AS ok,
            100 * abs(abolute_error) / absolute_error_limit AS percent_from_absolute_error_limit
  FROM main_error3;

INSERT
OR IGNORE INTO read_var (var, name, description)
VALUES (0, 'CCh0', 'концентрация - канал 1 (электрохимия 1)'),
       (2, 'CCh1', 'концентрация - канал 2 (электрохимия 2/оптика 1)'),
       (4, 'CCh2', 'концентрация - канал 3 (оптика 1/оптика 2)'),
       (6, 'PkPa', 'давление, кПа'),
       (8, 'Pmm', 'давление, мм. рт. ст'),
       (10, 'Tmcu', 'температура микроконтроллера, град.С'),
       (12, 'Vbat', 'напряжение аккумуляторной батареи, В'),
       (14, 'Vref', 'опорное напряжение для электрохимии, В'),
       (16, 'Vmcu', 'напряжение питания микроконтроллера, В'),
       (18, 'VdatP', 'напряжение на выходе датчика давления, В'),
       (640, 'CoutCh0', 'концентрация - первый канал оптики'),
       (642, 'TppCh0', 'температура пироприемника - первый канал оптики'),
       (644, 'ILOn0', 'лампа ВКЛ - первый канал оптики'),
       (646, 'ILOff0', 'лампа ВЫКЛ - первый канал оптики'),
       (648, 'Uw_Ch0', 'значение исходного сигнала в рабочем канале (АЦП) - первый канал оптики'),
       (650, 'Ur_Ch0', 'значение исходного сигнала в опорном канале (АЦП) - первый канал оптики'),
       (652, 'WORK0', 'значение нормализованного сигнала в рабочем канале (АЦП) - первый канал оптики'),
       (654, 'REF0', 'значение нормализованного сигнала в опроном канале (АЦП) - первый канал оптики'),
       (656, 'Var1Ch0', 'значение дифференциального сигнала - первый канал оптики'),
       (658, 'Var2Ch0', 'значение дифференциального сигнала с поправкой по нулю от температуры - первый канал оптики'),
       (660,
        'Var3Ch0',
        'значение дифференциального сигнала с поправкой по чувствительности от температуры - первый канал оптики'),
       (662, 'FppCh0', 'частота преобразования АЦП - первый канал оптики'),
       (672, 'CoutCh1', 'концентрация - второй канал оптики'),
       (674, 'TppCh1', 'температура пироприемника - второй канал оптики'),
       (676, 'ILOn1', 'лампа ВКЛ - второй канал оптики'),
       (678, 'ILOff1', 'лампа ВЫКЛ - второй канал оптики'),
       (680, 'Uw_Ch1', 'значение исходного сигнала в рабочем канале (АЦП) - второй канал оптики'),
       (682, 'Ur_Ch1', 'значение исходного сигнала в опорном канале (АЦП) - второй канал оптики'),
       (684, 'WORK1', 'значение нормализованного сигнала в рабочем канале (АЦП) - второй канал оптики'),
       (686, 'REF1', 'значение нормализованного сигнала в опроном канале (АЦП) - второй канал оптики'),
       (688, 'Var1Ch1', 'значение дифференциального сигнала - второй канал оптики'),
       (690, 'Var2Ch1', 'значение дифференциального сигнала с поправкой по нулю от температуры - второй канал оптики'),
       (692,
        'Var3Ch1',
        'значение дифференциального сигнала с поправкой по чувствительности от температуры - второй канал оптики'),
       (694, 'FppCh1', 'частота преобразования АЦП - второй канал оптики');

CREATE TABLE IF NOT EXISTS work (
  work_id        INTEGER   NOT NULL PRIMARY KEY,
  parent_work_id INTEGER,
  created_at     TIMESTAMP NOT NULL UNIQUE DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  work_name      TEXT      NOT NULL,
  work_index     INTEGER   NOT NULL,
  party_id       INTEGER,
  FOREIGN KEY (parent_work_id) REFERENCES work (work_id),
  FOREIGN KEY (party_id) REFERENCES party (party_id)
    ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_validate_work_party_id
  AFTER INSERT
  ON work
  FOR EACH ROW
  WHEN (NEW.party_id IS NULL)
BEGIN
  UPDATE work
  SET party_id = (SELECT current_party.party_id
                  FROM current_party)
  WHERE work_id = NEW.work_id;
END;

CREATE TABLE IF NOT EXISTS work_log (
  record_id      INTEGER   NOT NULL PRIMARY KEY,
  created_at     TIMESTAMP NOT NULL UNIQUE DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  work_id        INTEGER   NOT NULL,
  product_serial INTEGER,
  level          INTEGER   NOT NULL CHECK (level >= 0),
  message        TEXT      NOT NULL CHECK (message != ''),
  FOREIGN KEY (work_id) REFERENCES work (work_id)
);

CREATE VIEW IF NOT EXISTS work_log2 AS
  SELECT l.record_id,
         w.party_id,
         l.work_id,
         w.parent_work_id,
         l.created_at,
         w.work_name,
         w.work_index,
         l.level,
         l.message,
         l.product_serial
  FROM work_log l
         INNER JOIN work w on l.work_id = w.work_id;

CREATE VIEW IF NOT EXISTS last_work_root AS
  SELECT *
  FROM work
  WHERE parent_work_id ISNULL
  ORDER BY created_at DESC
  LIMIT 1;

CREATE VIEW IF NOT EXISTS last_work AS
  SELECT *
  FROM work
  WHERE created_at >= (SELECT created_at FROM last_work_root)
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS last_work_log AS
  SELECT b.work_id, a.parent_work_id, b.created_at, a.party_id, b.product_serial, b.level, b.message
  FROM last_work a
         INNER JOIN work_log b ON a.work_id = b.work_id
  ORDER BY b.created_at;

CREATE VIEW IF NOT EXISTS work_info AS
  SELECT w.work_id,
         w.parent_work_id,
         w.created_at,
         w.work_index,
         w.work_name,
         EXISTS(SELECT * FROM work ww WHERE ww.parent_work_id = w.work_id) AS has_children,
         (WITH RECURSIVE a(work_id, parent_work_id) AS
         (SELECT work_id, parent_work_id FROM work b WHERE b.work_id = w.work_id
                                                        OR b.parent_work_id = w.work_id
          UNION
          SELECT w.work_id, w.parent_work_id FROM a
                                                    INNER JOIN work w ON w.parent_work_id = a.work_id
         ) SELECT EXISTS(SELECT *
                         FROM a
                                INNER JOIN work_log ON a.work_id = work_log.work_id
                         WHERE work_log.level >= 4))                       AS has_error
  FROM work w;

CREATE TABLE IF NOT EXISTS coefficient (
  coefficient_id INTEGER NOT NULL  PRIMARY KEY CHECK (typeof(coefficient_id) = 'integer' AND coefficient_id >= 0),
  name           TEXT    NOT NULL,
  description    TEXT                                       DEFAULT '',
  checked        INTEGER NOT NULL CHECK (checked IN (0, 1)) DEFAULT 1
);

CREATE TABLE IF NOT EXISTS product_coefficient_value (
  party_id       INTEGER NOT NULL,
  product_serial INTEGER NOT NULL,
  coefficient_id INTEGER NOT NULL,
  value          REAL    NOT NULL,

  UNIQUE (party_id, product_serial, coefficient_id),

  FOREIGN KEY (coefficient_id)
  REFERENCES coefficient (coefficient_id),
  FOREIGN KEY (party_id, product_serial)
  REFERENCES product (party_id, product_serial)
    ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS current_party_coefficient_value AS
  SELECT ordinal, a.product_serial, coefficient_id, value
  FROM product_coefficient_value a
         INNER JOIN current_party_products_config b ON a.product_serial = b.product_serial
  WHERE party_id IN (SELECT party_id FROM current_party);

INSERT
OR IGNORE INTO coefficient (coefficient_id, name, description)
VALUES (0, 'VER_PO', 'номер версии ПО'),
       (1, 'PPRIBOR_TYPE', 'номер исполнения прибора'),
       (2, 'YEAR', 'год выпуска'),
       (3, 'SER_NUMBER', 'серийный номер'),
       (4, 'Kef4', 'максимальное число регистров в таблице регистров прибора'),
       (5, 'ED_IZMER_1', 'единицы измерения канала 1 ИКД'),
       (6, 'Gas_Type_1', 'величина, измеряемая каналом 1 ИКД'),
       (7, 'SHKALA_1', 'диапазон измерений канала 1 ИКД'),
       (8, 'PREDEL_LO_1', 'начало шкалы канала 1 ИКД'),
       (9, 'PREDEL_HI_1', 'конец шкалы канала 1 ИКД'),
       (10, 'Pgs1_1', 'значение ПГС1 (начало шкалы) канала 1 ИКД'),
       (11, 'Pgs3_1', 'значение ПГС3 (конец шкалы) канала 1 ИКД'),
       (12, 'KNull_1', 'коэффициент калибровки нуля канала 1 ИКД'),
       (13, 'KSens_1', 'коэффициент калибровки чувствительности канала 1 ИКД'),
       (14, 'ED_IZMER_2', 'единицы измерения канала 2 ИКД'),
       (15, 'Gas_Type_2', 'величина, измеряемая каналом 2 ИКД'),
       (16, 'SHKALA_2', 'диапазон измерений канала 2 ИКД'),
       (17, 'PREDEL_LO_2', 'начало шкалы канала 2 ИКД'),
       (18, 'PREDEL_HI_2', 'конец шкалы канала 2 ИКД'),
       (19, 'Pgs1_2', 'пГС1 (начало шкалы) канала 2 ИКД'),
       (20, 'Pgs3_2', 'пГС3 (конец шкалы) канала 2 ИКД'),
       (21, 'KNull_2', 'коэффициент калибровки нуля канала 2 ИКД'),
       (22, 'KSens_2', 'коэффициент калибровки чувствительности канала 2 ИКД'),
       (23, 'CLin1_0', '0-ой степени кривой линеаризации канала 1 ИКД'),
       (24, 'CLin1_1', '1-ой степени кривой линеаризации канала 1 ИКД'),
       (25, 'CLin1_2', '2-ой степени кривой линеаризации канала 1 ИКД'),
       (26, 'CLin1_3', '3-ей степени кривой линеаризации канала 1 ИКД'),
       (27, 'KNull_T1_0', '0-ой степени полинома коррекции нуля от температуры канала 1 ИКД'),
       (28, 'KNull_T1_1', '1-ой степени полинома коррекции нуля от температуры канала 1 ИКД'),
       (29, 'KNull_T1_2', '2-ой степени полинома коррекции нуля от температуры канала 1 ИКД'),
       (30, 'KSens_T1_0', '0-ой степени полинома кор. чувств. от температуры канала 1 ИКД'),
       (31, 'KSens_T1_1', '1-ой степени полинома кор. чувств. от температуры канала 1 ИКД'),
       (32, 'KSens_T1_2', '2-ой степени полинома кор. чувств. от температуры канала 1 ИКД'),
       (33, 'CLin2_0', '0-ой степени кривой линеаризации канала 2 ИКД'),
       (34, 'CLin2_1', '1-ой степени кривой линеаризации канала 2 ИКД'),
       (35, 'CLin2_2', '2-ой степени кривой линеаризации канала 2 ИКД'),
       (36, 'CLin2_3', '3-ей степени кривой линеаризации канала 2 ИКД'),
       (37, 'KNull_T2_0', '0-ой степени полинома коррекции нуля от температуры канала 2 ИКД'),
       (38, 'KNull_T2_1', '1-ой степени полинома коррекции нуля от температуры канала 2 ИКД'),
       (39, 'KNull_T2_2', '2-ой степени полинома коррекции нуля от температуры канала 2 ИКД'),
       (40, 'KSens_T2_0', '0-ой степени полинома кор. чувств. от температуры канала 2 ИКД'),
       (41, 'KSens_T2_1', '1-ой степени полинома кор. чувств. от температуры канала 2 ИКД'),
       (42, 'KSens_T2_2', '2-ой степени полинома кор. чувств. от температуры канала 2 ИКД'),
       (43, 'Coef_Pmmhg_0', '0-ой степени полинома калибровки датчика P (в мм.рт.ст.)'),
       (44, 'Coef_Pmmhg_1', '1-ой степени полинома калибровки датчика P (в мм.рт.ст.)'),
       (45, 'KNull_TP_0', '0-ой степени полинома кор. нуля датчика давления от температуры'),
       (46, 'KNull_TP_1', '1-ой степени полинома кор. нуля датчика давления от температуры'),
       (47, 'KNull_TP_2', '2-ой степени полинома кор. нуля датчика давления от температуры'),
       (48, 'KdFt', 'чувствительность датчика температуры микроконтроллера, град.С/В'),
       (49, 'KFt', 'смещение датчика температуры микроконтроллера, град.С');

CREATE VIEW IF NOT EXISTS coefficient_enumerated AS
  SELECT count(*) - 1     AS ordinal,
         a.coefficient_id AS coefficient_id,
         a.checked        AS checked,
         a.name           AS name,
         a.description    AS description
  FROM coefficient AS a
         LEFT JOIN coefficient AS b
  WHERE a.coefficient_id >= b.coefficient_id
  GROUP BY a.coefficient_id;


CREATE TABLE IF NOT EXISTS series (
  series_id  INTEGER   NOT NULL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL UNIQUE,
  name       TEXT      NOT NULL,
  party_id   INTEGER   NOT NULL,
  FOREIGN KEY (party_id) REFERENCES party (party_id)
    ON DELETE CASCADE
);


CREATE VIEW IF NOT EXISTS series_info AS
  SELECT *, cast(strftime('%Y', created_at) AS INT) AS year,
            cast(strftime('%m', created_at) AS INT) AS month,
            cast(strftime('%d', created_at) AS INT) AS day
  FROM series
  ORDER BY created_at;

CREATE TABLE IF NOT EXISTS chart_value (
  series_id      INTEGER NOT NULL,
  party_id       INTEGER NOT NULL,
  product_serial INTEGER NOT NULL,
  var            INTEGER NOT NULL,
  seconds_offset REAL    NOT NULL,
  value          REAL    NOT NULL,
  UNIQUE (series_id, party_id, product_serial, var, seconds_offset),
  FOREIGN KEY (series_id) REFERENCES series (series_id)
    ON DELETE CASCADE,
  FOREIGN KEY (var) REFERENCES read_var (var),
  FOREIGN KEY (party_id) REFERENCES party (party_id),
  FOREIGN KEY (party_id, product_serial) REFERENCES product (party_id, product_serial)
);

CREATE VIEW IF NOT EXISTS chart_value_info AS
  SELECT strftime('%d.%m.%Y %H:%M:%f', s.created_at, '+' || b.seconds_offset || ' seconds') AS created_at,
         value,
         product_serial,
         b.var                                                                              AS var,
         b.series_id                                                                        AS series_id,
         b.party_id                                                                         AS party_id,
         r.name                                                                             AS var_name
  FROM chart_value AS b
         INNER JOIN series_info s on b.series_id = s.series_id
         INNER JOIN read_var r on b.var = r.var;

