CREATE TABLE IF NOT EXISTS update_message (
    sensor_machine BIGINT REFERENCES machine_sensors,
    timestamp TIMESTAMP,
    meta JSON,
    column_definition JSON,
    data JSON,
    signature TEXT
);

CREATE TABLE IF NOT EXISTS next_analyses (
    id BIGSERIAL PRIMARY KEY,
    machine_sensors BIGINT REFERENCES machine_sensors(id) NOT NULL,
    previous_model BIGINT REFERENCES models(id),
    next_model BIGINT REFERENCES models(id)
);

CREATE TABLE IF NOT EXISTS pipeline (
    contract TEXT REFERENCES contracts(id),
    analyses BIGINT REFERENCES next_analyses(id)
);
