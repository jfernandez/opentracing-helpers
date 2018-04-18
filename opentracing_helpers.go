package opentracing_helpers

import (
	"net/http"
	"github.com/opentracing/opentracing-go"
)

// TraceHandler facilitates tracing of handlers registered with an
// http.ServeMux.  For example, to trace this code:
//
//    http.Handle("/foo", fooHandler)
//
// Perform this replacement:
//
//    http.Handle(opentracing_helpers.TraceHandler("/foo", fooHandler))
//
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
