// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

// Package echo provides functions to trace the labstack/echo package (https://github.com/labstack/echo).
package echo

import (
	"math"
	"strconv"

	"github.com/liquidm/dd-trace-go.v1/ddtrace"
	"github.com/liquidm/dd-trace-go.v1/ddtrace/ext"
	"github.com/liquidm/dd-trace-go.v1/ddtrace/tracer"
	"github.com/liquidm/dd-trace-go.v1/internal/appsec"

	"github.com/labstack/echo/v4"
)

// Middleware returns echo middleware which will trace incoming requests.
func Middleware(opts ...Option) echo.MiddlewareFunc {
	appsecEnabled := appsec.Enabled()
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			resource := request.Method + " " + c.Path()
			opts := []ddtrace.StartSpanOption{
				tracer.ServiceName(cfg.serviceName),
				tracer.ResourceName(resource),
				tracer.SpanType(ext.SpanTypeWeb),
				tracer.Tag(ext.HTTPMethod, request.Method),
				tracer.Tag(ext.HTTPURL, request.URL.Path),
				tracer.Measured(),
			}

			if !math.IsNaN(cfg.analyticsRate) {
				opts = append(opts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
			}
			if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(request.Header)); err == nil {
				opts = append(opts, tracer.ChildOf(spanctx))
			}
			var finishOpts []tracer.FinishOption
			if cfg.noDebugStack {
				finishOpts = append(finishOpts, tracer.NoDebugStack())
			}
			span, ctx := tracer.StartSpanFromContext(request.Context(), "http.request", opts...)
			defer func() { span.Finish(finishOpts...) }()

			// pass the span through the request context
			c.SetRequest(request.WithContext(ctx))
			// serve the request to the next middleware
			if appsecEnabled {
				afterMiddleware := useAppSec(c, span)
				defer afterMiddleware()
			}
			err := next(c)
			if err != nil {
				finishOpts = append(finishOpts, tracer.WithError(err))
				// invokes the registered HTTP error handler
				c.Error(err)
			}

			span.SetTag(ext.HTTPCode, strconv.Itoa(c.Response().Status))
			return err
		}
	}
}
