#!/bin/bash

DELAY=25

docker-compose down
# docker rm -f $(docker ps -a -q)
# docker volume rm $(docker volume ls -q)

docker-compose up &

echo "****** Waiting for ${DELAY} seconds for containers to go up ******"
sleep $DELAY

docker exec mongo1 /scripts/rs-init.sh

docker exec kafka1 /scripts/topics-init.sh