#!/bin/bash - 
set -o nounset                              # Treat unset variables as an error

http -a view:qwerty123 http://localhost:8080/api/kubeconfig

sleep 1
http -a alexander:qwerty123 http://localhost:8080/api/kubeconfig
sleep 1
http -a alexander:qwerty123 http://localhost:8080/api/kubeconfig
sleep 1
http -a alexander:qwerty123 http://localhost:8080/api/kubeconfig

sleep 5
http -a alexander:qwerty123 DELETE http://localhost:8080/api/kubeconfig

sleep 2
http -a alexander:qwerty123 http://localhost:8080/metrics | grep totem
