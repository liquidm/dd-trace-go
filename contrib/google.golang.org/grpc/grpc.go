// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

//go:generate protoc -I . fixtures_test.proto --go_out=plugins=grpc:.

// Package grpc provides functions to trace the google.golang.org/grpc package v1.2.
package grpc // import "github.com/liquidm/dd-trace-go.v1/contrib/google.golang.org/grpc"

import (
	"io"

	"github.com/liquidm/dd-trace-go.v1/contrib/google.golang.org/internal/grpcutil"
	"github.com/liquidm/dd-trace-go.v1/ddtrace"
	"github.com/liquidm/dd-trace-go.v1/ddtrace/ext"
	"github.com/liquidm/dd-trace-go.v1/ddtrace/tracer"

	context "golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func startSpanFromContext(
	ctx context.Context, method, operation, service string, opts ...tracer.StartSpanOption,
) (ddtrace.Span, context.Context) {
	opts = append(opts,
		tracer.ServiceName(service),
		tracer.ResourceName(method),
		tracer.Tag(tagMethodName, method),
		tracer.SpanType(ext.AppTypeRPC),
	)
	md, _ := metadata.FromIncomingContext(ctx) // nil is ok
	if sctx, err := tracer.Extract(grpcutil.MDCarrier(md)); err == nil {
		opts = append(opts, tracer.ChildOf(sctx))
	}
	return tracer.StartSpanFromContext(ctx, operation, opts...)
}

// finishWithError applies finish option and a tag with gRPC status code, disregarding OK, EOF and Canceled errors.
func finishWithError(span ddtrace.Span, err error, cfg *config) {
	if err == io.EOF || err == context.Canceled {
		err = nil
	}
	errcode := status.Code(err)
	if errcode == codes.OK || cfg.nonErrorCodes[errcode] {
		err = nil
	}
	span.SetTag(tagCode, errcode.String())

	// only allocate finishOptions if needed, and allocate the exact right size
	var finishOptions []tracer.FinishOption
	if err != nil {
		if cfg.noDebugStack {
			finishOptions = []tracer.FinishOption{tracer.WithError(err), tracer.NoDebugStack()}
		} else {
			finishOptions = []tracer.FinishOption{tracer.WithError(err)}
		}
	}
	span.Finish(finishOptions...)
}
