package main

const SQLComport = `
INSERT OR IGNORE INTO config (section_name, property_name, value)
VALUES
 ($1, 'port', 'COM1' ),
 ($1, 'baud',  9600 ),
 ($1, 'timeout', 1000 ),
 ($1, 'byte_timeout', 50 ),
 ($1, 'repeat_count', 0 ),
 ($1, 'bounce_timeout', 0 );`
