package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Joke struct {
	Id     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

// Retrieve the feed from icanhazdadjoke.com
func GetJoke() (Joke, error) {
	url := "https://icanhazdadjoke.com/"

	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Joke{}, err
	}

	req.Header.Set("Accept", "application/json")

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return Joke{}, getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return Joke{}, readErr
	}

	joke := Joke{}
	jsonErr := json.Unmarshal(body, &joke)
	if jsonErr != nil {

		return joke, jsonErr
	}

	return joke, nil
}

// Login to twitter via API and send the tweet
func SendTweet(tweet string) error {
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterAccessKey := os.Getenv("TWITTER_ACCESS_KEY")
	twitterAccessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	if twitterConsumerKey == "" {
		return errors.New("Twitter consumer key can not be null")
	}

	if twitterConsumerSecret == "" {
		return errors.New("Twitter consumer secret can not be null")
	}

	if twitterAccessKey == "" {
		return errors.New("Twitter access key can not be null")
	}

	if twitterAccessSecret == "" {
		return errors.New("Twitter access secret can not be null")
	}

	config := oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	token := oauth1.NewToken(twitterAccessKey, twitterAccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	_, _, err := client.Statuses.Update(tweet, nil)
	if err != nil {
		return err
	}
	return nil
}

func HandleRequest() (string, error) {
	joke, err := GetJoke()
	if err != nil {
		// we have an error, oopsies, let's skip this round.
		return "", err
	}

	sendError := SendTweet(joke.Joke)
	if sendError != nil {
		return "Failed to send", sendError
	} else {
		return "Sent!", nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
