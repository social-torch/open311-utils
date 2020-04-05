#!/usr/bin/env bash
# This file:
#
#  - Creates an AWS Lambda function code based on a template modified to your account
#  - zips it up and uplads to AWS
#
# Usage:
#
#   ./createLambdaFunction.sh
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Exit on error. 
set -o errexit
# Do not allow use of undefined vars. 
set -o nounset

# Create a file named publishFeedback.js based on a template. 
# Replace region and accountID with your AWS Region and account ID.
sed \
    -e "s|#{AWS_DEFAULT_REGION}|${AWS_DEFAULT_REGION}|g"     \
    -e "s|#{AWS_ACCOUNT_ID}|${AWS_ACCOUNT_ID}|g"             \
    -e "s|#{TOPIC_NAME}|${TOPIC_NAME}|g"                     \
    ${LAMBDA_TEMPATE} > ${LAMBDA_JS_FILE}

# Zip up the lambda code
zip ${LAMBDA_ZIP_FILE} ${LAMBDA_JS_FILE}

# Lookup Role ARN (depends on jq being installed) 
ROLE_ARN=$(aws iam get-role --role-name ${LAMBDA_ROLE_NAME} | jq -r '.Role.Arn')

# Create the lambda function
aws lambda create-function \
    --region ${AWS_DEFAULT_REGION} \
    --function-name ${LAMBDA_FUNCTION_NAME} \
    --zip-file fileb://${LAMBDA_ZIP_FILE} \
    --role ${ROLE_ARN} \
    --handler ${LAMBDA_FUNCTION_NAME}.handler \
    --timeout 5 \
    --runtime nodejs10.x
