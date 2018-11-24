package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Joke JSON object
type Joke struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

// GetSSMValue - Get the encrypted value from SSM
func GetSSMValue(keyname string) (string, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return "", err

	}

	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(os.Getenv("AWS_DEFAULT_REGION")))
	withDecryption := true
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

	if err != nil {
		return "", err
	}

	return *param.Parameter.Value, nil
}

// GetJoke - Retrieve the feed from icanhazdadjoke.com
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
	if err := json.Unmarshal(body, &joke); err != nil {
		return joke, err
	}

	return joke, nil
}

// SendTweet - Login to twitter via API and send the tweet
func SendTweet(tweet string, twitterAccessKey string, twitterAccessSecret string, twitterConsumerKey string, twitterConsumerSecret string) error {
	config := oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	token := oauth1.NewToken(twitterAccessKey, twitterAccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	if _, _, err := client.Statuses.Update(tweet, nil); err != nil {
		return err
	}
	return nil
}

// HandleRequest - Handle the incoming Lambda request
func HandleRequest() (string, error) {
	// Get the access keys from SSM, we do this first as there's no point continuing if we can't get them.
	twitterAccessKey, err := GetSSMValue(os.Getenv("TWITTER_ACCESS_KEY"))
	if err != nil {
		return "", err
	}

	twitterAccessSecret, err := GetSSMValue(os.Getenv("TWITTER_ACCESS_SECRET"))
	if err != nil {
		return "", err
	}

	twitterConsumerKey, err := GetSSMValue(os.Getenv("TWITTER_CONSUMER_KEY"))
	if err != nil {
		return "", err
	}

	twitterConsumerSecret, err := GetSSMValue(os.Getenv("TWITTER_CONSUMER_SECRET"))
	if err != nil {
		return "", err
	}

	if twitterConsumerKey == "" {
		return "", errors.New("Twitter consumer key can not be null")
	}

	if twitterConsumerSecret == "" {
		return "", errors.New("Twitter consumer secret can not be null")
	}

	if twitterAccessKey == "" {
		return "", errors.New("Twitter access key can not be null")
	}

	if twitterAccessSecret == "" {
		return "", errors.New("Twitter access secret can not be null")
	}

	// Fetch the latest joke
	joke, err := GetJoke()
	if err != nil {
		// we have an error, oopsies, let's skip this round.
		return "", err
	}

	jokeTweet := fmt.Sprintf("%s #dadjokes", joke.Joke)

	// Send the joke to twitter
	if err := SendTweet(jokeTweet, twitterAccessKey, twitterAccessSecret, twitterConsumerKey, twitterConsumerSecret); err != nil {
		return "Failed to send", err
	}

	return "Sent!", nil
}

func main() {
	lambda.Start(HandleRequest)
}
