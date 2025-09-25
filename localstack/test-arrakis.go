package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/elissonalvesilva/arrakis/pkg/sqs"
)

const (
	// LocalStack configuration
	localStackEndpoint = "http://localhost:4566"
	region             = "us-east-1"
	queueURL           = "http://localhost:4566/000000000000/arrakis-test-queue"
	highVolumeQueueURL = "http://localhost:4566/000000000000/arrakis-high-volume-queue"
)

func main() {
	fmt.Println("🚀 Arrakis SQS LocalStack Test")
	fmt.Println("===============================")

	// Create AWS config for LocalStack
	cfg, err := createLocalStackConfig()
	if err != nil {
		log.Fatalf("Failed to create AWS config: %v", err)
	}

	// Create SQS client with Arrakis
	sqsClient := sqs.NewSQS(&cfg)

	// Enable Arrakis adaptive polling
	sqsClient.EnableArrakis()
	fmt.Println("✅ Arrakis adaptive polling enabled")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received interrupt signal, shutting down...")
		cancel()
	}()

	// Start message processing
	go processMessages(ctx, sqsClient, "Standard Queue", queueURL)
	go processMessages(ctx, sqsClient, "High Volume Queue", highVolumeQueueURL)

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Println("👋 Arrakis test completed")
}

func createLocalStackConfig() (aws.Config, error) {
	// Custom endpoint resolver for LocalStack
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           localStackEndpoint,
			SigningRegion: region,
		}, nil
	})

	// Load config with LocalStack settings
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	return cfg, err
}

func processMessages(ctx context.Context, client *sqs.SQS, queueName, queueURL string) {
	fmt.Printf("🔄 Starting message processing for %s\n", queueName)

	messageCount := 0
	emptyPolls := 0
	lastMessageTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🛑 Stopping message processing for %s\n", queueName)
			return
		default:
			// Poll for messages using Arrakis
			startTime := time.Now()
			output, err := client.ReceiveMessage(ctx, queueURL, 10, nil)
			duration := time.Since(startTime)

			if err != nil {
				fmt.Printf("❌ Error receiving messages from %s: %v\n", queueName, err)
				time.Sleep(5 * time.Second)
				continue
			}

			messagesReceived := len(output.Messages)

			if messagesReceived > 0 {
				messageCount += messagesReceived
				emptyPolls = 0
				lastMessageTime = time.Now()

				fmt.Printf("📬 %s: Received %d messages (Total: %d) - Poll took %v\n",
					queueName, messagesReceived, messageCount, duration.Round(time.Millisecond))

				// Process and delete messages
				for _, message := range output.Messages {
					// Simulate message processing
					time.Sleep(100 * time.Millisecond)

					// Delete processed message
					_, err := client.DeleteMessage(ctx, queueURL, *message.ReceiptHandle)
					if err != nil {
						fmt.Printf("❌ Error deleting message: %v\n", err)
					}
				}
			} else {
				emptyPolls++
				timeSinceLastMessage := time.Since(lastMessageTime)

				if emptyPolls%10 == 1 { // Log every 10th empty poll
					fmt.Printf("⏳ %s: Empty poll #%d - Poll took %v (Last message: %v ago)\n",
						queueName, emptyPolls, duration.Round(time.Millisecond), timeSinceLastMessage.Round(time.Second))
				}
			}

			// Show Arrakis status periodically
			if messageCount > 0 && messageCount%20 == 0 {
				arrakisStatus := "enabled"
				if !client.IsArrakisEnabled() {
					arrakisStatus = "disabled"
				}
				fmt.Printf("🎯 %s: Arrakis status: %s, Total processed: %d\n", queueName, arrakisStatus, messageCount)
			}
		}
	}
}
