package opentracing_helpers

import (
	"net/http"
	"github.com/opentracing/opentracing-go"
	"net/http/httptrace"
	"context"
	"github.com/opentracing/opentracing-go/log"
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
		tracer := opentracing.GlobalTracer()
		parentSpanContext, _ := tracer.Extract(opentracing.HTTPHeaders, carrier)

		spanName := r.Method + " " + pattern
		var span opentracing.Span
		if parentSpanContext == nil {
			span = opentracing.StartSpan(spanName)
		} else {
			span = opentracing.StartSpan(spanName, opentracing.ChildOf(parentSpanContext))
		}
		defer span.Finish()
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

		handler.ServeHTTP(w, r)
	})
}

// TraceRequest facilities the tracing of a http.Request by injecting the
// span context into the request's headers. It uses the httptrace package
// to log events throughout the requests lifecycle. For example:
//
//    req, _ := http.NewRequest("GET", "http://example.com/", nil)
//    tracedReq, span := opentracing_helpers.TraceRequest("GET example.com", r.Context(), *req)
//    if _, err := http.DefaultTransport.RoundTrip(tracedReq); err != nil {
//	      span.SetTag("error", true)
//    }
//    span.Finish()

func TraceRequest(operationName string, ctx context.Context, r http.Request) (*http.Request, opentracing.Span) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			span.LogFields(
				log.String("event", "Get Connection "),
				log.String("host:port", hostPort),
			)
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			span.LogFields(
				log.String("event", "Got Connection"),
				log.Object("connection info", connInfo),
			)
		},
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			span.LogFields(
				log.String("event", "DNS Start"),
				log.Object("dns start info", dnsInfo),
			)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			span.LogFields(
				log.String("event", "DNS Done"),
				log.Object("dns done info", dnsInfo),
			)
		},
		ConnectDone: func(network, addr string, err error) {
			span.LogFields(
				log.String("event", "Connect Done"),
				log.Object("network", network),
				log.String("address", addr),
				log.Error(err),
			)
		},
		GotFirstResponseByte: func() {
			span.LogFields(log.String("event", "Got First Response Byte"))
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			span.LogFields(
				log.String("event", "Wrote Request"),
				log.Object("wrote request info", info),
			)
		},
	}

	return r.WithContext(httptrace.WithClientTrace(r.Context(), trace)), span
}
