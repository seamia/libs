package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/seamia/libs/snapshot"
)

const (
	endpoint = "http://localhost:8000"
	region   = "us-east-1"
)

func main() {

	// init section
	sess, err := session.NewSession(
		&aws.Config{
			Region:   aws.String(region),
			Endpoint: aws.String(endpoint),
		},
	)

	if err != nil {
		fmt.Println("Got error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	fmt.Println("Talking to", endpoint)

	snap, err := snapshot.New(svc, "")
	if err != nil {
		fmt.Println("failed to get snapshot")
		return
	}
	saved, err := snapshot.Load("saved")
	if err != nil {
		fmt.Println("failed to load snapshot")
		return
	}

	snap.CompareWith(saved, nil)
}
