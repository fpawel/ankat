INSERT OR IGNORE INTO config ( sort_order, section_name, property_name, hint, type, min, max, default_value, value)
VALUES
   (0, :section_name, 'port', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1', 'COM1' ),
   (1, :section_name, 'baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600, 9600 ),
   (2, :section_name, 'timeout', 'таймаут, мс', 'integer', 10, 10000, 1000, 1000 ),
   (3, :section_name, 'byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50, 50 ),
   (4, :section_name, 'repeat_count', 'количество повторов', 'integer', 0, 10, 0, 0 ),
   (5, :section_name, 'bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0, 0 );