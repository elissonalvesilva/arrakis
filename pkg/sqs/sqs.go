package sqs

// Package sqs provides an enhanced Amazon SQS client with adaptive polling capabilities.
//
// The Arrakis SQS client implements intelligent message polling that automatically adjusts
// polling intervals based on message volume using EWMA (Exponentially Weighted Moving Average).
// This optimization reduces unnecessary API calls during low-traffic periods while maintaining
// responsiveness during high-traffic scenarios.
//
// Key Features:
// - Adaptive polling with EWMA-based volume detection
// - Configurable wait times for different volume scenarios
// - Automatic EWMA decay during idle periods
// - Drop detection and recovery mechanisms
// - Standard SQS operations with enhanced polling intelligence
//
// Example usage:
//
//	awsConfig := aws.Config{...}
//	sqsClient := sqs.NewSQS(&awsConfig)
//	sqsClient.EnableArrakis()
//	messages, err := sqsClient.ReceiveMessage(ctx, queueURL, 10, nil)

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/elissonalvesilva/arrakis/pkg/internal/infra/utils"
)

// Default SQS configuration values
const (
	// _defaultNumberOfMessages is the default maximum number of messages to retrieve in a single poll
	_defaultNumberOfMessages = 10
)

// Default adaptive polling configuration values
const (
	// Wait time defaults for different message volume scenarios
	_defaultIdleWaitTimeSeconds           = 20 // Maximum SQS long polling wait time
	_defaultVisibilityTimeout             = 30 // Standard visibility timeout
	_defaultLowVolumeWaitTimeSeconds      = 15 // Wait time for low volume (< 2 messages)
	_defaultMediumVolumeWaitTimeSeconds   = 10 // Wait time for medium volume (2-5 messages)
	_defaultHighVolumeWaitTimeSeconds     = 5  // Wait time for high volume (5-10 messages)
	_defaultVeryHighVolumeWaitTimeSeconds = 1  // Wait time for very high volume (>10 messages)

	// EWMA algorithm defaults
	_defaultEwmaAlpha              = 0.3   // EWMA smoothing factor (balanced responsiveness)
	_defaultDropDetectionThreshold = 10    // Cycles before EWMA reset on volume drop
	_defaultEnableAdaptivePolling  = false // Adaptive polling disabled by default

	// EWMA calculation thresholds
	_lowVolumeMessageThreshold = 2   // Threshold to consider a cycle as low volume
	_ewmaResetAverageThreshold = 1.0 // EWMA average threshold for reset eligibility
	_minResetIntervalMinutes   = 1   // Minimum time between EWMA resets
	_consecutiveEmptyThreshold = 2   // Empty responses before triggering EWMA decay

	// Volume classification thresholds for wait time calculation
	_lowVolumeThreshold    = 2  // Threshold between idle and low volume
	_mediumVolumeThreshold = 5  // Threshold between low and medium volume
	_highVolumeThreshold   = 10 // Threshold between medium and high volume

	// EWMA decay configuration for idle period handling
	_minDecayGapSeconds = 2    // Minimum time gap before applying decay
	_halfLifeSeconds    = 30.0 // Half-life for exponential decay calculation
	_ewmaDecayThreshold = 0.2  // Threshold below which EWMA is reset to zero
)

// SQS represents an enhanced Amazon SQS client with adaptive polling capabilities.
// It wraps the standard AWS SQS client and adds intelligent polling features through
// the Arrakis adaptive polling algorithm.
type SQS struct {
	client *sqs.Client // The underlying AWS SQS client
	config config      // Configuration for SQS operations and adaptive polling
}

// NewSQS creates a new enhanced SQS client with adaptive polling capabilities.
// The client is initialized with sensible defaults but adaptive polling is disabled by default.
// Use EnableArrakis() to activate the adaptive polling features.
//
// Parameters:
//   - awsconfig: AWS configuration containing credentials, region, and other AWS-specific settings
//
// Returns:
//   - *SQS: A new SQS client instance with adaptive polling capabilities
//
// Example:
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	sqsClient := NewSQS(&cfg)
//	sqsClient.EnableArrakis()
func NewSQS(awsconfig *aws.Config) *SQS {
	var config config

	// Set default values for all configuration parameters
	setDefaults(&config)
	// Apply any provided options

	return &SQS{
		client: sqs.NewFromConfig(*awsconfig),
		config: config,
	}
}

