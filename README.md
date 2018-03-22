# Development

[![CircleCI](https://circleci.com/gh/iReflect/reflect-app.svg?style=svg)](https://circleci.com/gh/iReflect/reflect-app)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/21adfd09348b4de5b1aaec650a2d7462)](https://www.codacy.com/app/iReflect/reflect-app?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=iReflect/reflect-app&amp;utm_campaign=Badge_Grade)


## System Setup
Install GO - https://golang.org/doc/install  
Install dep - https://github.com/golang/dep   
Install Redis - https://redis.io/topics/quickstart

## Prepare DB
Create a database (postgresql) as below to use the default configuration,
```
host=localhost
user=ireflect
password=1Reflect
dbname=ireflect-dev
```

You can override the default DB connection information by setting an ENV variable
```
export DB_DSN="host=localhost user=ireflect password=1Reflect dbname=ireflect-dev"
export DB_DRIVER="mysql"
```

## Get Code
```
go get -d github.com/iReflect/reflect-app
cd ~/go/src/github.com/iReflect/reflect-app
```


## Pull vendor dependencies
```
make vendor
```

## Build
```
make all
```

## Run
```
make run
```
Vist API at - http://localhost:3000/  
Visit Admin at - http://localhost:3000/admin/  

## Run Tests
```
make test
```

# Migrations
```
make migrate up
make migrate status
```

# Adding Migrations
Examples:
```
make migrate create <migration_name> go

make migrate create <migration_name> sql
```

For help - make migrate

# Adding Dependencies
```
dep ensure -add github.com/foo/bar
```

# Time Tracker configuration
First, Generate a Refresh token using.
https://developers.google.com/oauthplayground/ with Timesheet App's client_id, client_secret and following scopes

```
https://www.googleapis.com/auth/spreadsheets
https://www.googleapis.com/auth/userinfo.email
```

Video instructions at https://www.youtube.com/watch?v=PJWrjAuIWWo


Use the Refresh token to create a JSON credentials file at `config/timetracker_credentials.json` using following format

```json
{
    "type":"authorized_user",
    "client_id":"xxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "client_secret":"xxxxxxxxx",
    "refresh_token": "xxxxxxxxx"
}
```

# Google OAuth Login Configuration
Generate a client_id/client_secret for the iReflect's Authentication app.
Note, select Web Application as the application type and provide origin and redirect url of the hosted webapp

Use the generated client_id/client_secret to create a JSON credentials file at `config/application_default_credentials.json` using following format  
```
{
  "type": "authorized_user",
  "web": {
    "client_id": "xxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "client_secret": "xxxxxxxxxxxxx",
    "redirect_uris": [
      "http://localhost:4200/auth"
    ]
  }
}
```    

# Sentry Logging
Specify an environment variable `SENTRY_DSN` to enable sentry logging for errors
```
SENTRY_DSN = https://<key>:<secret>@sentry.io/<project>
```

# References:
- https://github.com/golang/dep
- https://github.com/gin-gonic/gin
- https://github.com/jinzhu/gorm
- https://github.com/pressly/goose

# TODO:
- API Auth
- Admin Auth
- API Logging
