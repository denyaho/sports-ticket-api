package handler

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("42tokyo-road-to-dena-server/internal/handler")