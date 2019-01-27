package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"google.golang.org/api/gmail/v1"

	"github.com/jwmatthews/createmymeals/pkg/messages"
)

var flagSerial bool
var flagStore bool

func init() {
	flag.BoolVar(&flagSerial, "serial", false, "Run in serial mode if set")
	flag.BoolVar(&flagStore, "store", false, "Store messages in local database")
}

func fetchMessage(svc *gmail.Service, m *gmail.Message) *gmail.Message {
	fmt.Println("Fetching message id", m.Id)
	msg, _ := svc.Users.Messages.Get("me", m.Id).Format("full").Do()
	return msg
}

func displayMessages(c chan *gmail.Message) {
	for msg := range c {
		subject := messages.GetSubject(msg.Payload.Headers)
		content := messages.GetMessageContent(msg.Payload)
		urls := messages.GetAllURLs(content)
		urlOutput := ""
		if len(urls) > 0 {
			urlOutput = strings.Join(urls, ", ")
		} else {
			urlOutput = "N/A"
		}
		fmt.Println(msg.Id, subject, "\n\t", urlOutput)
	}
}

func processMessages(svc *gmail.Service, req *gmail.UsersMessagesListCall, concurrent bool, store bool) {

	var producerGroup sync.WaitGroup
	var consumerGroup sync.WaitGroup
	var messgChannel = make(chan *gmail.Message, 0)

	//
	// Creating the handling goroutine first so that the serial processing case
	// is not blocked sending messages to the channel
	//
	consumerGroup.Add(1)
	go func() {
		defer consumerGroup.Done()
		displayMessages(messgChannel)
	}()

	//
	// Iterate over API call to retrieve list of messages
	// nextToken is used on each request to fetch next batch of messages
	//
	nextToken := ""
	for ok := true; ok; ok = nextToken != "" {
		var r *gmail.ListMessagesResponse
		r, nextToken = messages.GetMessages(req, nextToken)
		fmt.Println(len(r.Messages), " messages found")

		//
		// Next we need to make a separate API call per message to get it's content
		//
		for _, m := range r.Messages {
			if concurrent {
				producerGroup.Add(1)
				go func(m *gmail.Message, c chan<- *gmail.Message) {
					defer producerGroup.Done()
					// Note that for concurrent case it's important we pass in a copy of 'm'
					// opposed to using the variable from for loop directly.
					msg := fetchMessage(svc, m)
					c <- msg
				}(m, messgChannel)
			} else {
				msg := fetchMessage(svc, m)
				messgChannel <- msg
			}
		}
	}
	//
	// We will wait for the producers to complete
	// Note: that the goroutine for consuming the messages is intentionally not part of this group
	//
	producerGroup.Wait()
	close(messgChannel)
	//
	// Now we wait for consumer to complete
	//
	consumerGroup.Wait()
}

func main() {
	flag.Parse()
	concurrent := !flagSerial
	client := messages.GetClient()
	svc, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	req := svc.Users.Messages.List("me").Q("label:Recipes").
		MaxResults(25)

	processMessages(svc, req, concurrent, flagStore)
}
