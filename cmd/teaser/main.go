package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shibafu528/teaser/precum"
)

func main() {
	addr := ":8080"
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		addr = ":" + port
	}
	s := newServer(addr)

	// handle SIGINT, SIGTERM
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		sig := <-sigch
		log.Printf("received signal %v, exiting gracefully...", sig)
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("error in shutdown server: %v", err)
		}
		wg.Done()
	}()

	// start http server
	log.Printf("http server started on %s", addr)
	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
	wg.Wait()
}

func newServer(addr string) *http.Server {
	m := http.NewServeMux()
	m.HandleFunc("/", logger(handlerWrapper(handler)))
	return &http.Server{
		Addr:    addr,
		Handler: m,
	}
}

func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s \"%s %s\" \"%s\"", r.RemoteAddr, r.Method, r.URL, r.UserAgent())
		next(w, r)
	}
}

type Handler func(r *http.Request) (*Response, error)

type Response struct {
	Code int
	Body interface{}
}

var internalErrorJson = []byte(`{"message":"internal error"}`)

func (r *Response) Write(w http.ResponseWriter) error {
	var body []byte
	switch b := r.Body.(type) {
	case []byte:
		body = b
	case string:
		body = []byte(b)
	default:
		j, err := json.Marshal(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(internalErrorJson)
			return err
		}
		body = j
	}

	w.WriteHeader(r.Code)
	_, err := w.Write(body)
	return err
}

func ErrorMessage(code int, msg string) *Response {
	return &Response{Code: code, Body: errorResponse{msg}}
}

type errorResponse struct {
	Error string `json:"error"`
}

func handlerWrapper(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := handler(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(internalErrorJson)
			log.Printf("error: %v", err)
			return
		}

		err = res.Write(w)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}
}

func handler(r *http.Request) (*Response, error) {
	if r.Method != http.MethodGet || r.URL.Path != "/" {
		return ErrorMessage(http.StatusNotFound, "not found"), nil
	}

	u := r.URL.Query().Get("url")
	if u == "" {
		return ErrorMessage(http.StatusBadRequest, "parameter url is required"), nil
	}

	target, err := url.Parse(u)
	if err != nil {
		return ErrorMessage(http.StatusBadRequest, "url is invalid URL"), nil
	}

	switch target.Scheme {
	case "http", "https":
		// ok
	default:
		return ErrorMessage(http.StatusBadRequest, "url has invalid scheme, must be http or https"), nil
	}

	material, err := precum.Resolve(r.Context(), u)
	if err != nil {
		return nil, err
	}

	return &Response{Code: http.StatusOK, Body: material}, nil
}
