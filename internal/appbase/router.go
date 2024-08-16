package appbase

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"time"
	"ulascansenturk/service/openapi"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

const ApplicationJSONType = "application/json"

func NewRouterMux(serviceName string, logger *zerolog.Logger, openAPIMiddleware *openapi.ValidationMiddleware, timeout time.Duration, db *gorm.DB) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.SetHeader("Content-Type", ApplicationJSONType))
	mux.Use(WithLogger(*logger))
	mux.Use(WithErrorLogs())
	mux.Use(chiMiddleware.Timeout(timeout))

	// Add GORM middleware
	mux.Use(GormMiddleware(db))

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	mux.Use(sentryMiddleware.Handle)

	mux.Use(chiMiddleware.Heartbeat("/"))
	mux.Use(chiMiddleware.Heartbeat("/healthz"))
	mux.Use(chiMiddleware.Heartbeat("/readyz"))

	mux.Use(openAPIMiddleware.Handler())

	return mux
}

func WithLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.WithContext(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func WithErrorLogs() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			recorder := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			if recorder.status >= http.StatusBadRequest && recorder.status <= http.StatusNetworkAuthenticationRequired {
				log.Ctx(r.Context()).Error().
					RawJSON("error_body", compactJSON(recorder.body.Bytes())).
					Str("http_method", r.Method).
					Str("uri", r.RequestURI).
					Int("http_status", recorder.status).
					Str("http_status_text", http.StatusText(recorder.status)).
					Msg("http error")
			}
		}

		return http.HandlerFunc(fn)
	}
}

type responseRecorder struct {
	http.ResponseWriter

	body   bytes.Buffer
	status int

	loggedStatusHeader bool
	loggedBody         bool
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.loggedStatusHeader {
		return
	}

	r.status = statusCode
	r.loggedStatusHeader = true

	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.loggedBody {
		return 0, nil
	}

	r.body.Write(body)
	r.loggedBody = true

	return r.ResponseWriter.Write(body)
}

func compactJSON(src []byte) []byte {
	if !isValidJSON(src) {
		rawRespBody := &rawResponseBody{RawBody: string(src)}
		wrappedJSON, _ := json.Marshal(rawRespBody)

		return wrappedJSON
	}

	dst := &bytes.Buffer{}
	_ = json.Compact(dst, src)

	return dst.Bytes()
}

func isValidJSON(raw []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(raw, &js) == nil
}

type rawResponseBody struct {
	RawBody string `json:"raw_body"`
}
