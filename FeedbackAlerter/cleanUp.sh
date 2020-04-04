#!/usr/bin/env bash
# This file:
#
#  - Cleans up all the environmental variables and temporary files created as a result of setting up the Database Alerter
#
# Usage:
#
#   ./cleanUP.sh
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# remove files the scripts created
rm ${POLICY_DOCUMENT}
rm ${LAMBDA_JS_FILE}
rm ${LAMBDA_ZIP_FILE}
rm testOutput.txt


# unset all those environmental variables we sourced
#   (but don't unset AWS_DEFAULT_REGION... that one is normally there)
unset AWS_ACCOUNT_ID
unset FEEDBACK_TABLE_NAME
unset SUBSCRIBING_EMAIL
unset FEEDBACK_PRIMARY_KEY
unset POLICY_NAME
unset TRUST_RELATIONSHIP_DOCUMENT
unset POLICY_TEMPLATE
unset POLICY_DOCUMENT
unset LAMBDA_ROLE_NAME
unset LAMBDA_FUNCTION_NAME
unset TOPIC_NAME
unset LAMBDA_TEMPATE
unset LAMBDA_JS_FILE
unset LAMBDA_ZIP_FILE
unset TEST_PAYLOAD_FILE