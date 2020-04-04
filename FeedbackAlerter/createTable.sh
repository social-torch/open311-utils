#!/usr/bin/env bash
# This file:
#
#  - Creates a DynamoDB Table with a Stream Enabled.  Only do this if the table doesn't alread exist.
#       For example, don't do this in Social Torch Production.
#
# Usage:
#
#   ./createTable.sh  
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Exit on error. 
set -o errexit
# Exit on error inside any functions or subshells.
set -o errtrace
# Do not allow use of undefined vars. 
set -o nounset


# This create a DyamoDB table with a a primary key that is a string and no secondary key. 
# It sets a low provisioned throughput to qualify for free tier pricing and turns on a Stream
aws dynamodb create-table \
    --table-name ${FEEDBACK_TABLE_NAME} \
    --attribute-definitions AttributeName=${FEEDBACK_PRIMARY_KEY},AttributeType=S \
    --key-schema AttributeName=${FEEDBACK_PRIMARY_KEY},KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES

# After running this, the console should display the ARN of the table, which will contain the AWS Region and accountID
#
#  ...
# "LatestStreamArn": "arn:aws:dynamodb:region:accountID:table/Feedback/stream/timestamp
# ...

# Ensure your setup.rc AWS_DEFAULT_REGION equals the region displayed in the ARN
# Ensure your setup.rc AWS_ACCOUNT_ID equals the accountID displayed in the ARN
