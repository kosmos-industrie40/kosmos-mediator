CREATE TABLE update_message (sensor_machine BIGINT REFERENCES machine_sensor, timestamp TIMESTAMP, meta JSON, column_definition JSON, data JSON, signature TEXT);
