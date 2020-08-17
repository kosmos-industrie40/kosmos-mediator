CREATE TABLE IF NOT EXISTS update_message (sensor_machine BIGINT REFERENCES machine_sensor, timestamp TIMESTAMP, meta JSON, column_definition JSON, data JSON, signature TEXT);
CREATE TABLE IF NOT EXISTS next_analyses (id BIGSERIAL PRIMARY KEY, machine_sensor BIGINT REFERENCES machine_sensor(id) NOT NULL, previous_model BIGINT REFERENCES model(id), next_model BIGINT REFERENCES model(id));
CREATE TABLE IF NOT EXISTS pipeline (contract TEXT REFERENCES contract(id), analyses BIGINT REFERENCES next_analyses(id));
