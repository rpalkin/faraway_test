#!/bin/bash

set -e
localport=8080
typename=svc/faraway-app
remoteport=80
kubectl port-forward -n faraway $typename $localport:$remoteport > /dev/null 2>&1 &
pid=$!
trap '{
    kill $pid
}' EXIT
while ! nc -vz localhost $localport > /dev/null 2>&1
do
    sleep 0.1
done
curl -s localhost:$localport
