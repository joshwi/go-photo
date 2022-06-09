FROM golang:1.16-alpine
RUN go env -w GO111MODULE=on
RUN apk add --no-cache git
RUN apk add --no-cache exiftool

# Set working directory for docker image
WORKDIR /go-media

RUN mkdir source
RUN mkdir target

RUN chmod 755 /go-media/source
RUN chmod 755 /go-media/target

# Copy go.mod and go.sum
COPY go.mod .
COPY go.sum .

# Copy go code and shell scripts
COPY app app/
COPY scripts scripts/
COPY config config/

# Install go module dependencies
RUN go mod tidy

RUN rm -rf /go-media/app/builds

# Run build script
RUN sh scripts/build.sh

RUN chmod 755 /go-media/scripts/*
RUN chmod 755 /go-media/app/builds/*

RUN /usr/bin/crontab /go-media/config/cron.txt

CMD ["sh", "scripts/start.sh"]