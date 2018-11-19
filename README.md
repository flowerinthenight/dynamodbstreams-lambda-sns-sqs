## Overview
This is an example of tracking a dynamodb table's changes using [dynamodbstreams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html). Table item updates will trigger a lambda function that will then forward the event payload to an SNS topic. An example SQS subscriber is also provided that will subscribe to the SNS topic and print all event data.

A cloudformation [template](https://github.com/flowerinthenight/dynamodbstreams-lambda-sns-sqs/blob/master/template/development.yml) is provided to create all the resources needed in this example. The [example-consumer](https://github.com/flowerinthenight/dynamodbstreams-lambda-sns-sqs/tree/master/example-consumer) will create the SQS queue that subscribes to the SNS topic upon execution, if needed. The lambda code is in [main.go](https://github.com/flowerinthenight/dynamodbstreams-lambda-sns-sqs/blob/master/main.go).

## How to run
You need to have the following required environment variables:
```bash
AWS_ACCT_ID={your-aws-account-id}
AWS_REGION={your-aws-region}
AWS_ACCESS_KEY_ID={key}
AWS_SECRET_ACCESS_KEY={secret}
```

This example assumes that you will use an existing table. In my case, I used a test table named TESTSTREAM with dynamodbstreams enabled. View type is "New and old images". The streams' ARN is in the template (`EventSourceArn`) so you might want to update that part.

Then run the following command from the repo's root folder:
```bash
# Note that the GO111MODULE=on variable is enabled during the build.
$ make deploy
```

Your stack should be ready at this point, provided that the AWS credentials has the permission to create the resources. You can now run the example consumer for testing.
```bash
$ cd example-consumer/
$ go build -v
$ ./example-consumer
```

While the consumer is running, you can update your test table by inserting new items, modifying or removing existing items. The running consumer should print all event data.

Finally, to cleanup the stack, run the following command from the repo's root folder:
```bash
# This command will delete the cloudformation stack. It will not delete the SQS queue
# created by example-consumer so you need to delete that manually.
$ make remove
```
