# Setup AWS to email subscribers upon new DynamoDB table entry

More specifically, when a Social Torch user clicks the "Report Innappropriate" button or submits feedback, alert the founders or other subscribers to take action.  

The general pattern is:

1. New item added to DynamoDB Feedback Table via mobile app
2. DynamoDB Streams grabs data and sends to lambda
3. Lambda formats the message
4. Amazon SNS distributes notifification to anyone subscribed to topic (SMS, email, push, whatev)

This repo is based on a [tutorial from Amazon](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.Tutorial.html) and documents a more specific case of how to set it up for the Social Torch configuration

## Prerequisites

* These scripts assume [AWS CLI](https://aws.amazon.com/cli/) is installed and you have an AWS Account.  
* These scripts assume you have [jq](https://stedolan.github.io/jq/) installed
* These scripts assume you have a 'zip' utility in your path

## Step 0: AWS CLI Setup

Make darn sure your AWS CLI is setup to configure the thing you want to configure.  

* See documentation [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)  
* Pro tip: try using [profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html)  

## Step 1: Set environment variables

Edit the `setup.rc` file and make sure the AWS Region and AccountID match the destination you want to use.  
Make sure the `SUBSCRIBING_EMAIL` field is yours.  
The rest should be fine as is.  

Run

```bash
source setup.rc
```

to load those vars globally, baby.  There is probably a better way to share info among the various scripts. Feedback requested.  Don't worry, we'll clean up later.

## Step 2: Create DynamoDB Table (Optional)

If you don't already have a DynamoDB table for feedback, use this shell script to create one.

```bash
./createTable.sh
```

If your table alread exists, please ensure Streams are turned on for this table. (In the console, go to the DynamoDB tables list, click the Feedback table, click overview, and look for the Streams heading.  If streams aren't enabled, click "Manage Stream")

## Step 3: Create Lambda Execution Role

AWS seems to be rather picky that one sets up proper security for what can talk to what.  Let's create an IAM policy and attach it to an IAM role we create.

```bash
./createLambdaRole.sh
```

Note: The policy has 4 statements that allow `${role_name}` to do the following:

1. Execute a Lambda function `${LAMBDA_FUNCTION_NAME}`.
2. Access Amazon CloudWatch Logs. The Lambda function writes diagnostics to CloudWatch Logs at runtime.
3. Read data from the DynamoDB stream for `${FEEDBACK_TABLE_NAME}`.
4. Publish messages to Amazon SNS.

## Step 4: Create AWS Simple Notification Service (SNS) Topic

Create an SNS topic and subscribe to it with an email address.
Edit the subscribing_email variable in the `createSNSTopic.sh` file to set it to the email you want to get notifications.

```bash
./createSNSTopic.sh
```

note: Amazon SNS sends a confirmation message to your email address. Choose the Confirm subscription link in that message to complete the subscription process.

## Step 5: Create Lambda Function

By using the lambda function template and modifying for your endpoint, this script, creates a node.js lambda, zips it up, and uploads to AWS.

```bash
./createLambdaFunction.sh
```

## Step 6: Test Lambda Function

While it isn't really the style of Social Torch to test things, I though I would try something new.

```bash
./testLambda.sh
```

If the test was successful,  you should see the following output.

```bash
{
    "StatusCode": 200
    "ExecutedVersion": '$LATEST'
}
```

Furthermore, the newly created testOutput.txt file should contain:

```bash
"Successfully processed 1 records."
```

... AND ....  
You should recieve an email message.  Go check... might take 2 or 3 minutes.

## Step 7: Create Trigger

Nice job getting all the infrastructure and policies set up.  Now, let's setup the trigger by associating the lambda function with the Feedback Table stream (event source)

```bash
./createTrigger.sh
```

## Step 8: Test the Trigger

I'm done typing for you.  No shell script for this one.  Type the following to put an item in the feedback table, which should trigger your trigger. (Assumes tablename is "Feedback")

```bash
aws dynamodb put-item \
    --table-name Feedback \
    --item id={S="Testing...1...2...3"}
```

Now... go check your email.

## Step 9: Clean up

Because I couldn't think of an easier way at the time, I created a bunch of environment variables to make this set of scripts easier for you.  Time to get rid of those.  

Also, these scripts created some temporary files.  Let's remove those as well.

```bash
./cleanUp.sh
```

## Step 10: Do a Happy Dance

Exercise left to reader.
