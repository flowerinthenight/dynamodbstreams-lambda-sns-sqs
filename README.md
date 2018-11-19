## Overview
This is an example of tracking a dynamodb table's changes using [dynamodbstreams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html). Table item updates will trigger a lambda function that will then forward the event payload to an SNS topic. An example SQS subscriber is also provided that can subscribe to the SNS topic and print the event data.

A cloudformation [template](https://github.com/flowerinthenight/dynamodbstreams-lambda-sns-sqs/blob/master/template/development.yml) is provided to create all the resources needed in this example. The [example-consumer](https://github.com/flowerinthenight/dynamodbstreams-lambda-sns-sqs/tree/master/example-consumer) will create the SQS queue that subscribes to the SNS topic upon execution, if needed. 

## How to run
