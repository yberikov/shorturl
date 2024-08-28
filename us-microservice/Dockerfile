FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

ADD ./ /app

RUN CGO_ENABLED=0 GOOS=linux go build -o build/app cmd/app/main.go

FROM alpine:latest

RUN apk add --no-cache bash

COPY --from=build /app/build/* /opt/

COPY wait-for-it.sh /opt/wait-for-it.sh

COPY config ./config

# Make the wait-for-it script executable
RUN chmod +x /opt/wait-for-it.sh

# Set the entrypoint to use wait-for-it.sh to wait for dependencies
ENTRYPOINT ["/opt/wait-for-it.sh", "mongodb-url:27017", "--", "/opt/wait-for-it.sh", "kafka1:19092", "--", "/opt/wait-for-it.sh", "redis:6379", "--", "/opt/app"]


EXPOSE 44045
# Command to run the executable
CMD ["run"]