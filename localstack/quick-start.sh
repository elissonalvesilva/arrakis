#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "🚀 Arrakis LocalStack Quick Start"
echo "================================="
echo -e "${NC}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Please run this script from the localstack directory.${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Prerequisites check...${NC}"
echo -e "${GREEN}✅ Docker is running${NC}"
echo -e "${GREEN}✅ In correct directory${NC}"

echo ""
echo -e "${YELLOW}🏗️  Setting up LocalStack environment...${NC}"

# Start LocalStack
echo -e "${BLUE}🚀 Starting LocalStack...${NC}"
docker-compose up -d

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to start LocalStack${NC}"
    exit 1
fi

echo -e "${GREEN}✅ LocalStack container started${NC}"

# Wait for LocalStack to be ready
echo -e "${YELLOW}⏳ Waiting for LocalStack to initialize...${NC}"
sleep 15

# Check if LocalStack is ready
echo -e "${BLUE}🏥 Checking LocalStack health...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:4566/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ LocalStack is healthy!${NC}"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ LocalStack failed to start properly${NC}"
        echo -e "${YELLOW}💡 Try running: docker-compose logs localstack${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}   Attempt $i/10...${NC}"
    sleep 3
done

# Verify SQS queues were created
echo -e "${BLUE}📋 Verifying SQS queues...${NC}"
docker-compose exec -T localstack awslocal sqs list-queues --region us-east-1 > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ SQS queues are ready${NC}"
else
    echo -e "${YELLOW}⚠️  SQS queues not found, this is normal on first run${NC}"
fi

echo ""
echo -e "${GREEN}🎉 LocalStack setup completed successfully!${NC}"
echo ""
echo -e "${PURPLE}📍 Service Information:${NC}"
echo -e "   LocalStack Endpoint: ${BLUE}http://localhost:4566${NC}"
echo -e "   SQS Standard Queue: ${BLUE}http://localhost:4566/000000000000/arrakis-test-queue${NC}"
echo -e "   SQS High Volume Queue: ${BLUE}http://localhost:4566/000000000000/arrakis-high-volume-queue${NC}"
echo -e "   SQS FIFO Queue: ${BLUE}http://localhost:4566/000000000000/arrakis-fifo-queue.fifo${NC}"

echo ""
echo -e "${PURPLE}🚀 Next Steps:${NC}"
echo -e "   1. Test Arrakis: ${YELLOW}make test-basic${NC}"
echo -e "   2. Send messages: ${YELLOW}make send-messages${NC}"
echo -e "   3. Check status: ${YELLOW}make status${NC}"
echo -e "   4. View logs: ${YELLOW}make logs${NC}"
echo -e "   5. Stop LocalStack: ${YELLOW}make stop${NC}"

echo ""
echo -e "${BLUE}💡 Pro Tips:${NC}"
echo -e "   • Run the test and message sender in separate terminals"
echo -e "   • Watch how Arrakis adapts polling intervals to message volume"
echo -e "   • Try different message patterns to see adaptive behavior"
echo -e "   • Use 'make purge' to clear queues between tests"

echo ""
echo -e "${GREEN}✨ Happy testing with Arrakis! ✨${NC}"