#!/bin/bash

go build backend.go 
go build frontend.go 
echo "Binaries created; starting backends"

./backend --listen 8090 --backend :8091,:8092 & # start backend 1
./backend --listen 8091 --backend :8090,:8092 & # start backend 2
./backend --listen 8092 --backend :8090,:8091 & # start backend 3

sleep 5s #Just to make sure the backends are stable before starting the frontend

echo "Backends up; starting frontend"

./frontend --listen 8080 --backend :8090,:8091,:8092
killall -s 9 -w backend #To kill all of the backends that were started up

go clean