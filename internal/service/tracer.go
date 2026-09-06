package service

import (
	"go.opentelemetry.io/otel"
)

//nolint:gochecknoglobals // this is why tracer is a package-level immutable instrumentation handle.
var tracer = otel.Tracer("42tokyo-road-to-dena-server/internal/service")
