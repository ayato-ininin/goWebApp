# WebApplication and API in [Go](https://golang.org/dl/)

* A simple web application that allows a user to log in and upload a profile picture.
* A simple REST API built on the same code base as the web application which allows users to authenticate using JWT tokens and perform operations against a Postgres database. We'll go through the entire authentication process, including using refresh tokens, and thoroughly test all aspects of the code.
* A simple Single Page Web Application (SPA), written in Vanilla JavaScript, that demonstrates how to use JWT and Refresh Tokens with a SPA, and how to test that functionality.

## Prerequisites

Before you begin, ensure you have met the following requirements:

* You have installed the latest version of [Go](https://golang.org/dl/)
* You have a `Unix/Linux` machine.

## Using GoProject

This project uses Docker for postgres. To make sure Docker is installed and running on your machine.

```bash
docker compose up
```

To start the Web Application service, run the following command:

```bash
go run ./cmd/web
```

To start the API server, run the following command:

```bash
go run ./cmd/api
```

To generate a token, so that we can test our api, run the following command:

```bash
go run ./cmd/cli -action=valid     // will produce a valid token
go run ./cmd/cli -action=expired   // will produce an expired token
```

## Running Tests

To run tests, use the following command:

```bash
go test -v .
go test -v -tags=integration ./pkg/repository/dbrepo
go test -cover . && go test . -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## References

This project was made from the following resources:

1. [Introduction to Testing in Go (Golang)](https://www.udemy.com/course/introduction-to-testing-in-go-golang/) - udemy course.
