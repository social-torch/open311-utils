#!/usr/bin/env bash
# This file:
#
#  - Creates a AWS Simple Notification Service Topic and Subscribes to it with the specified email addy configures in setup.rc
#
# Usage:
#
#   ./createSNSTopic.sh  
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Exit on error. 
set -o errexit
# Do not allow use of undefined vars. 
set -o nounset

# Create the SNS topic
aws sns create-topic --name ${TOPIC_NAME}

# Subscribe an email address to the newly created topic
aws sns subscribe \
    --topic-arn arn:aws:sns:${AWS_DEFAULT_REGION}:${AWS_ACCOUNT_ID}:${TOPIC_NAME} \
    --protocol email \
    --notification-endpoint ${SUBSCRIBING_EMAIL}
