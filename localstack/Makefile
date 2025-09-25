# Arrakis LocalStack Testing Environment
# =====================================

.PHONY: help start stop status logs test-basic test-volume clean purge send-messages

# Default target
help:
	@echo "🚀 Arrakis LocalStack Testing Environment"
	@echo "========================================"
	@echo ""
	@echo "Available commands:"
	@echo "  start         - Start LocalStack with SQS"
	@echo "  stop          - Stop LocalStack"
	@echo "  status        - Check LocalStack status"
	@echo "  logs          - Show LocalStack logs"
	@echo "  test-basic    - Run basic Arrakis test"
	@echo "  test-volume   - Run volume testing scenarios"
	@echo "  send-messages - Interactive message sender"
	@echo "  purge         - Purge all SQS queues"
	@echo "  clean         - Stop and remove all containers"
	@echo ""

# Start LocalStack
start:
	@echo "🚀 Starting LocalStack for Arrakis testing..."
	docker-compose up -d
	@echo "⏳ Waiting for LocalStack to be ready..."
	@sleep 10
	@echo "✅ LocalStack is ready!"
	@echo "📍 SQS Endpoint: http://localhost:4566"
	@make status

# Stop LocalStack
stop:
	@echo "🛑 Stopping LocalStack..."
	docker-compose down

# Check status
status:
	@echo "📊 LocalStack Status:"
	@docker-compose ps
	@echo ""
	@echo "📋 Available SQS Queues:"
	@docker-compose exec localstack awslocal sqs list-queues --region us-east-1 || echo "❌ LocalStack not ready or no queues found"

# Show logs
logs:
	@echo "📋 LocalStack Logs:"
	docker-compose logs -f localstack

# Run basic test
test-basic:
	@echo "🧪 Running basic Arrakis test..."
	@cd .. && go run localstack/test-arrakis.go

# Run volume testing scenarios  
test-volume:
	@echo "🧪 Running volume testing scenarios..."
	@echo "📤 This will send different message patterns to test Arrakis adaptation"
	@echo "💡 Run 'make send-messages' in another terminal to control message flow"
	@cd .. && go run localstack/test-arrakis.go

# Interactive message sender
send-messages:
	@echo "📤 Starting interactive message sender..."
	@./scripts/send-messages.sh

# Purge all queues
purge:
	@echo "🧹 Purging all SQS queues..."
	@docker-compose exec localstack awslocal sqs purge-queue --queue-url "http://localhost:4566/000000000000/arrakis-test-queue" --region us-east-1 || true
	@docker-compose exec localstack awslocal sqs purge-queue --queue-url "http://localhost:4566/000000000000/arrakis-high-volume-queue" --region us-east-1 || true
	@docker-compose exec localstack awslocal sqs purge-queue --queue-url "http://localhost:4566/000000000000/arrakis-fifo-queue.fifo" --region us-east-1 || true
	@echo "✅ All queues purged!"

# Clean everything
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v --remove-orphans
	@echo "✅ Cleanup completed!"

# Development helpers
dev-setup: start
	@echo "🔧 Setting up development environment..."
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "💡 Quick start:"
	@echo "   1. Run 'make test-basic' to start Arrakis testing"
	@echo "   2. In another terminal, run 'make send-messages' to send test messages"
	@echo "   3. Watch how Arrakis adapts polling intervals based on message volume!"

# Check if LocalStack is healthy
health-check:
	@echo "🏥 Checking LocalStack health..."
	@curl -s http://localhost:4566/health | python3 -m json.tool || echo "❌ LocalStack health check failed"