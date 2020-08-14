# intern-mqtt-db

This project will insert received messages into a database and send a mqtt messages to the next ml-tool in the ml-pipeline

## Table of Content

- [Description](#description)
- [Dependencies](#dependencies)
- [Build](#build)
- [Test](#test)
- [Configuration](#configration)

## Description
This project has to two tasks.
1. receiving messages from the mqtt broker and write this data into a database
1. based on the received data finding the next step which has to be done in the ml pipeline

The first part of the task has to be extended by a non persistent part, because there are messages on which the data should not store in the
database. The topics on which this project will subscribe:
- `kosmos/analyses/+`
- `kosmos/analyses/+/temporary`
- `kosmos/machine-data/+/sensor/+/update`
- `kosmos/machine-data/+/sensor/+/update/temporary`

In the second part of the program produce messages. Those messages will be published on the topic matching the topic `kosmos/analyses/+/+`.
The first wild card subtopic contains the URL of the used ml-image. We use a URL encoding to remove special characters and replace all `/` with the `-`.
In the second wild card subtopic contains the tag of the image.

## Dependencies
Golang 1.14 is used to write this endpoint. So golang is 
one of the requirements. We are using go modules to organize the sufficient dependencies. Those
are organized in the `go.mod` file.

There are a few extra infrastructure dependencies. You need to set up a PostgreSQL database server 
and a MQTT-Server. To install PostgreSQL check out [Download PostgresSQL page](https://www.postgresql.org/download/). 
As MQTT-Broker you can use Mosquitto from the eclipse foundation. To deploy
or install Mosquitto check out [Download Mosquitto page](https://mosquitto.org/download/).

## Build 
You can build this program by executing `make` or `go build ./...`. 

The database layout is given in two files. The first part can be found in the  
[KOSMoS-Analyses-Cloud-Connector](https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector) 
repository. Part two is written in the local
`createDatabse.sql` file. To create the missing database table, you can execute the following command:
```bash
psql -h <host> -d <database> -U <database user> <createDatabase.sql
```
The following variables has to be set to you specific deployment:
- host
- database
- database user

## Test
To insert the used data (sensor, machine contract) into the database you can execute the following command:
```bash
psql -d <database> -h <host> -U <user> <test/insert.sql
```

The following script will publish the sensor update messages and analyse result to the mqtt broker.
```bash
for x in test/examplePayload/data*; do
mosquitto_pub -h <host> -p <port> -t kosmos/machine-data/machine/sensor/sensor/update -f $x
mosquitto_pub -h <host> -p <port> -t kosmos/machine-data/machine/sensor/sensor/update/temporary -f $x
done

for x in test/examplePayload/analyse*; do
mosquitto_pub -h <host> -p <port> -t kosmos/analyses/contract -f $x
mosquitto_pub -h <host> -p <port> -t kosmos/analyses/contract/temporary -f $x
done
```

To receive the messages from the mediator, you can use the following command:
```bash
mosquitto_sub -h <host> -p <port> -t 'kosmos/analytics/+/+'
``` 

## Configuration
The configuration of the application will be made through two configuration files and command line flags. 
The configuration parameters will be explained in the next three sections.

### CLI-Flags
In this section the command line parameters will be displayed. Flags which are created by the logging tool `klog` will not be
acknowledge in this chapter.

| flag | default value | description |
|------|---------------|-------------|
| pass | examplePassword.yaml | is the path to the password configuration file |
| config | exampleConfig.yaml | is the path to the configuration file |

### Password
The password configuration contains passwords for the database connection and the mqtt connection. An example can be
found in the `examplePassword.yaml` file.

|parameter|description|
| ------- | --------- |
| mqtt.user | is the user name of the mqtt user which is used for the mqtt connection |
| mqtt.password | is the password which is used by the mqtt.user for the mqtt connection |
| database.user | is the user for the postgresql database connection |
| database.password | is the password for the postgresql database connection |

### Configuration
The configuration file will be used to configure the system without including credentials. An example configuration
can be found in the `exampleConfiguration.yaml` file.

| parameter | description |
| --------- | ----------- |
| webserver.address | is the IP address on which this application will be open the web server|
| webserver.port | is the port this application used for the web server |
| database.address | is the IP address (or URL), where the PostgreSQL server could be found |
| database.port | is the port of the PostgreSQL server |
| database.database | is the name of the PostgreSQL database |
| mqtt.address | is the IP address (or URL) of the mqtt broker |
| mqtt.port | is the port of the mqtt broker|
| mqtt.tls | enables tls of the mqtt broker |
