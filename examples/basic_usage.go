package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/elissonalvesilva/arrakis/pkg/sqs"
)

func main() {
	// Example: Basic usage of Arrakis SQS client with adaptive polling
	basicExample()

	// Example: Advanced configuration with custom parameters
	advancedExample()

	// Example: Processing messages in a loop
	messageProcessingLoop()
}

// basicExample demonstrates the simplest way to use Arrakis with default settings
func basicExample() {
	fmt.Println("=== Basic Example ===")

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return
	}

	// Create SQS client with Arrakis adaptive polling
	sqsClient := sqs.NewSQS(&cfg)

	// Enable adaptive polling - this is where the magic happens!
	sqsClient.EnableArrakis()

	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue"

	// Receive messages - Arrakis will automatically optimize polling intervals
	ctx := context.Background()
	messages, err := sqsClient.ReceiveMessage(ctx, queueURL, 10, nil)
	if err != nil {
		log.Printf("Error receiving messages: %v", err)
		return
	}

	fmt.Printf("Received %d messages using adaptive polling\n", len(messages.Messages))

	// Process and delete messages
	for _, message := range messages.Messages {
		fmt.Printf("Processing message: %s\n", *message.Body)

		// Delete message after processing
		_, err := sqsClient.DeleteMessage(ctx, queueURL, *message.ReceiptHandle)
		if err != nil {
			log.Printf("Error deleting message: %v", err)
		}
	}
}

// advancedExample shows how to configure Arrakis with custom parameters
func advancedExample() {
	fmt.Println("\n=== Advanced Configuration Example ===")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return
	}

	// Configure custom adaptive polling parameters
	// These values are tuned for a high-throughput scenario
	option := sqs.WithAdaptivePolling(
		20,  // idleWaitTimeSeconds - wait time when no messages
		60,  // visibilityTimeout - how long messages stay hidden
		12,  // lowVolumeWaitTimeSeconds - wait time for low volume
		8,   // mediumVolumeWaitTimeSeconds - wait time for medium volume
		4,   // highVolumeWaitTimeSeconds - wait time for high volume
		1,   // veryHighVolumeWaitTimeSeconds - wait time for very high volume
		0.4, // ewmaAlpha - higher value = more responsive to changes
		8,   // dropDetectionThreshold - cycles before resetting EWMA
	)

	// Create SQS client
	sqsClient := sqs.NewSQSWithOptions(&cfg, option)

	// Apply configuration (this would typically be done during initialization)
	// For demonstration, we'll show the function signature
	fmt.Printf("Custom configuration applied with EWMA alpha: 0.4 (more responsive)\n")
	fmt.Printf("Wait times: Idle=20s, Low=12s, Medium=8s, High=4s, VeryHigh=1s\n")

	sqsClient.EnableArrakis()

	fmt.Printf("Arrakis adaptive polling enabled: %t\n", sqsClient.IsArrakisEnabled())
}

// messageProcessingLoop demonstrates continuous message processing with adaptive polling
func messageProcessingLoop() {
	fmt.Println("\n=== Message Processing Loop Example ===")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return
	}

	sqsClient := sqs.NewSQS(&cfg)
	sqsClient.EnableArrakis()

	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-processing-queue"
	ctx := context.Background()

	fmt.Println("Starting message processing loop (simulated)...")

	// Simulate a processing loop (normally this would run indefinitely)
	for i := 0; i < 3; i++ {
		// Receive messages with adaptive polling
		messages, err := sqsClient.ReceiveMessage(ctx, queueURL, 5, map[string]string{
			"Author":    "",
			"Timestamp": "",
			"MessageId": "",
		})

		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			continue
		}

		if len(messages.Messages) == 0 {
			fmt.Printf("Iteration %d: No messages received (Arrakis will increase wait time)\n", i+1)
		} else {
			fmt.Printf("Iteration %d: Received %d messages (Arrakis will optimize wait time)\n",
				i+1, len(messages.Messages))

			// Process each message
			for j, message := range messages.Messages {
				fmt.Printf("  Message %d: %s\n", j+1, truncateString(*message.Body, 50))

				// Simulate processing time
				// time.Sleep(100 * time.Millisecond)

				// Delete processed message
				_, err := sqsClient.DeleteMessage(ctx, queueURL, *message.ReceiptHandle)
				if err != nil {
					log.Printf("Error deleting message: %v", err)
				}
			}
		}
	}

	fmt.Println("Processing loop completed. Arrakis has learned the message patterns!")
}

// truncateString limits string length for display purposes
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Example output when running this program:
//
// === Basic Example ===
// Received 3 messages using adaptive polling
// Processing message: Hello from SQS message 1
// Processing message: Hello from SQS message 2
// Processing message: Hello from SQS message 3
//
// === Advanced Configuration Example ===
// Custom configuration applied with EWMA alpha: 0.4 (more responsive)
// Wait times: Idle=20s, Low=12s, Medium=8s, High=4s, VeryHigh=1s
// Arrakis adaptive polling enabled: true
//
// === Message Processing Loop Example ===
// Starting message processing loop (simulated)...
// Iteration 1: Received 2 messages (Arrakis will optimize wait time)
//   Message 1: Processing task data: {"id": 12345, "action": "pr...
//   Message 2: Processing task data: {"id": 12346, "action": "pr...
// Iteration 2: No messages received (Arrakis will increase wait time)
// Iteration 3: Received 1 messages (Arrakis will optimize wait time)
//   Message 1: Processing task data: {"id": 12347, "action": "pr...
// Processing loop completed. Arrakis has learned the message patterns!
