package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/headers"
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
		handler400(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerProxy(w, req)
		return
	}

	handler200(w, req)
}

func writeResponse(w *response.Writer, statusCode response.StatusCode, contentType string, body []byte) {
	if err := w.WriteStatusLine(statusCode); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Override("content-type", contentType)
	if err := w.WriteHeaders(headers); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	if _, err := w.WriteBody(body); err != nil {
		log.Printf("Error writing body: %v", err)
	}
}

func handler400(w *response.Writer, _ *request.Request) {
	body := []byte(`
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
	writeResponse(w, response.StatusBadRequest, "text/html", body)
}

func handler500(w *response.Writer, _ *request.Request) {
	body := []byte(`
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
	writeResponse(w, response.StatusInternalServerError, "text/html", body)
}

func handler200(w *response.Writer, _ *request.Request) {
	body := []byte(`
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
	writeResponse(w, response.StatusOK, "text/html", body)
}

func handlerVideo(w *response.Writer, req *request.Request) {
	const filepath = "assets/vim.mp4"
	videoBytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Printf("Error reading video file: %v", err)
		handler500(w, req)
		return
	}

	writeResponse(w, response.StatusOK, "video/mp4", videoBytes)
}

func handlerProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	r, err := http.Get("https://httpbin.org/" + target)
	if err != nil {
		log.Printf("error proxying request: %v", err)
	}
	defer r.Body.Close()

	if err := w.WriteStatusLine(response.StatusCode(r.StatusCode)); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	if err := w.WriteHeaders(h); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	const maxChunkSize = 1024
	buff := make([]byte, maxChunkSize)
	fullBody := make([]byte, 0)
	for {
		n, err := r.Body.Read(buff)
		if n > 0 {
			w.WriteChunkedBody(buff[:n])
			fullBody = append(fullBody, buff[:n]...)
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("error reading response body: %v", err)
			break
		}

		_, err = w.WriteChunkedBody(buff[:n])
		if err != nil {
			log.Printf("error writing chunked body: %v", err)
		}
	}

	if _, err = w.WriteChunkedBodyDone(); err != nil {
		log.Printf("error writing chunked body done: %v", err)
	}

	hash := sha256.Sum256(fullBody)
	hashStr := fmt.Sprintf("%x", hash)
	trailers := headers.NewHeaders()
	trailers.Override("X-Content-SHA256", hashStr)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	if err = w.WriteTrailers(trailers); err != nil {
		log.Printf("error writing chunked body trailer: %v", err)
	}
}
