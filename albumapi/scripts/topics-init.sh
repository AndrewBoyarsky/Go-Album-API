#!/bin/bash


/bin/kafka-topics --create --bootstrap-server kafka1:9092,kafka2:9093,kafka3:9094 --replication-factor 2 --partitions 3 --topic ProcessedAlbums
/bin/kafka-topics --create --bootstrap-server kafka1:9092,kafka2:9093,kafka3:9094 --replication-factor 2 --partitions 3 --topic NewAlbums