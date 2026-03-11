package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/request"
	"github.com/dennisdijkstra/go-tcp-to-http/internal/response"
	"github.com/dennisdijkstra/go-tcp-to-http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		writeHTMLResponse(w, response.StatusBadRequest, body400)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		writeHTMLResponse(w, response.StatusInternalServerError, body500)
		return
	}

	writeHTMLResponse(w, response.StatusOK, body200)
}

func writeHTMLResponse(w *response.Writer, statusCode response.StatusCode, body []byte) {
	if err := w.WriteStatusLine(statusCode); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Override("content-type", "text/html")
	if err := w.WriteHeaders(headers); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	if _, err := w.WriteBody(body); err != nil {
		log.Printf("Error writing body: %v", err)
	}
}

var (
	body400 = []byte(`
		<html>
			<head>
				<title>400 Bad Request</title>
			</head>
			<body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
	`)

	body500 = []byte(`
		<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
	`)

	body200 = []byte(`
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`)
)
