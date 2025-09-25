#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENDPOINT_URL="http://localhost:4566"
QUEUE_URL="http://localhost:4566/000000000000/arrakis-test-queue"
HIGH_VOLUME_QUEUE_URL="http://localhost:4566/000000000000/arrakis-high-volume-queue"
REGION="us-east-1"

echo -e "${BLUE}🚀 Arrakis SQS Message Sender${NC}"
echo -e "${BLUE}================================${NC}"

# Function to send a single message
send_message() {
    local queue_url=$1
    local message_body=$2
    local message_attributes=$3
    
    if [ -z "$message_attributes" ]; then
        awslocal sqs send-message \
            --endpoint-url $ENDPOINT_URL \
            --region $REGION \
            --queue-url "$queue_url" \
            --message-body "$message_body" > /dev/null
    else
        awslocal sqs send-message \
            --endpoint-url $ENDPOINT_URL \
            --region $REGION \
            --queue-url "$queue_url" \
            --message-body "$message_body" \
            --message-attributes "$message_attributes" > /dev/null
    fi
}

# Function to send burst of messages
send_burst() {
    local queue_url=$1
    local count=$2
    local prefix=$3
    
    echo -e "${YELLOW}📤 Sending $count messages to queue...${NC}"
    
    for i in $(seq 1 $count); do
        message_body="{\"id\": $i, \"message\": \"$prefix message $i\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)\", \"data\": \"Sample data for testing Arrakis adaptive polling algorithm\"}"
        
        # Add message attributes for some messages
        if [ $((i % 3)) -eq 0 ]; then
            attributes='{"Priority":{"DataType":"String","StringValue":"High"},"Source":{"DataType":"String","StringValue":"ArrakisTest"}}'
            send_message "$queue_url" "$message_body" "$attributes"
        else
            send_message "$queue_url" "$message_body"
        fi
        
        # Progress indicator
        if [ $((i % 10)) -eq 0 ]; then
            echo -e "${GREEN}  ✓ Sent $i/$count messages${NC}"
        fi
    done
    
    echo -e "${GREEN}✅ Successfully sent $count messages!${NC}"
}

# Menu function
show_menu() {
    echo ""
    echo -e "${BLUE}Select an option:${NC}"
    echo "1) Send single message"
    echo "2) Send low volume burst (5 messages)"
    echo "3) Send medium volume burst (15 messages)"
    echo "4) Send high volume burst (50 messages)"
    echo "5) Send very high volume burst (100 messages)"
    echo "6) Send gradual increase pattern"
    echo "7) Send decrease pattern"
    echo "8) Check queue status"
    echo "9) Purge all queues"
    echo "0) Exit"
    echo ""
}

# Function to check queue status
check_queue_status() {
    echo -e "${YELLOW}📊 Checking queue status...${NC}"
    
    echo -e "${BLUE}Standard Queue:${NC}"
    awslocal sqs get-queue-attributes \
        --endpoint-url $ENDPOINT_URL \
        --region $REGION \
        --queue-url "$QUEUE_URL" \
        --attribute-names ApproximateNumberOfMessages,ApproximateNumberOfMessagesNotVisible
    
    echo -e "${BLUE}High Volume Queue:${NC}"
    awslocal sqs get-queue-attributes \
        --endpoint-url $ENDPOINT_URL \
        --region $REGION \
        --queue-url "$HIGH_VOLUME_QUEUE_URL" \
        --attribute-names ApproximateNumberOfMessages,ApproximateNumberOfMessagesNotVisible
}

# Function to purge queues
purge_queues() {
    echo -e "${RED}🧹 Purging all queues...${NC}"
    
    awslocal sqs purge-queue \
        --endpoint-url $ENDPOINT_URL \
        --region $REGION \
        --queue-url "$QUEUE_URL"
    
    awslocal sqs purge-queue \
        --endpoint-url $ENDPOINT_URL \
        --region $REGION \
        --queue-url "$HIGH_VOLUME_QUEUE_URL"
    
    echo -e "${GREEN}✅ All queues purged!${NC}"
}

# Function to send gradual increase pattern
send_gradual_increase() {
    echo -e "${YELLOW}📈 Sending gradual increase pattern...${NC}"
    
    # Start with 1 message, then 3, 5, 10, 20
    volumes=(1 3 5 10 20)
    
    for volume in "${volumes[@]}"; do
        echo -e "${BLUE}Sending $volume messages...${NC}"
        send_burst "$QUEUE_URL" $volume "GradualIncrease"
        echo -e "${YELLOW}Waiting 30 seconds before next batch...${NC}"
        sleep 30
    done
    
    echo -e "${GREEN}✅ Gradual increase pattern completed!${NC}"
}

# Function to send decrease pattern
send_decrease_pattern() {
    echo -e "${YELLOW}📉 Sending decrease pattern...${NC}"
    
    # Start with 20 messages, then 10, 5, 3, 1
    volumes=(20 10 5 3 1)
    
    for volume in "${volumes[@]}"; do
        echo -e "${BLUE}Sending $volume messages...${NC}"
        send_burst "$QUEUE_URL" $volume "DecreasePattern"
        echo -e "${YELLOW}Waiting 30 seconds before next batch...${NC}"
        sleep 30
    done
    
    echo -e "${GREEN}✅ Decrease pattern completed!${NC}"
}

# Main loop
while true; do
    show_menu
    read -p "Enter your choice: " choice
    
    case $choice in
        1)
            message_body="{\"id\": 1, \"message\": \"Single test message\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)\"}"
            send_message "$QUEUE_URL" "$message_body"
            echo -e "${GREEN}✅ Single message sent!${NC}"
            ;;
        2)
            send_burst "$QUEUE_URL" 5 "LowVolume"
            ;;
        3)
            send_burst "$QUEUE_URL" 15 "MediumVolume"
            ;;
        4)
            send_burst "$QUEUE_URL" 50 "HighVolume"
            ;;
        5)
            send_burst "$HIGH_VOLUME_QUEUE_URL" 100 "VeryHighVolume"
            ;;
        6)
            send_gradual_increase
            ;;
        7)
            send_decrease_pattern
            ;;
        8)
            check_queue_status
            ;;
        9)
            purge_queues
            ;;
        0)
            echo -e "${GREEN}👋 Goodbye!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Invalid option. Please try again.${NC}"
            ;;
    esac
done