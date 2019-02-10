/**
Pieces of below started from the examples published at:
	- https://github.com/gsuitedevs/go-samples/blob/master/gmail/quickstart/quickstart.go
	- https://developers.google.com/gmail/api/quickstart/go
*/

// Package messages handles logic to talk to gmail API
package messages

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/mvdan/xurls"
)

// GetClient will retrieve a token, saves the token, then returns the generated client.
func GetClient() *http.Client {
	b, err := ioutil.ReadFile("./credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "./token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// will request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	// nolint: gosec
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalln(err)
	}
}

// GetMessages will read a page of messages, returning a token to iterate on for next page
func GetMessages(req *gmail.UsersMessagesListCall, nextToken ...string) (*gmail.ListMessagesResponse, string) {
	ntoken := ""
	if len(nextToken) > 0 {
		ntoken = nextToken[0]
	}
	r, err := req.PageToken(ntoken).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}
	fmt.Println("nextToken = ", nextToken, "   r.NextPageToken = ", r.NextPageToken)
	return r, r.NextPageToken
}

func getRFC282Headers(headers []*gmail.MessagePartHeader) (from string, subject string, err error) {
	for _, h := range headers {
		switch h.Name {
		case "Subject":
			subject = h.Value
		case "From":
			from = h.Value
		}
	}
	return
}

// GetFrom will return the sender of this email
func GetFrom(headers []*gmail.MessagePartHeader) (from string) {
	from, _, err := getRFC282Headers(headers)
	if err != nil {
		log.Fatalf("Unable to parse headers: %v", err)
	}
	return
}

// GetSubject will return the subject of a message
func GetSubject(headers []*gmail.MessagePartHeader) (subject string) {
	_, subject, err := getRFC282Headers(headers)
	if err != nil {
		log.Fatalf("Unable to parse headers: %v", err)
	}
	return
}

// GetMessageContent will read the payload of a single or multi-part message
func GetMessageContent(payload *gmail.MessagePart) (content string) {
	if len(payload.Parts) > 0 {
		for _, part := range payload.Parts {
			if part.MimeType == "text/html" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					log.Fatalf("Unable to decode message: %v", err)
				}
				content = string(data)
			}
		}
	} else {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			log.Fatalf("Unable to decode message: %v", err)
		}
		content = string(data)
	}
	return
}

// GetAllURLs will return the urls present in a string
func GetAllURLs(message string) []string {
	urls := make([]string, 0, 10)
	allUrls := xurls.Relaxed().FindAllString(message, -1)
	for i := range allUrls {
		if strings.HasPrefix(allUrls[i], "http") {
			urls = append(urls, allUrls[i])
		}
	}
	return urls
}
