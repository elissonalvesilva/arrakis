# Arrakis LocalStack Testing Environment

This directory contains a complete LocalStack setup for testing the Arrakis adaptive SQS polling library in a local environment.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Go 1.19+ installed
- `awscli` with `awscli-local` plugin (optional, for manual testing)

### One-Command Setup
```bash
./quick-start.sh
```

This script will:
- ✅ Check prerequisites
- 🚀 Start LocalStack with SQS
- 📋 Create test queues
- 🎯 Provide next steps

### Manual Setup
```bash
# Start LocalStack
make start

# Run Arrakis test (in one terminal)
make test-basic

# Send test messages (in another terminal) 
make send-messages
```

## 📁 Directory Structure

```
localstack/
├── docker-compose.yml          # LocalStack configuration
├── Makefile                   # Convenient commands
├── quick-start.sh             # One-command setup
├── test-arrakis.go           # Go test application
├── init-scripts/             # LocalStack initialization
│   └── 01-setup-sqs.sh      # SQS queue creation
├── scripts/
│   └── send-messages.sh     # Interactive message sender
└── README.md                # This file
```

## 🎯 Available Commands

| Command | Description |
|---------|-------------|
| `make start` | Start LocalStack with SQS |
| `make stop` | Stop LocalStack |
| `make status` | Check LocalStack and queue status |
| `make logs` | Show LocalStack logs |
| `make test-basic` | Run basic Arrakis test |
| `make send-messages` | Interactive message sender |
| `make purge` | Clear all SQS queues |
| `make clean` | Stop and remove everything |

## 📬 Test Queues

The setup creates three SQS queues for different testing scenarios:

### 1. Standard Queue
- **URL**: `http://localhost:4566/000000000000/arrakis-test-queue`
- **Purpose**: General testing and basic functionality
- **Configuration**: Default SQS settings

### 2. High Volume Queue  
- **URL**: `http://localhost:4566/000000000000/arrakis-high-volume-queue`
- **Purpose**: Testing high-throughput scenarios
- **Configuration**: Optimized for volume testing

### 3. FIFO Queue
- **URL**: `http://localhost:4566/000000000000/arrakis-fifo-queue.fifo`
- **Purpose**: Testing ordered message processing
- **Configuration**: FIFO with content-based deduplication

## 🧪 Testing Scenarios

### 1. Basic Functionality Test
```bash
make test-basic
```
Tests basic Arrakis polling with standard queue.

### 2. Volume Pattern Testing
The message sender provides several patterns to test Arrakis adaptation:

- **Low Volume**: 5 messages
- **Medium Volume**: 15 messages  
- **High Volume**: 50 messages
- **Very High Volume**: 100 messages
- **Gradual Increase**: 1→3→5→10→20 messages
- **Decrease Pattern**: 20→10→5→3→1 messages

### 3. Manual Testing
```bash
# Send specific message patterns
./scripts/send-messages.sh

# Monitor queue metrics
awslocal sqs get-queue-attributes \
  --queue-url http://localhost:4566/000000000000/arrakis-test-queue \
  --attribute-names All
```

## 📊 Observing Arrakis Behavior

When running tests, watch for these Arrakis behaviors:

### Adaptive Polling Intervals
- **Idle**: 20 second waits when no messages
- **Low Volume**: 15 second waits for <2 messages
- **Medium Volume**: 10 second waits for 2-5 messages
- **High Volume**: 5 second waits for 5-10 messages
- **Very High Volume**: 1 second waits for >10 messages

### EWMA Calculations
The test output shows how the EWMA (Exponentially Weighted Moving Average) adapts to message patterns:

```
📬 Standard Queue: Received 5 messages (Total: 25) - Poll took 243ms
⏳ Standard Queue: Empty poll #3 - Poll took 187ms (Last message: 45s ago)
🎯 Standard Queue: Arrakis status: enabled, Total processed: 40
```

### Expected Behaviors
- **Quick Adaptation**: Fast response to volume increases
- **Gradual Decay**: Slower transition to longer waits during idle
- **Spike Protection**: Prevents single large batches from skewing averages
- **Recovery**: Quick return to fast polling after idle periods

## 🐛 Troubleshooting

### LocalStack Not Starting
```bash
# Check Docker status
docker info

# View LocalStack logs
make logs

# Restart LocalStack
make stop && make start
```

### SQS Connection Issues
```bash
# Verify LocalStack health
curl http://localhost:4566/health

# Check queue existence
docker-compose exec localstack awslocal sqs list-queues
```

### Go Module Issues
```bash
# Run from project root
cd ..
go mod tidy
go run localstack/test-arrakis.go
```

## 🔧 Configuration

### LocalStack Settings
The `docker-compose.yml` configures:
- SQS service only (lightweight)
- Persistent data storage
- Debug logging enabled
- Port 4566 exposed

### Test Application Settings
The `test-arrakis.go` configures:
- LocalStack endpoint
- Fake AWS credentials
- Concurrent queue processing
- Graceful shutdown handling

### Message Sender Settings
The message sender script provides:
- Various volume patterns
- Message attributes testing
- Queue status monitoring
- Interactive menu interface

## 💡 Tips for Testing

1. **Start Simple**: Begin with basic functionality test
2. **Use Two Terminals**: Run test in one, send messages in another
3. **Watch Patterns**: Observe how Arrakis adapts to different volumes
4. **Test Edge Cases**: Try empty queues, sudden spikes, gradual changes
5. **Monitor Logs**: Look for EWMA values and wait time adjustments
6. **Clean Between Tests**: Use `make purge` to reset queue state

## 🎓 Learning Objectives

This testing environment helps you understand:
- How Arrakis adapts polling intervals based on message volume
- EWMA algorithm behavior with real message patterns
- Cost optimization through reduced API calls during idle periods
- Performance benefits during high-volume scenarios
- Trade-offs between latency and API cost efficiency

## 🤝 Contributing

To extend the testing environment:
1. Add new test scenarios in `test-arrakis.go`
2. Create additional message patterns in `send-messages.sh`
3. Add new queues in `init-scripts/01-setup-sqs.sh`
4. Extend the Makefile with new commands

Happy testing with Arrakis! 🚀