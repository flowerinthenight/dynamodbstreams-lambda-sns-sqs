package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	awszcnf "github.com/NYTimes/gizmo/config/aws"
	awszpubsub "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

// getSqsAllowAllPolicy returns a policy that can be used when creating an SQS queue that allow
// all SQS actions for everybody.
func getSqsAllowAllPolicy(region, acct, queueName string) string {
	return `{
  "Version":"2008-10-17",
  "Id":"idteststreams01",
  "Statement":[
    {
	  "Sid":"sidteststreams01",
	  "Effect":"Allow",
	  "Principal":"*",
	  "Action":"SQS:*",
	  "Resource":"` + fmt.Sprintf("arn:aws:sqs:%s:%s:%s", region, acct, queueName) + `"
    }
  ]
}`
}

func runSink(quit, done chan error) {
	acct := os.Getenv("AWS_ACCT_ID")
	region := os.Getenv("AWS_REGION")
	topicArn := fmt.Sprintf("arn:aws:sns:%s:%s:teststreams-dbstreams-snstopic", region, acct)
	queueName := "teststreams-dbstreams-snstopic-subscription-teststreams"
	policy := getSqsAllowAllPolicy(region, acct, queueName)

	log.Printf("acct=%v, region=%v, topic=%v, queue=%v", acct, region, topicArn, queueName)

	sess := session.New(&aws.Config{
		Region: aws.String(region),
	})

	svc := sqs.New(sess)

	var qUrl *sqs.GetQueueUrlOutput
	var err error

	// Get queue to check if exists.
	qUrl, err = svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})

	if err != nil {
		log.Printf("GetQueueUrl failed, err=%v, attempt to create queue", err)

		// Attempt to create the queue.
		_, err = svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(queueName),
			Attributes: map[string]*string{
				"Policy": aws.String(policy),
			},
		})

		if err != nil {
			log.Fatalf("CreateQueue failed, err=%v", err)
		}

		qUrl, err = svc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(queueName),
		})
	}

	// We only need the arn bit actually.
	qAttr, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: qUrl.QueueUrl,
		AttributeNames: []*string{
			aws.String("All"),
		},
	})

	if err != nil {
		log.Fatalf("GetQueueAttributes failed, err=%v", err)
	}

	log.Printf("attr=%v", qAttr)

	snssvc := sns.New(sess)
	_, err = snssvc.Subscribe(&sns.SubscribeInput{
		TopicArn: aws.String(topicArn),
		Protocol: aws.String("sqs"),
		Endpoint: qAttr.Attributes["QueueArn"],
	})

	if err != nil {
		log.Fatalf("Subscribe failed, err=%v", err)
	}

	sub, err := awszpubsub.NewSubscriber(awszpubsub.SQSConfig{
		Config: awszcnf.Config{
			Region: region,
		},
		QueueName:           queueName,
		QueueOwnerAccountID: acct,
		ConsumeBase64:       aws.Bool(false),
	})

	if err != nil {
		log.Fatalf("NewSubscriber failed, err=%v", err)
	}

	pipe := sub.Start()
	defer sub.Stop()

	for {
		select {
		case m := <-pipe:
			if m != nil {
				var w sync.WaitGroup
				w.Add(1)

				go func() {
					defer func() {
						m.Done() // so we can always remove msg from queue
						w.Done()
					}()

					defer func(begin time.Time) {
						log.Printf("duration=%v", time.Since(begin))
					}(time.Now())

					// Define only the ones we are interested in.
					// Ref: https://docs.aws.amazon.com/sns/latest/dg/sns-sqs-as-subscriber.html
					type _event struct {
						MessageId string `json:"MessageId"`
						Message   string `json:"Message"`
						Timestamp string `json:"Timestamp"`
					}

					var e _event
					var rec events.DynamoDBEventRecord
					err = json.Unmarshal(m.Message(), &e)
					if err != nil {
						log.Printf("unmarshal wrapper failed, err=%v", err)
						return
					}

					log.Printf("event=%+v", e)

					err = json.Unmarshal([]byte(e.Message), &rec)
					if err != nil {
						log.Printf("unmarshal payload failed, err=%v", err)
						return
					}

					log.Printf("rec=%v", rec)
				}()

				w.Wait()
			}
		case <-quit:
			log.Printf("%v sink requested to terminate", queueName)
			done <- nil
			return
		}
	}
}

func main() {
	log.Printf("start testconsumer on %v", time.Now())

	quit := make(chan error)
	done := make(chan error)

	go runSink(quit, done)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		quit <- errors.Errorf("%s", <-c)
	}()

	err := <-done
	if err != nil {
		log.Fatalf("run sink failed: %v", err)
	}
}
