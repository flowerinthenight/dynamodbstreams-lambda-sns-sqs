package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func handler(ctx context.Context, e events.DynamoDBEvent) {
	region, topicArn := os.Getenv("AWS_REGION"), os.Getenv("TOPIC_ARN")
	log.Printf("region=%v, topicArn=%v", region, topicArn)

	sess := session.New(&aws.Config{
		Region: aws.String(region),
	})

	svc := sns.New(sess)
	for _, record := range e.Records {
		log.Printf("processing request data, event=%s, type=%s", record.EventID, record.EventName)

		// Marshal to []byte then string.
		b, err := json.Marshal(record)
		if err != nil {
			log.Printf("marshal failed, err=%v", err)
			continue
		}

		_, err = svc.Publish(&sns.PublishInput{
			TopicArn: aws.String(topicArn),
			Message:  aws.String(string(b)),
		})

		if err != nil {
			log.Printf("publish failed, err=%v", err)
			continue
		}
	}
}

func main() {
	lambda.Start(handler)
}
