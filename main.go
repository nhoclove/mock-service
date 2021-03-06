package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	addr = ":3000"

	mGet  = "GET"
	mPost = "POST"
)

func main() {
	router, err := initRouter()
	if err != nil {
		log.Fatalf("Failed to init router: %v", err)
	}

	log.Println("############### ROUTES ################")
	for _, route := range router.Routes() {
		log.Println(route)
	}
	log.Println("############### ROUTES ################")

	log.Fatal(http.ListenAndServe(addr, router))
}

func initRouter() (*mux, error) {
	services, err := getServices()
	if err != nil {
		return nil, err
	}

	router := newMux()
	loggingMiddleware := logging()
	for _, service := range services {
		s := service
		if s.Method == mGet {
			router.GET(s.Path, loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(s.Response.StatusCode)

				for k, v := range s.Response.Header {
					w.Header().Add(k, v)
				}

				w.Write([]byte(s.Response.Body))
			}))
			continue
		}

		if s.Method == mPost {
			router.POST(s.Path, loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(s.Response.StatusCode)

				for k, v := range s.Response.Header {
					w.Header().Add(k, v)
				}

				w.Write([]byte(s.Response.Body))

			}))
			continue
		}

		log.Printf("Unsupported method: %s", service.Method)
	}
	return router, nil
}

type mux struct {
	handlers map[string]func(http.ResponseWriter, *http.Request)
}

func newMux() *mux {
	return &mux{
		handlers: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
}

// Routes returns all routes have been registered
func (m *mux) Routes() []string {
	var routes []string
	for k := range m.handlers {
		routes = append(routes, k)
	}

	return routes
}

// ServeHTTP is called by go server
// Its main functionality is to find a handler for the given request
func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := m.handlers[fmt.Sprintf("%s:%s", r.Method, r.URL.Path)]
	if !ok {
		notFound(w)
		return
	}

	f(w, r)
}

// This function is called when we can not find a handler for the provided path
func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (m *mux) GET(path string, handler http.HandlerFunc) {
	m.handlers[fmt.Sprintf("%s:%s", mGet, path)] = handler
}

func (m *mux) POST(path string, handler http.HandlerFunc) {
	m.handlers[fmt.Sprintf("%s:%s", mPost, path)] = handler
}

// ----------------------- Read data from JSON file --------------------

type (
	request struct {
		Header map[string]string
		Body   string
	}

	response struct {
		StatusCode int
		Header     map[string]string
		Body       string
	}

	service struct {
		Path     string
		Method   string
		Request  request
		Response response
	}
)

func getServices() ([]service, error) {
	var services []service
	err := filepath.Walk("./data", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		defer f.Close()

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		var service service
		if err = json.Unmarshal(bytes, &service); err != nil {
			return err
		}

		services = append(services, service)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

// -------------------- Middleware ------------------------------
type middleware func(http.HandlerFunc) http.HandlerFunc

func logging() middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {
			preRequest(r)
			rw := ResponseWriter{w, bytes.NewBuffer([]byte("")), 0}

			// Call the next middleware/handler in the chain
			next(&rw, r)

			postRequest(rw)
		}
	}
}

func preRequest(r *http.Request) {
	path := r.URL.Path
	method := r.Method

	if method != mGet {
		if r.Body == nil {
			log.Printf("Request started - path: %v - method: %s", path, method)
		}

		buf, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

		log.Printf("Request started - path: %v - method: %s - body: %s", path, method, string(buf))
		return
	}

	log.Printf("Request started - path: %v - method: %s", path, method)
}

func postRequest(w ResponseWriter) {
	log.Printf("Request Ended - status: %d - body: %v", w.StatusCode, w.Data.String())
}

type ResponseWriter struct {
	http.ResponseWriter
	Data       *bytes.Buffer
	StatusCode int
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	r.Data.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *ResponseWriter) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}
