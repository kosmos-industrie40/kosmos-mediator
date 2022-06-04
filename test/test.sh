for x in examplePayload/data*; do
content=$(<$x)
mosquitto_pub -h localhost -p 1883 -t kosmos/machine-data/machine/sensor/sensor/update -m "$content"
mosquitto_pub -h localhost -p 1883 -t kosmos/machine-data/machine/sensor/sensor/update/temporary -m "$content"
done

for x in examplePayload/analyse*; do
content=$(<$x)
mosquitto_pub -h localhost -p 1883 -t kosmos/analyses/contract -m "$content"
mosquitto_pub -h localhost -p 1883 -t kosmos/analyses/contract/temporary -m "$content"
done