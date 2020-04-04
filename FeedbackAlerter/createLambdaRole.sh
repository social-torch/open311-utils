#!/usr/bin/env bash
# This file:
#
#  - Creates a AWS Identity and Access Management (IAM) role and assigns permissions to it.
#
# Usage:
#
#   ./createLambdaRole.sh
#
# Based on a Amazon Tutorial: Process New Items with DynamoDB Streams and Lambda
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html
#

# Exit on error. 
set -o errexit
# Do not allow use of undefined vars. 
set -o nounset

# Update the policy template with your Region and Account ID
sed \
    -e "s|#{AWS_DEFAULT_REGION}|${AWS_DEFAULT_REGION}|g"     \
    -e "s|#{AWS_ACCOUNT_ID}|${AWS_ACCOUNT_ID}|g"             \
    -e "s|#{FEEDBACK_TABLE_NAME}|${FEEDBACK_TABLE_NAME}|g"   \
    -e "s|#{LAMBDA_FUNCTION_NAME}|${LAMBDA_FUNCTION_NAME}|g" \
    ${POLICY_TEMPLATE} > ${POLICY_DOCUMENT}

# Create the IAM Role
 aws iam create-role --role-name ${LAMBDA_ROLE_NAME} \
     --path "/service-role/" \
     --assume-role-policy-document file://${TRUST_RELATIONSHIP_DOCUMENT}

# Attach the IAM policy to the new role
 aws iam put-role-policy --role-name ${LAMBDA_ROLE_NAME} \
     --policy-name ${POLICY_NAME} \
     --policy-document file://${POLICY_DOCUMENT}

