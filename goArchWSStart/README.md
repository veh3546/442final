# Checkers Turn Tracker

A simple Go web application demonstrating layered architecture (data access, business logic, service/HTTP layer) plus a tiny front-end. It simulates tracking whose turn it is among a fixed set of players.

## Prerequisites
- Go 1.25+ (module file specifies 1.25.1)

## Project Structure
```
checkers/
  main.go                # Server entry point
  business_logic/        # Domain logic (turns, login token generator)
  data_access/           # Database communication
  service/               # HTTP handlers & middleware
  static/                # Front-end assets (index.html, assets/...)
```

## Run (development)

```
go run .
```
Or explicitly:
```
go run main.go
```
The server listens on: `http://localhost:8080`

Then open a browser to:
```
http://localhost:8080/
```

## Build (optional)
```
go build -o checkers-server .
./checkers-server
```

## Add gorilla/websocket dependency
```
go get github.com/gorilla/websocket
```

## Go Watch Runners
Go-specific: https://github.com/mitranim/gow

Run with:
```
gow run .
```
<br>
<br>
General purpose: https://www.npmjs.com/package/nodemon