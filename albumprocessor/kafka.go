package main

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/AndrewBoyarsky/common/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var kafkaConsumer *kafka.Consumer
var kafkaProducer *kafka.Producer

func initKafka() {
	logrus.Infof("Init Consumer/Producer for Kafka")
	producer, ep := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Config.KafkaBrokers,
		// Enable the Idempotent Producer
		"enable.idempotence":     true,
		"transactional.id":       "ProcessedAlbumsTransactional-producer-0 ",
		"batch.size":             10,
		"queue.buffering.max.ms": 2000,
		"request.required.acks":  -1,
	})
	if ep != nil {
		panic(ep)
	}
	consumer, ec := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.Config.KafkaBrokers,
		"group.id":          "processedAlbumsGroup1",
		"fetch.max.bytes":   1000000,
		"fetch.min.bytes":   10,
		"fetch.wait.max.ms": 3000,
		// "auto.offset.reset": "earliest",
		"isolation.level": "read_committed",
		// "allow.auto.create.topics": true,
		"enable.auto.commit": false, // disable auto-commit to make sure we got new albums at least once
	})
	if ec != nil {
		panic(ec)
	}
	consumer.Subscribe(config.Config.NewAlbumsTopic, nil)
	errInitTx := producer.InitTransactions(context.TODO())
	if errInitTx != nil {
		logrus.Fatalf("Unable to init transactional Kafka producer: %s", errInitTx.Error())
	}
	kafkaConsumer = consumer
	kafkaProducer = producer
}

func DoKafkaProcessing() {
	initKafka()
	logrus.Info("Processing Service, start processing messages")
	for {
		doMessageRead()
	}
}

func doMessageRead() {
	// the `ReadMessage` method blocks until we receive the next event
	partitions, _ := kafkaConsumer.Assignment()
	initialOffsets, _ := kafkaConsumer.Position(partitions)
	msg, err := kafkaConsumer.ReadMessage(-1)
	defer func() {
		changedOffsets, _ := kafkaConsumer.Position(partitions)
		partitionsToCommit := []kafka.TopicPartition{}
		for i, o := range changedOffsets {
			if o.Offset == initialOffsets[i].Offset {
				continue
			}
			partitionsToCommit = append(partitionsToCommit, o)
		}
		metadata, _ := kafkaConsumer.GetConsumerGroupMetadata()
		kafkaProducer.SendOffsetsToTransaction(context.TODO(), partitionsToCommit, metadata)
		kafkaProducer.CommitTransaction(context.TODO())
	}()

	if err != nil {
		panic("could not read message " + err.Error())
	}
	kafkaProducer.BeginTransaction()
	// after receiving the message, log its value
	logrus.Infof("Processing Service received kafka message [NewAlbums]: %s, Id: %s", string(msg.Value), string(msg.Key))

	var alb Album
	errr := json.Unmarshal(msg.Value, &alb)

	if errr != nil {
		logrus.Errorf("unable to read json album from Kafka: %s, err: %s", hex.EncodeToString(msg.Value), err.Error())
	} else {
		logrus.Infof("Processing Service, unmarshalled  album: %v", alb)
		alb.Price *= 2 // do something on model to see results sent back to original producer

		jsonValueBinary, _ := json.Marshal(alb)

		deliveryChan := make(chan kafka.Event)
		kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &config.Config.ProcessedAlbumsTopic},
			Key:            msg.Key,
			Value:          jsonValueBinary,
		}, deliveryChan)

		event := <-deliveryChan // do a synchroinous send
		switch ev := event.(type) {
		case *kafka.Message:
			{
				if ev.TopicPartition.Error != nil {
					logrus.Errorf("Processing Service failed to public %v to Kafka[ProcessedAlbums], reason: %s", alb, ev.TopicPartition.Error)
				} else {
					logrus.Infof("Processing Service published %v to Kafka, [ProcessedAlbums]", alb)
				}
			}
		}
	}
}

type Album struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Price    float64 `json:"price"`
	UserName string  `json:"userName"`
	Status   string  `json:"status"`
}
