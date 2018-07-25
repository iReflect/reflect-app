# iReflect Server-Side Application

[![CircleCI](https://circleci.com/gh/iReflect/reflect-app.svg?style=svg)](https://circleci.com/gh/iReflect/reflect-app)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/21adfd09348b4de5b1aaec650a2d7462)](https://www.codacy.com/app/iReflect/reflect-app?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=iReflect/reflect-app&amp;utm_campaign=Badge_Grade)


## Installation & Setup
#### Platform & tools
- Install the GoLang - https://golang.org/doc/install
> **Note:** If you have installed Go to a custom location, make sure the $GOROOT variable is set properly. Refer [Installing to a custom location](https://golang.org/doc/install#install).

- Install the Go dependency management tool `dep` - https://github.com/golang/dep

- Install Redis in your machine - https://redis.io/topics/quickstart

#### Get the Code
> Make sure to use `go get` command instead of `git clone` to download the repository since `go get` directly adds the repository according to the [$GOPATH](https://golang.org/doc/code.html#GOPATH) environment variable, which is required since go can only access the installed go binaries once the required environment variables are set.
```
go get -d github.com/iReflect/reflect-app
cd ~/go/src/github.com/iReflect/reflect-app
```

#### About Makefile
The `make` tool and the [Makefile](Makefile) file, present at the project root folder, can be used for wrapping Go commands with specific build targets that simplify usage on the command line. Some of the commands wrapped in the Makefile are used in the setup instructions given below, to know more about the wrapped commands refer to the Makefile.

#### Install vendor dependencies
1. This command (a wrapper for the `dep ensure` command) places all the dependencies in the vendor folder based on the [Gopkg.toml](Gopkg.toml) file. If a dependency is in the [Gopkg.lock](Gopkg.lock) file, use the version specified there otherwise use the most recent version.
    ```
    make vendor
    ```

 2. To add new dependencies, run the command given below:
    ```
    dep ensure -add github.com/foo/bar
    ```

## Database Configuration
iReflect only supports Postgresql as of now. So below mention configurations are postgresql specific.

Create a database as below to use the default configuration (or override the default values),
```
host=localhost
user=ireflect
password=1Reflect
dbname=ireflectdev
```

You can override the default DB connection information by setting an ENV variable
```
export DB_DSN="host=localhost user=ireflect password=1Reflect dbname=ireflectdev sslmode=disable"
export DB_DRIVER="postgres"
```

## Migrations Management
For help, run `make migrate` in the command line.

#### Applying migrations
All the migrations are present under the `db/migrations` folder (which can be configured using the `MigrationsDir` config under `config/config.go`)

Examples:
```
make migrate up
make migrate down
```
Refer https://github.com/pressly/goose

#### Adding new migrations
Examples:
```
make migrate create <migration_name> go
make migrate create <migration_name> sql
```

## Time Tracker configuration (For Google Sheets)

Visit https://developers.google.com/oauthplayground/ and follow the given steps to configure Google Sheets as a Time Tracker (video instructions at https://www.youtube.com/watch?v=PJWrjAuIWWo):

**Step 1:** Select the following two scopes from "Select & authorize APIs" section.
> You can also add the scopes manually in the text box provided there, by writing the scopes separated by comma.
```
https://www.googleapis.com/auth/spreadsheets
https://mail.google.com/
```

**Step 2:** Generate a Refresh token using the Google timesheet App's client_id, client_secret.

**Step 3:** Use the Refresh token to create a JSON credentials file at `config/timetracker_credentials.json` using the following format, use the same client_id and client_secret here, used in the previous step.

```json
{
    "type":"authorized_user",
    "client_id":"xxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "client_secret":"xxxxxxxxx",
    "refresh_token": "xxxxxxxxx"
}
```

## Login Configuration (For Google OAuth)

Visit https://console.cloud.google.com/apis/dashboard and follow the given steps to configure Google OAuth as the login method:

**Step 1:** Enable APIs and Services

**Step 2:** Create a new project, if not already created, and select that project.

**Step 3:** To create a new credential, you must first set a product name on the consent screen under the Credentials section (visible once you have selected a project). Go to the "OAuth consent screen" section and provide a product name which will be shown to the users whenever requesting access.

**Step 4:** Generate a pair of client_id and client_secret for the iReflect's Authentication app, using "Create Credentials".

> **Note:** Select `OAuth client ID` as the credentials type and then select Web Application as the application type and provide origin and redirect url of the hosted webapp.

Use the generated client_id and client_secret to create a JSON credentials file at `config/application_default_credentials.json` using following format
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

## Sentry Logging (Optional)
Specify an environment variable `SENTRY_DSN` to enable sentry logging for errors
```
SENTRY_DSN = https://<key>:<secret>@sentry.io/<project>
```

## Running the Application
Once everything is configured properly, run the below command to start the API server.
```
make run
```

## Accessing the Admin interface
Admin Interface is accessible at http://localhost:3000/admin/.

Before that, create an admin user using the command line. Execute the following command for postgresql:
> **Note:** The given SQL query is based on the current user model definition and assuming Google Sheets as the only available time tracker.
```
psql -U username -d database_name -c "INSERT INTO USERS (email, first_name, last_name, time_provider_config, is_admin) values ('<email_id>', '<first_name>', '<last_name>', '[{"data": {"email": "<email_id>"}, "type": "gsheet"}]', true)"
```

## Build the application
To generate a binary distribution file for the application, run the following command.
```
make all
```

## Running Tests
Execute the following command to run the tests.
```
make test
```
To learn how to write and execute test cases in Go, refer https://golang.org/pkg/testing/ .

# References:
- https://github.com/golang/dep
- https://github.com/gin-gonic/gin
- https://github.com/jinzhu/gorm
- https://github.com/pressly/goose
- https://github.com/qor/qor

# QOR Admin Ref
- [Building Admin](https://doc.getqor.com/admin/)
- [General Configuration](https://doc.getqor.com/admin/general.html)
- [Manage Resources](https://doc.getqor.com/admin/resources.html)
  - [Fields](https://doc.getqor.com/admin/fields.html)
  - [Data Validation](https://doc.getqor.com/admin/processing_validation.html)
- [Authentication](https://doc.getqor.com/admin/authentication.html)
- [Theming & Customization](https://doc.getqor.com/admin/theming_and_customization.html)
- [Extend QOR Admin](https://doc.getqor.com/admin/extend_admin.html)
- [Integrate with WEB frameworks](https://doc.getqor.com/admin/integration.html)
- [Deploy To Production](https://doc.getqor.com/admin/deploy.html)

## Contributing
- Get the package or fork the repo:
    ```
    go get -d github.com/iReflect/reflect-app
    ```
- Set your fork as a remote:
    ```
    git remote add fork git@github.com:GITHUB_USERNAME/reflect-app.git
    ```
- Make changes, commit to your fork.
- Send a pull request with your changes.