// NewSQSWithOptions creates a new enhanced SQS client with adaptive polling capabilities.
// The client is initialized with sensible defaults but adaptive polling is disabled by default.
// Use EnableArrakis() to activate the adaptive polling features.
//
// Parameters:
//   - awsconfig: AWS configuration containing credentials, region, and other AWS-specific settings
//   - options: A list of functional options to configure the client
//
// Returns:
//   - *SQS: A new SQS client instance with adaptive polling capabilities
//
// Example:
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	sqsClient := NewSQSWithOptions(&cfg, option1, option2)
//	sqsClient.EnableArrakis()
func NewSQSWithOptions(awsconfig *aws.Config, options ...Option) *SQS {
	var config config

	// Set default values for all configuration parameters
	setDefaults(&config)

	// Apply any provided options
	for _, opt := range options {
		opt(&config)
	}

	return &SQS{
		client: sqs.NewFromConfig(*awsconfig),
		config: config,
	}
}

// EnableArrakis activates the adaptive polling algorithm for this SQS client.
// When enabled, the client will automatically adjust polling intervals based on message volume
// using EWMA calculations to optimize API usage and responsiveness.
//
// This should be called after creating the SQS client if you want to use adaptive polling.
// The algorithm starts learning message patterns immediately upon activation.
func (s *SQS) EnableArrakis() {
	s.config.AdaptivePolling.EnableAdaptivePolling = true
}

// DisableArrakis deactivates the adaptive polling algorithm for this SQS client.
// When disabled, the client will use standard SQS polling without any wait time optimizations.
// The EWMA state is preserved and will resume if adaptive polling is re-enabled.
func (s *SQS) DisableArrakis() {
	s.config.AdaptivePolling.EnableAdaptivePolling = false
}

// IsArrakisEnabled returns the current state of the adaptive polling algorithm.
//
// Returns:
//   - bool: true if adaptive polling is active, false otherwise
func (s *SQS) IsArrakisEnabled() bool {
	return s.config.AdaptivePolling.EnableAdaptivePolling
}

// ReceiveMessage retrieves messages from the specified SQS queue with optional adaptive polling.
// When Arrakis is enabled, this method automatically calculates optimal wait times based on
// historical message volume patterns using EWMA. The response is analyzed to update the
// adaptive polling algorithm's state for future optimizations.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - queueURL: The URL of the SQS queue to receive messages from
//   - maxMsg: Maximum number of messages to retrieve (1-10). If 0, defaults to 10
//   - messageAttributes: Map of message attribute names to retrieve. Keys become attribute names
//
// Returns:
//   - *sqs.ReceiveMessageOutput: The SQS response containing received messages
//   - error: Any error that occurred during the operation
//
// Example:
//
//	messages, err := sqsClient.ReceiveMessage(ctx, "https://sqs.us-east-1.amazonaws.com/123456789012/myqueue", 5, map[string]string{"Author": "", "Timestamp": ""})
//	if err != nil {
//	    log.Printf("Error receiving messages: %v", err)
//	    return
//	}
//	fmt.Printf("Received %d messages\n", len(messages.Messages))
func (s *SQS) ReceiveMessage(ctx context.Context, queueURL string, maxMsg int32, messageAttributes map[string]string) (*sqs.ReceiveMessageOutput, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueURL),
		MaxNumberOfMessages:   utils.GetOrDefault(maxMsg, _defaultNumberOfMessages).(int32),
		VisibilityTimeout:     int32(s.config.VisibilityTimeout),
		MessageAttributeNames: utils.MapKeys(messageAttributes),
	}

	// Apply adaptive polling wait time if Arrakis is enabled
	if s.IsArrakisEnabled() {
		input.WaitTimeSeconds = int32(s.calculateWaitTime())
	}

	output, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, err
	}

	// Update adaptive polling algorithm with the response
	s.handleReceiveResponse(output)

	return output, nil
}

// DeleteMessage removes a message from the specified SQS queue using its receipt handle.
// This is a standard SQS operation that is not affected by the adaptive polling algorithm.
// Messages should be deleted after successful processing to prevent redelivery.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - queueURL: The URL of the SQS queue containing the message
//   - receiptHandle: The receipt handle of the message to delete (obtained from ReceiveMessage)
//
// Returns:
//   - *sqs.DeleteMessageOutput: The SQS response confirming message deletion
//   - error: Any error that occurred during the operation
//
// Example:
//
//	for _, message := range messages.Messages {
//	    // Process the message...
//	    _, err := sqsClient.DeleteMessage(ctx, queueURL, *message.ReceiptHandle)
//	    if err != nil {
//	        log.Printf("Error deleting message: %v", err)
//	    }
//	}
func (s *SQS) DeleteMessage(ctx context.Context, queueURL string, receiptHandle string) (*sqs.DeleteMessageOutput, error) {
	output, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
