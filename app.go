package main

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/aws/aws-sdk-go/aws"
    "encoding/json"
    "fmt"
    _ "github.com/joho/godotenv/autoload"
)

var queueUrl = "https://sqs.us-west-2.amazonaws.com/110303772622/ocs-order-accepted-driver-notifations"
var sqsClient *sqs.SQS

func init() {
    awsSession := session.Must(session.NewSession())
    sqsClient = sqs.New(awsSession)
}

func main() {
    println("Started.")
    receiveInput := &sqs.ReceiveMessageInput{
        QueueUrl: aws.String(queueUrl),
        MaxNumberOfMessages: aws.Int64(1),
    }

    for (true) {
        result, err := sqsClient.ReceiveMessage(receiveInput)
        if (err != nil) {
            println("ERROR", err.Error())
            continue
        }

        for _, msg := range result.Messages {
            handleMessage(msg)
        }

    }

    println("Goodbye.");
}

func handleMessage(msg *sqs.Message) error {
    sns, err := parseSNSMessage(msg.Body)
    if err != nil {
        return err
    }

    order, err := sns.parseOrder()
    if err != nil {
        return err
    }

    println(fmt.Sprintf("%v", order))

    deleteMessage(msg)

    return nil
}

func deleteMessage(msg *sqs.Message) error {
    sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
        QueueUrl: aws.String(queueUrl),
        ReceiptHandle: msg.ReceiptHandle,
    })

    return nil
}

type snsMessage struct {
    Type    string `json:"Type"`
    Message string `json:"Message"`
}

func parseSNSMessage(msgBody *string) (*snsMessage, error) {
    sns := &snsMessage{}
    err := json.Unmarshal([]byte(*msgBody), sns)
    if err != nil {
        return nil, err
    }

    return sns, nil
}

type order struct {
    ID                     string `json:"id"`
    Name                   string `json:"name"`
    Phone                  string `json:"string"`
    Address1               string `json:"address1"`
    Address2               string `json:"address2"`
    PostalCode             string `json:"postalCode"`
    AdditionalInstructions string `json:"additionalInstructions"`
}

func (m *snsMessage) parseOrder() (*order, error) {
    order := &order{}
    err := json.Unmarshal([]byte(m.Message), order)
    if err != nil {
        return nil, err
    }

    return order, nil
}