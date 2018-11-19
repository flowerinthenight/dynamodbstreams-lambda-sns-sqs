## Overview
This is an example of tracking a dynamodb table's changes using [dynamodbstreams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html). Table item updates will trigger a lambda function that will then forward the update payload to an SNS topic. An example SQS subscriber is also provided that can subscribe to the SNS topic and print the event data. 
