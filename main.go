package main

import (
	"bufio"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var embeddingMap EmbeddingMap

func init() {
	log.Print("loading word embeddings")
	embeddings, err := loadGloveEmbeddings()
	if err != nil {
		log.Fatal(err)
	}
	embeddingMap = embeddings
	log.Print("loaded word embeddings")
}

func main() {
	router := mux.NewRouter()
	router.Use(loggingMiddleware())
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("word")
		if word == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		embedding, exists := embeddingMap[word]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := writeResponseAsJson(w, embedding); err != nil {
			panic(err)
		}
	}).Methods(http.MethodGet)

	s := &http.Server{
		Addr:           ":9090",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Print("serving requests")
	log.Fatal(s.ListenAndServe())
}

type responseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (w *responseWriter) WriteStatusCode(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.StatusCode = code
}

func loggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			ww := newResponseWriter(w)
			defer func() {
				log.Printf(
					"[%s] [%v] [%d] %s %s %s",
					req.Method,
					time.Since(start),
					ww.StatusCode,
					req.Host,
					req.URL.Path,
					req.URL.RawQuery,
				)
			}()
			next.ServeHTTP(ww, req)
		})
	}
}

func writeResponseAsJson(w http.ResponseWriter, v interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

type Embedding []float64
type EmbeddingMap map[string]Embedding

func loadGloveEmbeddings() (EmbeddingMap, error) {
	embeddingFile, err := os.Open(os.Getenv("MODEL_FILE"))
	if err != nil {
		return nil, err
	}
	return loadGloveFile(embeddingFile)
}

func loadGloveFile(r io.Reader) (EmbeddingMap, error) {
	embeddings := make(EmbeddingMap, 0)
	s := bufio.NewScanner(r)
	lines := make(chan string)
	go func() {
		defer close(lines)
		for s.Scan() {
			lines <- s.Text()
		}
	}()
	for line := range lines {
		arr := strings.Split(line, " ")
		e := make(Embedding, len(arr)-1)
		for i := 1; i < len(arr); i++ {
			n, err := strconv.ParseFloat(arr[i], 64)
			if err != nil {
				return nil, err
			}
			e[i-1] = n
		}
		embeddings[arr[0]] = e
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return embeddings, nil
}
