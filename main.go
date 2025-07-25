package main

import (
	"os"
	"os/signal"
	"syscall"

	"kafka-consumer/application/logger"
	"kafka-consumer/application/service"
	"kafka-consumer/config"
)

func main() {
	con, _, err := config.GetConfigByEnv()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Initialize logger with config
	logger.Newlogger(con.Logger)
	l := logger.GetLogger()

	// Create Kafka service with dependencies
	kafkaService := service.NewKafkaService(l, con)

	// Start consumer in a goroutine
	go func() {
		if err := kafkaService.StartConsumer(); err != nil {
			l.Errorf("Failed to start consumer: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	l.Info("Received shutdown signal, closing consumer...")

	// Gracefully close the Kafka service
	if err := kafkaService.Close(); err != nil {
		l.Errorf("Error closing Kafka service: %v", err)
	}

	l.Info("Application shutdown complete")
	os.Exit(0)
}
