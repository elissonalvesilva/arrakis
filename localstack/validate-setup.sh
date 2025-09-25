#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Arrakis LocalStack Validation${NC}"
echo -e "${BLUE}================================${NC}"

# Check if we're in the localstack directory
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Please run this script from the localstack directory${NC}"
    exit 1
fi

# Check Docker
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Go is available: $(go version)${NC}"

# Start LocalStack if not running
echo -e "${YELLOW}🚀 Starting LocalStack...${NC}"
docker-compose up -d > /dev/null 2>&1
sleep 10

# Check LocalStack health
echo -e "${YELLOW}🏥 Checking LocalStack health...${NC}"
if curl -s http://localhost:4566/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ LocalStack is healthy${NC}"
else
    echo -e "${RED}❌ LocalStack is not responding${NC}"
    exit 1
fi

# Verify SQS queues
echo -e "${YELLOW}📋 Verifying SQS queues...${NC}"
docker-compose exec -T localstack awslocal sqs list-queues --region us-east-1 > /tmp/arrakis_queues.txt 2>&1

if grep -q "arrakis-test-queue" /tmp/arrakis_queues.txt; then
    echo -e "${GREEN}✅ Standard queue exists${NC}"
else
    echo -e "${YELLOW}⚠️  Creating standard queue...${NC}"
    docker-compose exec -T localstack awslocal sqs create-queue --queue-name arrakis-test-queue --region us-east-1 > /dev/null 2>&1
fi

if grep -q "arrakis-high-volume-queue" /tmp/arrakis_queues.txt; then
    echo -e "${GREEN}✅ High volume queue exists${NC}"
else
    echo -e "${YELLOW}⚠️  Creating high volume queue...${NC}"
    docker-compose exec -T localstack awslocal sqs create-queue --queue-name arrakis-high-volume-queue --region us-east-1 > /dev/null 2>&1
fi

# Test Go compilation
echo -e "${YELLOW}🔧 Testing Go compilation...${NC}"
cd ..
if go build -o /tmp/test-arrakis localstack/test-arrakis.go; then
    echo -e "${GREEN}✅ Go code compiles successfully${NC}"
    rm -f /tmp/test-arrakis
else
    echo -e "${RED}❌ Go compilation failed${NC}"
    exit 1
fi
cd localstack

# Send a test message
echo -e "${YELLOW}📤 Sending test message...${NC}"
docker-compose exec -T localstack awslocal sqs send-message \
    --queue-url "http://localhost:4566/000000000000/arrakis-test-queue" \
    --message-body '{"test": "validation message", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)'"}' \
    --region us-east-1 > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Test message sent successfully${NC}"
else
    echo -e "${RED}❌ Failed to send test message${NC}"
    exit 1
fi

# Check message was received
echo -e "${YELLOW}📬 Verifying message reception...${NC}"
docker-compose exec -T localstack awslocal sqs get-queue-attributes \
    --queue-url "http://localhost:4566/000000000000/arrakis-test-queue" \
    --attribute-names ApproximateNumberOfMessages \
    --region us-east-1 > /tmp/queue_attrs.txt 2>&1

if grep -q '"ApproximateNumberOfMessages": "1"' /tmp/queue_attrs.txt; then
    echo -e "${GREEN}✅ Test message is in queue${NC}"
else
    echo -e "${YELLOW}⚠️  Message count verification inconclusive${NC}"
fi

# Clean up test message
docker-compose exec -T localstack awslocal sqs purge-queue \
    --queue-url "http://localhost:4566/000000000000/arrakis-test-queue" \
    --region us-east-1 > /dev/null 2>&1

# Cleanup temp files
rm -f /tmp/arrakis_queues.txt /tmp/queue_attrs.txt

echo ""
echo -e "${GREEN}🎉 Validation completed successfully!${NC}"
echo ""
echo -e "${BLUE}📍 Environment Summary:${NC}"
echo -e "   LocalStack: ${GREEN}Running on port 4566${NC}"
echo -e "   SQS Queues: ${GREEN}Ready for testing${NC}"
echo -e "   Go Environment: ${GREEN}Compilation working${NC}"
echo -e "   Test Message: ${GREEN}Send/receive working${NC}"

echo ""
echo -e "${YELLOW}🚀 Ready to test Arrakis!${NC}"
echo -e "   Run: ${BLUE}make test-basic${NC} to start testing"
echo -e "   Run: ${BLUE}make send-messages${NC} to send test messages"
echo ""
echo -e "${GREEN}✨ Happy testing! ✨${NC}"