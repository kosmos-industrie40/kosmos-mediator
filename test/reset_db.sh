#!/bin/bash

read -r -p 'Enter DB user: ' db_user
read -r -sp 'Enter DB password: ' db_pass

PGPASSWORD=${db_pass} psql -d connector -h localhost -U "$db_user" -f reset.sql
PGPASSWORD=${db_pass} psql -d connector -h localhost -U "$db_user" -f ../createDatabase.sql
PGPASSWORD=${db_pass} psql -d connector -h localhost -U "$db_user" -f grant.sql
PGPASSWORD=${db_pass} psql -d connector -h localhost -U "$db_user" -f test.sql

echo "Success."