# groanbot

Groanbot is a twitter bot that tweets daily dad jokes from icanhazdadjoke.com.

This was built live on stream at twitch.tv/digitalsparky for the #noservernovember challenge.

Check out my twitter at twitter.com/digitalsparky

This project uses Golang with GO111MODULES enabled.

You will need the following environment variables to build:

```
AWS_ACCOUNT_ID="<your account id>" - [REQUIRED] - This defaults to "1111222233334444" and must be changed
AWS_DEFAULT_REGION="<your deploy region>" - This defaults to "us-east-2"
AWS_PROFILE="<your profile name>" - This defaults to 'default', this is the AWS CLI Profile name.
GROANBOT_BUILD_STAGE="<your build stage>" - This defaults to "prod"
```

Build the binary by running

```make```

Deploy using:

```make deploy```