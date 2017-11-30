# Development

[![CircleCI](https://circleci.com/gh/iReflect/reflect-app.svg?style=svg)](https://circleci.com/gh/iReflect/reflect-app)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/21adfd09348b4de5b1aaec650a2d7462)](https://www.codacy.com/app/iReflect/reflect-app?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=iReflect/reflect-app&amp;utm_campaign=Badge_Grade)
[![codecov](https://codecov.io/gh/iReflect/reflect-app/branch/master/graph/badge.svg)](https://codecov.io/gh/iReflect/reflect-app)


## System Setup
Install GO - https://golang.org/doc/install  
Install dep - https://github.com/golang/dep

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
git checkout develop
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

# References:
- https://github.com/golang/dep
- https://github.com/gin-gonic/gin
- https://github.com/jinzhu/gorm
- https://github.com/pressly/goose

[0]: Links:
[1]: https://github.com/pressly/goose



# TODO:
- API Auth
- Admin Auth
- API Logging
