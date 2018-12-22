package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
)

// Version
var Version string

// Joke JSON object
type Joke struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

// Twitter Access
type Twitter struct {
	AccessKey      string
	AccessSecret   string
	ConsumerKey    string
	ConsumerSecret string
}

func (t *Twitter) Init() {
	// Setup the new oauth client
	t.Config = oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	t.Token = oauth1.NewToken(twitterAccessKey, twitterAccessSecret)
	t.HttpClient = config.Client(oauth1.NoContext, token)

	// Twitter client
	t.Client = twitter.NewClient(httpClient)

	// Get the access keys from SSM, we do this first as there's no point continuing if we can't get them.
	t.AccessKey = GetSSMValue(os.Getenv("TWITTER_ACCESS_KEY"))
	t.AccessSecret = GetSSMValue(os.Getenv("TWITTER_ACCESS_SECRET"))
	t.ConsumerKey = GetSSMValue(os.Getenv("TWITTER_CONSUMER_KEY"))
	t.ConsumerSecret = GetSSMValue(os.Getenv("TWITTER_CONSUMER_SECRET"))

	if t.twitterConsumerKey == "" {
		log.Fatal("Twitter consumer key can not be null")
	}

	if t.twitterConsumerSecret == "" {
		log.Fatal("Twitter consumer secret can not be null")
	}

	if t.twitterAccessKey == "" {
		log.Fatal("Twitter access key can not be null")
	}

	if t.twitterAccessSecret == "" {
		log.Fatal("Twitter access secret can not be null")
	}

	// This is the format of the tweet, ie: Mature puns are fully groan #pun #dadjoke
	t.TweetFormat = "%s #pun #dadjoke"
}

// Send the tweet to twitter
func (t *Twitter) Send(tweet string) {
	if _, _, err := t.client.Statuses.Update(tweet, nil); err != nil {
		log.Fatalf("Error sending tweet to twitter: %s\n", err)
	}
}

// Check30Days
// We want to make sure that we've not tweeted the joke in the last 30 days
// So we get the currently list of tweets and make sure it's not in there
// Before sending the tweet
func (t *Twitter) Check30Days(tweet string) bool {
	// Set the format
	tweetMessage := fmt.Sprintf(t.TweetFormat, tweet)

	// Get the latest 30 days of tweets from twitter

	// Check to make sure tweetMessage isn't in the last 30 days

	// If it's in the last 30 days, return true, otherwise false

	return false
}

// GetSSMValue - Get the encrypted value from SSM
func GetSSMValue(keyname string) string {

	// Setup the SSM Session
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatalf("Error occurred retrieving SSM session: %s\n", err)
	}

	// Create a new SSM service using the SSM session with the specific region
	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(os.Getenv("AWS_DEFAULT_REGION")))
	// Enable Server side decryption
	withDecryption := true
	// Get the parameter from SSM
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

	// If we get an error, fatal out with the error message
	if err != nil {
		log.Fatalf("Error occurred retrieving SSM parameter: %s\n", err)
	}

	// Return the dereferenced value
	return *param.Parameter.Value
}

// GetJoke - Retrieve the feed from icanhazdadjoke.com
func GetJoke() Joke {
	url := "https://icanhazdadjoke.com/"

	// Setup a new HTTP Client with 2 second timeout
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	// Create a new HTTP Request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// An error has occurred that we can't recover from, bail.
		log.Fatalf("Error occurred creating new request: %s\n", err)
	}

	// Set the user agent to Groanbot <verion> - twitter.com/groanbot
	req.Header.Set("User-Agent", fmt.Sprintf("GroanBot %s - twitter.com/groanbot", Version))

	// Tell the remote server to send us JSON
	req.Header.Set("Accept", "application/json")

	invalidJoke := true
	try := 0
	maxTries := 3

	var joke Joke

	for invalidJoke {
		// Execute the request
		res, getErr := httpClient.Do(req)
		if getErr != nil {
			// We got an error, lets bail out, we can't do anything more
			log.Fatalf("Error occurred retrieving joke from API: %s\n", getErr)
		}

		// BGet the body from the result
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			// This shouldn't happen, but if it does, error out.
			log.Fatalf("Error occurred reading from result body: %s\n", readErr)
		}

		if err := json.Unmarshal(body, &joke); err != nil {
			// Invalid JSON was received, bail out
			log.Fatalf("Error occurred decoding joke: %s\n", err)
		}

		// We're only going to try maxTries times, otherwise we'll fatal out.
		if try >= maxTries {
			log.Fatal("Exiting after attempts to retrieve joke failed.")
		}

		// Make sure it's not 0 characters
		if len(joke.Joke) == 0 {
			try += 1
			continue
		}

		// check to make sure the tweet hasn't been sent before
		if t.Check30Days(joke.Joke) {
			continue
		}

		// If we get here we've found a tweet, exit the loop
		invalidJoke = false
	}

	// Return the valid joke response
	return joke
}


// HandleRequest - Handle the incoming Lambda request
func HandleRequest() {
	twitter := new Twitter
	twitter.Setup()
	joke := GetJoke()
	twitter.Send(joke.Joke)

}

func init() {
	// Set logging configuration
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func main() {
	// Start the bot
	log.Printf("GroanBot %s\n", Version)
	lambda.Start(HandleRequest)
}
