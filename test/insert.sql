INSERT INTO machines (id) VALUES ('machine');
INSERT INTO sensors (id, transmitted_id) VALUES (0, 'sensor');
INSERT INTO machine_sensors (id, machine, sensor) VALUES (0, 'machine', 0);
INSERT INTO contract (id) VALUES ('contract');
INSERT INTO models (id, url, tag) VALUES (0, 'url', 'tag');
INSERT INTO contracts (id) VALUES ('analyse');
INSERT INTO next_analyses(id, machine_sensor, previous_model, next_model) VALUES (0, 0, 0, 0);
INSERT INTO next_analyses(id, machine_sensor, previous_model, next_model) VALUES (0, 0, NULL, 0);
INSERT INTO pipeline (contract, analyses) VALUES ('analyse', 0);
INSERT INTO pipeline (contract, analyses) VALUES ('analyse', 1);
INSERT INTO machine_contract (machine, contract) VALUES ('machine', 'analyse');