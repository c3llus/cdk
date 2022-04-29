package http

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	router       *mux.Router
	metrics      map[string]Metrics
	defaultProme bool // flag to add prome wrapper on every handler
}

type httpError struct {
	HTTPStatus int
	Message    string
}

func (e httpError) Error() string {
	return e.Message
}

type HandlerFunc func(ctx CtxHTTP) error

// HandleFunc register tdk handler to serve the given path and http methods
func (r *Router) HandleFunc(path string, handler HandlerFunc, methods ...string) {
	for _, method := range methods {
		route := r.router.Methods(method).Path(path)
		route.HandlerFunc(r.wrapCtxHTTP(route, handler))
	}
}

type CtxHTTP interface {
	// Request returns the underlying request
	Request() *http.Request

	// Writer return the underlying response writer
	Writer() http.ResponseWriter

	// Body returns the request body
	Body() ([]byte, error)

	// JSON write json response with the given http status code
	JSON(code int, data interface{}) error

	// Write raw []byte response with the given http status code
	Write(code int, data []byte) (int, error)

	// DecodeJSONBody json.Unmarshal request body to data
	DecodeJSONBody(data interface{}) error
}

type defaultContext struct {
	writer  http.ResponseWriter
	request *http.Request
	body    []byte
	vars    map[string]string
}

// Writer return default http.ResponseWriter
func (c *defaultContext) Writer() http.ResponseWriter {
	return c.writer
}

// Request return default http.Request
func (c *defaultContext) Request() *http.Request {
	return c.request
}

// For quicker body retrieval
// also for handy mocking
func (c *defaultContext) Body() ([]byte, error) {
	if c.body == nil {
		b, err := ioutil.ReadAll(c.request.Body)
		if err != nil {
			return nil, err
		}
		c.body = b
	}
	return c.body, nil
}

// JSON for rendering json data
func (c *defaultContext) JSON(code int, data interface{}) error {
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	return json.NewEncoder(c.writer).Encode(data)
}

// easier access for write
// also handy for testing
func (c *defaultContext) Write(code int, resp []byte) (int, error) {
	c.writer.WriteHeader(code)
	return c.writer.Write(resp)
}

func (c *defaultContext) DecodeJSONBody(data interface{}) error {
	body, err := c.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}

func (r *Router) wrapCtxHTTP(route *mux.Route, handler HandlerFunc) http.HandlerFunc {
	stdHandler := toStdHandler(handler)
	return r.wrapWithProme(route, stdHandler)
}

func toStdHandler(handler HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// create tdk context
		ctx := newContext(w, r)

		// executes handler with chained middleware
		err := handler(ctx)
		if err == nil {
			return // if no error, all done
		}

		// executes the error handler
		status := http.StatusInternalServerError
		httpError, ok := err.(httpError)
		if ok {
			status = httpError.HTTPStatus
		}
		ctx.Write(status, []byte(err.Error()))

	}
}

type Metrics struct {
	CounterVec   *prometheus.CounterVec
	HistogramVec *prometheus.HistogramVec
}

// this will convert path string into prometheus label
var regexCleanPath = regexp.MustCompile(`{|(:.*})|}`)

// wrap the given handler with prometheus metrics, if `defaultProme` flag is on
func (r *Router) wrapWithProme(route *mux.Route, handler http.HandlerFunc) http.HandlerFunc {
	if r.defaultProme {
		return r.doWrapWithProme(route, handler)
	}
	return handler
}

func convertPathToName(path string) string {
	if path == "" {
		return ""
	}
	byt := regexCleanPath.ReplaceAll([]byte(path), []byte(""))
	name := strings.Replace(string(byt), "/", "_", -1)

	name = name[1:]
	if name == "" {
		name = "index"
	}
	return "http_" + name
}

func (r *Router) doWrapWithProme(route *mux.Route, handler http.HandlerFunc) http.HandlerFunc {
	var (
		countervec   *prometheus.CounterVec
		durationhist *prometheus.HistogramVec
	)

	if r.metrics == nil {
		r.metrics = make(map[string]Metrics)
	}

	path, err := route.GetPathTemplate()
	if err != nil {
		log.Fatal(err)
	}

	lblHandler := convertPathToName(path)

	if m, ok := r.metrics[lblHandler]; ok {
		countervec = m.CounterVec
		durationhist = m.HistogramVec
	} else {
		countervec = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: lblHandler + "_counter",
				Help: "A counter for requests to the wrapped handler.",
			},
			[]string{"code", "method"},
		)

		durationhist = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: lblHandler + "_duration",
				Help: "A histogram of request latencies.",
			},
			[]string{"code", "method"},
		)

		err = prometheus.DefaultRegisterer.Register(countervec)
		if err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				log.Fatal(err)
			}
		}

		err = prometheus.DefaultRegisterer.Register(durationhist)
		if err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				log.Fatal(err)
			}
		}

		mtx := Metrics{
			CounterVec:   countervec,
			HistogramVec: durationhist,
		}
		r.metrics[lblHandler] = mtx
	}

	chain := promhttp.InstrumentHandlerCounter(countervec,
		promhttp.InstrumentHandlerDuration(durationhist,
			handler,
		),
	)
	return chain
}

func newContext(w http.ResponseWriter, r *http.Request) CtxHTTP {
	return &defaultContext{
		writer:  w,
		request: r,
	}
}
