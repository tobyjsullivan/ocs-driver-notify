package main

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/sfreiberg/gotwilio"
    "encoding/json"
    "fmt"
    _ "github.com/joho/godotenv/autoload"
    "os"
    "errors"
    "net/url"
)

var queueUrl = "https://sqs.us-west-2.amazonaws.com/110303772622/ocs-order-accepted-driver-notifations"
var sqsClient *sqs.SQS
var twilioClient *gotwilio.Twilio
var servicePhone string
var driverPhone string

func init() {
    awsSession := session.Must(session.NewSession())
    sqsClient = sqs.New(awsSession)

    accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")
    twilioClient = gotwilio.NewTwilioClient(accountSid, authToken)

    servicePhone = os.Getenv("SERVICE_PHONE")
    driverPhone = os.Getenv("DRIVER_PHONE_NUMBER")
}

func main() {
    println("Started.")
    receiveInput := &sqs.ReceiveMessageInput{
        QueueUrl: aws.String(queueUrl),
        MaxNumberOfMessages: aws.Int64(10),
    }

    for (true) {
        result, err := sqsClient.ReceiveMessage(receiveInput)
        if (err != nil) {
            println("ERROR", err.Error())
            continue
        }

        for _, msg := range result.Messages {
            go handleMessage(msg)
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

    err = sendDriverSMS(order)
    if err != nil {
        return err
    }

    println(fmt.Sprintf("SMS sent to driver for Order %s", order.ID))

    deleteMessage(msg)

    return nil
}

func sendDriverSMS(order *order) error {
    content := fmt.Sprintf("New Order!\n%s (%s)\n%s\n%s\n%s\n\n%s\n\n%s", order.Name, order.Phone, order.Address1,
        order.Address2, order.PostalCode, order.AdditionalInstructions, wazeUrl(order))

    resp, exception, err := twilioClient.SendSMS(servicePhone, driverPhone, content, "", "")
    if err != nil {
        return err
    }

    if exception != nil {
        return errors.New(fmt.Sprintf("Twilio Exception: %s (Code: %d)", exception.Message, exception.Code))
    }

    println(fmt.Sprintf("Sent Twilio message %s", resp.Sid))

    return nil
}

func wazeUrl(order *order) string {
    wazeQuery := &url.Values{}
    wazeQuery.Set("q", order.Address1)
    wazeUrl := &url.URL{
        Scheme: "waze",
        ForceQuery: true,
        RawQuery: wazeQuery.Encode(),
    }

    return wazeUrl.String()
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
    Phone                  string `json:"phone"`
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