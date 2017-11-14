# Development

## System Setup
Install GO - https://golang.org/doc/install  
Install dep - https://github.com/golang/dep

## Get Code
```
go get github.com/iReflect/reflect-app
cd ~/go/src/github.com/iReflect/reflect-app
git checkout develop
```

## Build
```
make build
```

## Run
```
make run
```
Vist API at - http://localhost:3000/  
Visit Admin at - http://localhost:9000/admin/  

## Run Tests
```
make test
```

# Migrations
```
make migrate up
make migrate status
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