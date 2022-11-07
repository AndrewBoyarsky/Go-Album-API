module github.com/AndrewBoyarsky/albumprocessor

go 1.19

replace github.com/AndrewBoyarsky/common => ../common

require (
	github.com/segmentio/kafka-go v0.4.35
	github.com/sirupsen/logrus v1.9.0
	github.com/AndrewBoyarsky/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/klauspost/compress v1.15.7 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
