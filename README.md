# GroanBot

Groanbot is a twitter bot that tweets daily dad jokes from icanhazdadjoke.com.

This was built live on stream at twitch.tv/digitalsparky for the #noservernovember challenge.

Check out my twitter at twitter.com/digitalsparky

This project uses Golang with GO111MODULES enabled.

You will need the following environment variables to deploy:

```
AWS_ACCOUNT_ID="<your account id>" - [REQUIRED] - This defaults to "1111222233334444" and must be changed
AWS_DEFAULT_REGION="<your deploy region>" - This defaults to "us-east-2"
AWS_PROFILE="<your profile name>" - This defaults to 'default', this is the AWS CLI Profile name.
GROANBOT_BUILD_STAGE="<your build stage>" - This defaults to "prod"
```

Tweet .env file needs the following variables (tweet/.env)

```
TWITTER_SCREEN_NAME=GroanBot
TWITTER_ACCESS_KEY=XXX
TWITTER_ACCESS_SECRET=XXX
TWITTER_CONSUMER_KEY=XXX
TWITTER_CONSUMER_SECRET=XXX
MAX_RETRIES=10
SKIP_PREVIOUS_TWEETS=120
```

Build the binary by running

```make```

Deploy using:

```make deploy```


# Like my stuff?

Would you like to buy me a coffee or send me a tip?
While it's not expected, I would really appreciate it.

[![Paypal](https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png)](https://paypal.me/MattSpurrier) <a href="https://www.buymeacoffee.com/digitalsparky" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/white_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
