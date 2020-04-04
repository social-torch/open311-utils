#!/usr/bin/env bash
# This file:
#
#  - uses a dummy json input to test the lambda deployed as part of this set of scripts
#
# Usage:
#
#   ./testLambda.sh
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Do not allow use of undefined vars. 
set -o nounset

# Call the lambda function with a dummy input.  Write output to testOutput.txt file
aws lambda invoke  --function-name ${LAMBDA_FUNCTION_NAME} --payload file://${TEST_PAYLOAD_FILE} testOutput.txt
