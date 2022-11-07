package albums

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/AndrewBoyarsky/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var kafkaConsumer *kafka.Consumer
var kafkaProducer *kafka.Producer

func init() {
	producer, ep := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Config.KafkaBrokers,
		// Enable the Idempotent Producer
		"enable.idempotence":     true,
		"batch.size":             10,
		"queue.buffering.max.ms": 2000,
		"request.required.acks":  -1,
	})
	if ep != nil {
		panic(ep)
	}
	consumer, ec := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        config.Config.KafkaBrokers,
		"group.id":                 "processedAlbumsGroup1",
		"fetch.max.bytes":          1000000,
		"fetch.min.bytes":          10,
		"fetch.wait.max.ms":        3000,
		// "allow.auto.create.topics": true,
		"enable.auto.commit":       false, // disable auto-commit to make sure we processed album at least ones
	})
	if ec != nil {
		panic(ec)
	}

	consumer.Subscribe(config.Config.ProcessedAlbumsTopic, nil)
	kafkaConsumer = consumer
	kafkaProducer = producer
	go consumeProcessedFromKafka()
}

func produceToKafka(album *Album, id string) {
	jsonValueBinary, _ := json.Marshal(album)
	syncChan := make(chan kafka.Event)
	kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &config.Config.NewAlbumsTopic, Partition: kafka.PartitionAny},
		Key:            []byte(id),
		Value:          jsonValueBinary,
	}, syncChan)
	event := <-syncChan
	switch ev := event.(type) {
	case *kafka.Message:
		if ev.TopicPartition.Error != nil {
			logrus.Errorf("Failed to send to Kafka: %v", album)
			panic(ev.TopicPartition.Error)
		} else {
			logrus.Infof("Published %v to Kafka, topic NewAlbums", album)
		}
	}
}

func consumeProcessedFromKafka() {
	for {
		// the `ReadMessage` method blocks until we receive the next event
		msg, err := kafkaConsumer.ReadMessage(-1)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// after receiving the message, log its value
		logrus.Infof("Received: %s, Id: %s", string(msg.Value), string(msg.Key))

		var alb Album
		errr := json.Unmarshal(msg.Value, &alb)

		if errr != nil {
			panic(fmt.Sprintf("unable to read json album from Kafka: %s, err: %s", hex.EncodeToString(msg.Value), err.Error()))
		}
		logrus.Infof("Received processed album: %v", alb)
		repo := NewAlbumRepo()
		alb.Status = "PROCESSING_FINISHED"
		savedAlb := repo.GetById(nil, string(msg.Key), alb.UserName)
		if savedAlb == nil {
			logrus.Errorf("No Album with id=%s found in database. Maybe API server failed during sending it to Kafka without committing a MongoDb transaction", string(msg.Key))
		}
		repo.Save(nil, alb, string(msg.Key))
		_, errCommit := kafkaConsumer.Commit()
		if errCommit != nil {
			logrus.Errorf("Unable to commit Kafka offset for message: %v", alb)
		}
	}
}
