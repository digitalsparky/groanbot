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
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Version
var Version string

// Environment
var Env string

// Joke JSON object
type Joke struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

// Twitter Access
type Twitter struct {
	config      *oauth1.Config
	token       *oauth1.Token
	httpClient  *http.Client
	client      *twitter.Client
	tweetFormat string
	screenName  string
}

func (t *Twitter) Setup() {
	log.Debug("Setting up twitter client")
	var twitterAccessKey string
	var twitterAccessSecret string
	var twitterConsumerKey string
	var twitterConsumerSecret string

	if Env == "production" {
		// Get the access keys from SSM
		twitterAccessKey = GetSSMValue(os.Getenv("TWITTER_ACCESS_KEY"))
		twitterAccessSecret = GetSSMValue(os.Getenv("TWITTER_ACCESS_SECRET"))
		twitterConsumerKey = GetSSMValue(os.Getenv("TWITTER_CONSUMER_KEY"))
		twitterConsumerSecret = GetSSMValue(os.Getenv("TWITTER_CONSUMER_SECRET"))
	} else {
		// Get the access keys from ENV
		twitterAccessKey = os.Getenv("TWITTER_ACCESS_KEY")
		twitterAccessSecret = os.Getenv("TWITTER_ACCESS_SECRET")
		twitterConsumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
		twitterConsumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	}

	twitterScreenName := os.Getenv("TWITTER_SCREEN_NAME")

	if twitterScreenName == "" {
		log.Fatalf("Twitter screen name cannot be null")
	}

	if twitterConsumerKey == "" {
		log.Fatal("Twitter consumer key can not be null")
	}

	if twitterConsumerSecret == "" {
		log.Fatal("Twitter consumer secret can not be null")
	}

	if twitterAccessKey == "" {
		log.Fatal("Twitter access key can not be null")
	}

	if twitterAccessSecret == "" {
		log.Fatal("Twitter access secret can not be null")
	}

	log.Debug("Setting up oAuth for twitter")
	// Setup the new oauth client
	t.config = oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	t.token = oauth1.NewToken(twitterAccessKey, twitterAccessSecret)
	t.httpClient = t.config.Client(oauth1.NoContext, t.token)

	// Twitter client
	t.client = twitter.NewClient(t.httpClient)

	// Set the screen name for later use
	t.screenName = twitterScreenName

	// This is the format of the tweet, ie: Mature puns are fully groan #pun #dadjoke
	t.tweetFormat = "%s #pun / #dadjoke"
	log.Debug("Twitter client setup complete")
}

func (t *Twitter) GetTweetString(tweet string) string {
	return fmt.Sprintf(t.tweetFormat, tweet)
}

// Send the tweet to twitter
func (t *Twitter) Send(tweet string) {
	log.Debug("Sending tweet")
	if Env != "production" {
		log.Infof("Non production mode, would've tweeted: %s", tweet)
	}
	if Env == "production" {
		if _, _, err := t.client.Statuses.Update(t.GetTweetString(tweet), nil); err != nil {
			log.Fatalf("Error sending tweet to twitter: %s", err)
		}
	}
}

// CheckLast30
// We want to make sure that we've not tweeted the joke in the last 120 tweets
// So we get the currently list of tweets and make sure it's not in there
// Before sending the tweet
func (t *Twitter) CheckLast30(checkTweet string) bool {
	log.Debug("Checking to see if the tweet appeared in the last 120 tweets")

	tweets, _, err := t.client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: t.screenName,
		Count:      120,
		TweetMode:  "extended",
	})

	if err != nil {
		log.Fatalf("Error getting last 30 tweets from user: %s", err)
	}

	for _, tweet := range tweets {
		if t.GetTweetString(checkTweet) == tweet.Text {
			return true
		}
	}

	return false
}

// Twitter API constant
var tw Twitter

// GetSSMValue - Get the encrypted value from SSM
func GetSSMValue(keyname string) string {
	log.Debugf("Getting SSM Value for %s", keyname)

	// Setup the SSM Session
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatalf("Error occurred retrieving SSM session: %s", err)
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
		log.Fatalf("Error occurred retrieving SSM parameter: %s", err)
		os.Exit(1)
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
		log.Fatalf("Error occurred creating new request: %s", err)
	}

	// Set the user agent to Groanbot <verion> - twitter.com/groanbot
	req.Header.Set("User-Agent", fmt.Sprintf("GroanBot %s - twitter.com/groanbot", Version))

	// Tell the remote server to send us JSON
	req.Header.Set("Accept", "application/json")

	invalidJoke := true
	try := 0
	maxTries := 10

	var joke Joke

	for invalidJoke {
		// We're only going to try maxTries times, otherwise we'll fatal out.
		if try >= maxTries {
			log.Fatal("Exiting after attempts to retrieve joke failed.")
			os.Exit(1)
		}

		// Execute the request
		log.Debugf("Attempting request to %s", req)
		res, getErr := httpClient.Do(req)
		if getErr != nil {
			// We got an error, lets bail out, we can't do anything more
			log.Errorf("Error occurred retrieving joke from API: %s", getErr)
			try += 1
			continue
		}

		// BGet the body from the result
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			// This shouldn't happen, but if it does, error out.
			log.Errorf("Error occurred reading from result body: %s", readErr)
			try += 1
			continue
		}

		if err := json.Unmarshal(body, &joke); err != nil {
			// Invalid JSON was received, bail out
			log.Errorf("Error occurred decoding joke: %s", err)
			try += 1
			continue
		}

		// Make sure it's not 0 characters
		if len(joke.Joke) == 0 {
			try += 1
			continue
		}

		// check to make sure the tweet hasn't been sent before
		if tw.CheckLast30(joke.Joke) {
			try += 1
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
	log.Debug("Started handling request")
	tw.Setup()
	joke := GetJoke()
	tw.Send(joke.Joke)
}

// Set the local environment
func setRunningEnvironment() {
	// Get the environment variable
	switch os.Getenv("APP_ENV") {
	case "production":
		Env = "production"
	case "development":
		Env = "development"
	case "testing":
		Env = "testing"
	default:
		Env = "development"
	}

	if Env != "production" {
		Version = Env
	}
}

func shutdown() {
	log.Info("Shutdown request registered")
}

func init() {
	// Set the environment
	setRunningEnvironment()

	// Set logging configuration
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log.SetReportCaller(true)
	switch Env {
	case "development":
		log.SetLevel(log.DebugLevel)
	case "production":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	// Start the bot
	log.Debug("Starting main")
	log.Printf("GroanBot %s", Version)
	if Env == "production" {
		lambda.Start(HandleRequest)
	} else {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
		HandleRequest()
	}
}
