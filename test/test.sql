INSERT INTO machines (id) VALUES ('machine');
INSERT INTO sensors (id, transmitted_id) VALUES (0, 'sensor');

INSERT INTO machine_sensors (id, machine, sensor) VALUES (0, 'machine', 0);

INSERT INTO contracts (id) VALUES ('analyse');

INSERT INTO contract_machine_sensors (id, contract, machine_sensor) VALUES (0, 'analyse', 0);

INSERT INTO containers (id, url, tag) VALUES (0, 'url', 'tag');
INSERT INTO containers (id, url, tag) VALUES (1, 'url2', 'tag2');
INSERT INTO containers (id, url, tag) VALUES (2, 'url3', 'tag3');

INSERT INTO models (id, container) VALUES (0, 0);
INSERT INTO models (id, container) VALUES (1, 1);
INSERT INTO models (id, container) VALUES (2,2);

INSERT INTO systems (id, name) VALUES (0, 'system');

INSERT INTO pipelines (id, contract_machine_sensor, system) VALUES (0,0,0);

INSERT INTO analysis (pipeline, prev_model, next_model, execute, persist) VALUES (0, NULL, 1, 0, 'true');
INSERT INTO analysis (pipeline, prev_model, next_model, execute, persist) VALUES (0, 0, 2, 1, 'true');
INSERT INTO analysis (pipeline, prev_model, next_model, execute, persist) VALUES (0, 1, NULL, 2, 'true');

