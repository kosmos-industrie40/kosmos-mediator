# intern-mqtt-db

This project receives data from mqtt and write those data into a postgres database

## Table of Content

- [Dependencies](#dependencies)
- [Build](#build)
- [Test](#test)
- [Configuration](#configration)

## Dependencies
Golang 1.14 is used to write this endpoint. So golang is 
one of the requirements. We are using go modules to organize the sufficient dependencies. Those
are organized in the `go.mod` file.

There are a few extra infrastructure dependencies. You need to set up a PostgreSQL database server 
and a MQTT-Server. To install PostgreSQL check out [Download PostgresSQL page](https://www.postgresql.org/download/). 
As MQTT-Broker you can use Mosquitto from the eclipse foundation. To deploy
or install Mosquitto check out [Download Mosquitto page](https://mosquitto.org/download/)

## Build 
You can build this program by executing `make` or `go build ./...`. 

The database layout is given into two files. Part one can be found in the README of the [KOSMoS-Analyses-Cloud-Conector](https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector) reposistory. Part two will be used with the local createDatabase.sql file. To create the missing database table, you can execute the following command:
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

The following script will publish the sensor update messages and analys result to the mqtt broker.
```bash
for x in 'test/examplePayload/sensor*'; do
mosquitto_pub -h <host> -p <port> -t kosmos/machine-data/machine/sensor/sensor/update -f $x
done

for x in 'test/examplePayload/analyse*'; do
mosquitto_pub -h <host> -p <port> -t kosmos/analyses/contract -f $x
done
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
