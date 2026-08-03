package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "event-service/internal/core/proto"
	kafka "event-service/internal/core/kafka"
	repository "event-service/internal/layers/repository"
	service "event-service/internal/layers/service"
	transport "event-service/internal/layers/transport"

	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		postgresHost = "localhost"
	}
	postgresPort := os.Getenv("POSTGRES_PORT")
	if postgresPort == "" {
		postgresPort = "5432"
	}
	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		postgresUser = "admin"
	}
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		postgresPassword = "0324!"
	}
	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		postgresDB = "main_db"
	}

	dbURL := "postgres://" + postgresUser + ":" + postgresPassword + "@" + postgresHost + ":" + postgresPort + "/" + postgresDB + "?sslmode=disable"

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	kafkaBroker := os.Getenv("SPRING_KAFKA_BOOTSTRAP_SERVERS")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	kafkaBrokers := []string{kafkaBroker}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "event-service"
	}

	producer := kafka.NewProducer(kafkaBrokers, kafkaTopic)
	defer producer.Close()
	log.Printf("Connected to Kafka brokers: %v, topic: %s", kafkaBrokers, kafkaTopic)

	eventRepo := repository.NewPostgresRepository(pool)
	eventService := service.NewEventService(eventRepo, producer)
	grpcHandler := transport.NewGRPCHandler(eventService)

	grpcPort := os.Getenv("EVENT_SERVICE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9096"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEventServiceServer(grpcServer, grpcHandler)

	log.Printf("gRPC server listening on port %s", grpcPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}
