#!/usr/bin/env bash
# This file:
#
#  - Accociate lambda function with event source (dynamoDB stream)
#
# Usage:
#
#   ./createTrigger.sh
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Exit on error. 
set -o errexit
# Do not allow use of undefined vars. 
set -o nounset

# Get Stream ARN (depends on jq being installed) 
STREAM_ARN=$(aws dynamodb describe-table --table-name ${FEEDBACK_TABLE_NAME} | jq -r '.Table.LatestStreamArn')

# Create the trigger
aws lambda create-event-source-mapping \
    --region ${AWS_DEFAULT_REGION} \
    --function-name ${LAMBDA_FUNCTION_NAME} \
    --event-source ${STREAM_ARN}   \
    --batch-size 1 \
    --starting-position TRIM_HORIZON