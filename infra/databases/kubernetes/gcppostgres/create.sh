#!/bin/bash

# PREPARATION
# You need to create the credentials.json, follow this tutorial
# https://cloud.google.com/sql/docs/postgres/connect-container-engine
# place your credentials in ./infra/databases/kubernetes/gcppostgresbench/credentials.json

# TODO ONCE
createService () {
    gcloud beta sql instances create gcppostgresbench --cpu=1 --memory=3840MiB --region=us-central1 --database-version=POSTGRES_9_6;
    gcloud beta sql users set-password postgres no-host --instance gcppostgresbench --password test;
    kubectl create secret generic cloudsql-instance-credentials --from-file=credentials.json=./infra/databases/kubernetes/gcppostgres/credentials.json;
    kubectl create secret generic cloudsql-db-credentials --from-literal=username=root --from-literal=password=test;
    kubectl create -f ./infra/databases/kubernetes/gcppostgres/cloud-proxy.yml;
    initdb;
}

initdb () {
    # still not functional, to do manually on https://console.cloud.google.com/sql/instances/gcppostgresbench/databases?project=playwithsql
    ACCESS_TOKEN=$(GOOGLE_APPLICATION_CREDENTIALS="./infra/databases/kubernetes/gcppostgres/credentials.json" gcloud auth application-default print-access-token);
    curl --header "Authorization: Bearer ${ACCESS_TOKEN}" \
        https://www.googleapis.com/sql/v1beta4/projects/playwithsql/instances/playwithsql:us-central1:gcppostgresbench/databases/playwithsql -X DELETE;
    curl --header "Authorization: Bearer ${ACCESS_TOKEN}" \
        --header 'Content-Type: application/json' \
        --data '{"project": "playwithsql", "instance": "playwithsql:us-central1:gcppostgresbench", "name": "playwithsql"}' \
        https://www.googleapis.com/sql/v1beta4/projects/playwithsql/instances/playwithsql:us-central1:gcppostgresbench/databases -X POST
}

createService;