package opentracing_helpers

import (
	"net/http"
	"github.com/opentracing/opentracing-go"
)

// The Handle function is the primary way to set up your chain of middlewares to be called by rye.
// It returns a http.HandlerFunc from net/http that can be set as a route in your http server.
func TraceHandler(pattern string, handler http.Handler) (string, http.Handler) {
	return pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look for the request caller's SpanContext in the headers
		// If not found create a new SpanContext
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		parentSpanContext, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)

		var span opentracing.Span
		if parentSpanContext == nil {
			span = opentracing.StartSpan(pattern)
		} else {
			span = opentracing.StartSpan(pattern, opentracing.ChildOf(parentSpanContext))
		}
		defer span.Finish()
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

		handler.ServeHTTP(w, r)
	})
}
