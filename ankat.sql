PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_log (
  record_id INTEGER PRIMARY KEY,
  parent_record_id INTEGER,
  created_at TIMESTAMP NOT NULL UNIQUE DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  work TEXT,
  work_index INTEGER,
  party_id INTEGER,
  product_serial INTEGER,

  level INTEGER ,
  message TEXT,

  CHECK( message ISNULL  AND level ISNULL  AND work NOTNULL AND work_index NOTNULL  OR
         message NOTNULL AND level NOTNULL AND work ISNULL  AND work_index ISNULL ),

  FOREIGN KEY(parent_record_id) REFERENCES work_log(record_id) ON DELETE CASCADE,
  FOREIGN KEY(party_id) REFERENCES party(party_id)  ON DELETE CASCADE,
  FOREIGN KEY(party_id, product_serial) REFERENCES product(party_id, product_serial)  ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_validate_work_party_id
  AFTER INSERT ON work_log
  FOR EACH ROW
  WHEN (NEW.party_id IS NULL)
BEGIN
  UPDATE work_log
  SET party_id = (SELECT current_party.party_id FROM current_party)
  WHERE record_id = NEW.record_id;
END;


WITH RECURSIVE acc(record_id, parent_record_id, created_at, work_index, party_id, product_serial, level, message) AS (
  SELECT
    record_id, parent_record_id, created_at, work_index, party_id, product_serial, level, message
  FROM last_work_log WHERE last_work_log.work_index = 9
  UNION
  SELECT
    w.record_id, w.parent_record_id, w.created_at,
    w.work_index  ,
    w.party_id, w.product_serial, w.level, w.message
  FROM acc
    INNER JOIN last_work_log w ON w.parent_record_id = acc.record_id
)
SELECT acc.created_at, l.work_index, acc.level, acc.message FROM acc
  INNER JOIN work_log l ON acc.parent_record_id = l.record_id
WHERE acc.message NOT NULL AND acc.level NOT NULL;

SELECT a.record_id, a.message, a.level, a.created_at, b.work
FROM work_log a
  INNER JOIN work_log b ON a.parent_record_id = b.record_id
where a.record_id = 37;


INSERT OR IGNORE  INTO party_var(sort_order, var, name, type, min, max, def_val) VALUES
  (0, 'product_type_number', 'номер исполнения', 'integer', 1, NULL , 10),
  (1, 'sensors_count', 'количество каналов', 'integer', 1, 2, 1),
  (1, 'gas1', 'газ канала 1', 'text', NULL, NULL, 'CH₄' ),
  (2, 'gas2', 'газ канала 2', 'text', NULL, NULL, 'CH₄'),
  (3, 'scale1', 'шкала канала 1', 'real', 0, NULL, 2 ),
  (4, 'scale2', 'шкала канала 2', 'real', 0, NULL, 2),
  (1, 'gas1','азот, ПГС1', 'real', 0, NULL, 0),
  (2, 'gas2ch1','середина, к.1, ПГС2', 'real', 0, NULL, 0.67),
  (3, 'gas2ch1+','доп. CO2, к.1, ПГС3', 'real', 0, NULL, 1.33),
  (4, 'gas3ch1', 'шкала, к.1, ПГС4', 'real', 0, NULL, 2),
  (5, 'gas2ch2', 'середина, к.2, ПГС5', 'real', 0, NULL, 1.33),
  (6, 'gas3ch2', 'шкала, к.2, ПГС6', 'real', 0, NULL, 2),
  (7, 't-', 'T-,"С', 'real', NULL, NULL, -30 ),
  (8, 't+', 'T+,"С', 'real', NULL, NULL, 45);

INSERT OR IGNORE INTO var_value_list(var, value) VALUES
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
  ('scale2', 100);

INSERT OR REPLACE INTO party_value (party_id, var, value)
VALUES ((SELECT * FROM current_party_id), 'gas1', 'CH₄');

SELECT pv.var, value, typeof(value) FROM party_value
  INNER JOIN party_var pv on party_value.var = pv.var
WHERE party_id IN  current_party_id
ORDER BY pv.sort_order;

CREATE VIEW IF NOT EXISTS party_info AS
  SELECT
    a.value || ' ' || b.value || ' ' || c.value
  FROM party p
    INNER JOIN party_value a on p.party_id = a.party_id and a.var = 'product_type_number'
    INNER JOIN party_value b on p.party_id = b.party_id and b.var = 'gas1'
    INNER JOIN party_value c on p.party_id = c.party_id and b.var = 'gas2';

SELECT
  a.value || ' ' || b.value || ' ' || cast(e.value AS INTEGER)  || (
    SELECT
      CASE d.value
      WHEN 1 THEN ''
      ELSE ' ' || c.value || ' ' || cast(g.value AS INTEGER)
      END )
FROM party p
  INNER JOIN party_value a on p.party_id = a.party_id and a.var = 'product_type_number'
  INNER JOIN party_value b on p.party_id = b.party_id and b.var = 'gas1'
  INNER JOIN party_value c on p.party_id = c.party_id and c.var = 'gas2'
  INNER JOIN party_value d on p.party_id = d.party_id and d.var = 'sensors_count'
  INNER JOIN party_value e on p.party_id = e.party_id and e.var = 'scale1'
  INNER JOIN party_value g on p.party_id = g.party_id and g.var = 'scale2'
;

WITH RECURSIVE acc(record_id, parent_record_id ) AS (
  SELECT
    record_id, parent_record_id
  FROM work_log WHERE work_log.record_id = 2
  UNION
  SELECT
    w.record_id, w.parent_record_id
  FROM acc
    INNER JOIN work_log w ON w.parent_record_id = acc.record_id
)
SELECT a.created_at, b.work_index,  a.level, a.message from acc
  INNER JOIN work_log a on a.record_id = acc.record_id
  INNER JOIN work_log b on a.parent_record_id = b.record_id
WHERE a.work ISNULL;