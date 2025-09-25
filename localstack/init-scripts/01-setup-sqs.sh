#!/bin/bash

echo "🚀 Initializing LocalStack for Arrakis SQS testing..."

# Wait for LocalStack to be ready
sleep 2

# Create SQS queues for testing
echo "📬 Creating SQS queues..."

# Standard queue for basic testing
awslocal sqs create-queue \
    --queue-name arrakis-test-queue \
    --region us-east-1

# High throughput queue for volume testing
awslocal sqs create-queue \
    --queue-name arrakis-high-volume-queue \
    --region us-east-1 \
    --attributes VisibilityTimeoutSeconds=30,MessageRetentionPeriod=1209600

# FIFO queue for ordered message testing
awslocal sqs create-queue \
    --queue-name arrakis-fifo-queue.fifo \
    --region us-east-1 \
    --attributes FifoQueue=true,ContentBasedDeduplication=true

echo "✅ SQS queues created successfully!"

# List all queues to confirm
echo "📋 Available queues:"
awslocal sqs list-queues --region us-east-1

echo "🎯 LocalStack initialization complete!"
echo "📍 SQS Endpoint: http://localhost:4566"
echo "🔗 Queue URLs:"
echo "   - Standard: http://localhost:4566/000000000000/arrakis-test-queue"
echo "   - High Volume: http://localhost:4566/000000000000/arrakis-high-volume-queue"
echo "   - FIFO: http://localhost:4566/000000000000/arrakis-fifo-queue.fifo"