# Arrakis - Adaptive SQS Polling Library for Go

A Go library that implements adaptive polling for Amazon SQS using EWMA (Exponentially Weighted Moving Average) algorithms. Arrakis automatically optimizes polling intervals based on message volume, reducing API costs and improving responsiveness.

## 🚀 Features

- 🎯 **Intelligent Adaptive Polling**: Automatically adjusts SQS polling intervals
- 📊 **EWMA Algorithm**: Uses exponentially weighted moving average to detect volume patterns
- 💰 **Cost Optimization**: Reduces unnecessary API calls during idle periods
- ⚡ **Low Latency**: Frequent polling during traffic spikes
- 🛡️ **Spike Protection**: Prevents distortions caused by outliers
- 🔍 **Drop Detection**: Quickly adapts to reductions in message volume
- 📈 **Temporal Decay**: Gradually reduces frequency during idle periods
- ⚙️ **Highly Configurable**: Adjustable parameters for different scenarios

## 📦 Installation

```bash
go get github.com/elissonalvesilva/arrakis
```

## 🎯 Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/elissonalvesilva/arrakis/pkg/sqs"
)

func main() {
    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Create SQS client with Arrakis
    sqsClient := sqs.NewSQS(&cfg)
    
    // Enable adaptive polling - this is where the magic happens!
    sqsClient.EnableArrakis()
    
    // Use normally - Arrakis optimizes automatically
    queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue"
    messages, err := sqsClient.ReceiveMessage(context.Background(), queueURL, 10, nil)
    if err != nil {
        log.Printf("Error receiving messages: %v", err)
        return
    }
    
    log.Printf("Received %d messages with adaptive polling", len(messages.Messages))
}
```

## ⚙️ Advanced Configuration

```go
// Custom configuration for high volume scenarios
option := sqs.WithAdaptivePolling(
    20, // idleWaitTimeSeconds - wait time when there are no messages
    60, // visibilityTimeout - message visibility timeout
    12, // lowVolumeWaitTimeSeconds - wait for low volume
    8,  // mediumVolumeWaitTimeSeconds - wait for medium volume
    4,  // highVolumeWaitTimeSeconds - wait for high volume
    1,  // veryHighVolumeWaitTimeSeconds - wait for very high volume
    0.4, // ewmaAlpha - smoothing factor (more responsive)
    8,   // dropDetectionThreshold - cycles before EWMA reset
)

// Apply custom configuration
sqsClient := sqs.NewSQS(&cfg, option)
sqsClient.EnableArrakis()
```

## 📊 How It Works

Arrakis automatically classifies message volume into categories and adjusts polling intervals:

| Volume | Criteria (EWMA) | Wait Time | Scenario |
|--------|---------------|-----------|----------|
| **Idle** | = 0 messages | 20 seconds | Empty queue |
| **Low** | < 2 messages | 15 seconds | Low traffic |
| **Medium** | 2-5 messages | 10 seconds | Moderate traffic |
| **High** | 5-10 messages | 5 seconds | Heavy traffic |
| **Very High** | > 10 messages | 1 second | Traffic spike |

### EWMA Algorithm
```
new_value = α × current_observation + (1-α) × previous_value
```

- **Low α (0.1-0.3)**: More stable, gradual changes
- **High α (0.4-0.7)**: More responsive, rapid adaptation
- **Recommended**: 0.2-0.4 for most cases

## 📈 Benefits

### Cost Reduction
- **Up to 70% fewer API calls** during idle periods
- Intelligent polling based on real traffic patterns
- Prevention of unnecessary polling

### Performance Improvement
- **Reduced latency** during traffic spikes
- Automatic adaptation to volume changes
- Elimination of manual interval configuration

### Reliability
- Protection against isolated spikes
- Automatic detection of volume drops
- Quick recovery after idle periods

## 🏗️ Project Structure

```
arrakis/
├── pkg/sqs/                    # Public library API
│   ├── sqs.go                 # Main SQS client
│   ├── arrakis.go             # Adaptive polling algorithm
│   ├── options.go             # Configuration and options
│   └── sqs_test.go            # Unit tests
├── pkg/internal/infra/utils/   # Internal utilities
├── examples/                   # Usage examples
└── docs/                      # Technical documentation
```

## 📚 Documentation

- [Technical Documentation](docs/TECHNICAL.md) - EWMA algorithm details
- [Usage Examples](examples/) - Practical use cases
- [API Reference](docs/API.md) - Complete API reference

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark
```

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Inspiration

The name "Arrakis" is a tribute to the planet from the Dune universe, known for its valuable resources and the need for optimization for survival - just like this library optimizes SQS resource usage.