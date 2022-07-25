#!/bin/bash

for file_name in examplePayload/data*; do
  echo "Sending ${file_name}";
  content=$(<"$file_name")
  mosquitto_pub -h localhost -p 1883 -t kosmos/machine-data/machine/sensor/sensor/update -m "$content"
  mosquitto_pub -h localhost -p 1883 -t kosmos/machine-data/machine/sensor/sensor/update/temporary -m "$content"
done

for file_name in examplePayload/analyse*; do
  echo "Sending ${file_name}";
  content=$(<"$file_name")
  mosquitto_pub -h localhost -p 1883 -t kosmos/analyses/analyse -m "$content"
  mosquitto_pub -h localhost -p 1883 -t kosmos/analyses/analyse/temporary -m "$content"
done