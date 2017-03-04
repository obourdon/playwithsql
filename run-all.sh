#!/bin/bash

rm ./results*.log
./infra/databases/docker_swarm/cockroachdb/launch-solo.sh && ./run-docker.sh cockroachdb pws_cockroachdb 10000
./infra/databases/docker_swarm/mssql/launch-solo.sh && ./run-docker.sh mssql pws_mssql 10000
./infra/databases/docker_swarm/mysql/launch-solo.sh && ./run-docker.sh mysql pws_mysql 10000
# ./infra/databases/docker_swarm/mariadb/launch-solo.sh && ./run-docker.sh mariadb pws_mariadb 10000
# ./infra/databases/docker_swarm/percona/launch-solo.sh && ./run-docker.sh percona pws_percona 10000
./infra/databases/docker_swarm/oracle/launch-solo.sh && ./run-docker.sh oracle pws_oracle 10000
./infra/databases/docker_swarm/postgres/launch-solo.sh && ./run-docker.sh postgres pws_postgres 10000
./run-docker.sh sqlite pws_sqlite 5000 && rm 

 # ./infra/databases/docker_swarm/cockroachdb/launch-cluster.sh && ./run-docker.sh cockroachdb pws_cockroachdb-0 2000 && docker service logs -f pws-cmd