FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

ADD ./ /app

RUN CGO_ENABLED=0 GOOS=linux go build -o build/app cmd/app/main.go

FROM alpine:latest

COPY --from=build /app/build/* /opt/

COPY config ./config

ENTRYPOINT [ "/opt/app" ]

EXPOSE 8080
# Command to run the executable
CMD ["run"]