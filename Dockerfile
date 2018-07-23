FROM alpine:latest

ARG IR_VERSION
ENV IR_VERSION=${IR_VERSION}

ADD .gopath/bin/reflect-app-${IR_VERSION} /usr/src/app

# Create app directory
WORKDIR /usr/src/app

RUN chmod +x ./reflect-app

CMD ["./reflect-app"]