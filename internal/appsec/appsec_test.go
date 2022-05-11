// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

//go:build appsec
// +build appsec

package appsec_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/liquidm/dd-trace-go/ddtrace/tracer"
	"github.com/liquidm/dd-trace-go/internal/appsec"
	"github.com/liquidm/dd-trace-go/internal/appsec/waf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnabled(t *testing.T) {
	enabledConfig, _ := strconv.ParseBool(os.Getenv("DD_APPSEC_ENABLED"))
	canBeEnabled := enabledConfig && waf.Health() == nil

	require.False(t, appsec.Enabled())
	tracer.Start()
	assert.Equal(t, canBeEnabled, appsec.Enabled())
	tracer.Stop()
	assert.False(t, appsec.Enabled())
}
