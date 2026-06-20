"""
FastAPI OTel instrumentation.
Import this module at the top of main.py before creating the FastAPI app:
  from instrumentation import configure_otel, instrument_app

Then call them:
  configure_otel()
  app = FastAPI()
  instrument_app(app)

Install dependencies:
  pip install \
    opentelemetry-distro \
    opentelemetry-exporter-otlp-proto-grpc \
    opentelemetry-instrumentation-fastapi \
    opentelemetry-instrumentation-sqlalchemy \
    opentelemetry-instrumentation-asyncpg \
    opentelemetry-instrumentation-httpx \
    opentelemetry-instrumentation-logging

Required env vars:
  OTEL_EXPORTER_OTLP_ENDPOINT=http://obs-otel-collector:4317
  OTEL_SERVICE_NAME=fastapi-service
  OTEL_SERVICE_NAMESPACE=client-name
"""

import os

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_NAMESPACE
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor


def configure_otel() -> None:
    """Initialize OTel SDK. Call before creating the FastAPI app instance."""
    resource = Resource.create({
        SERVICE_NAME: os.getenv("OTEL_SERVICE_NAME", "fastapi-service"),
        SERVICE_NAMESPACE: os.getenv("OTEL_SERVICE_NAMESPACE", "default"),
    })

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(
        BatchSpanProcessor(
            OTLPSpanExporter(
                endpoint=os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
                insecure=True,
            )
        )
    )
    trace.set_tracer_provider(provider)

    LoggingInstrumentor().instrument(set_logging_format=True)
    HTTPXClientInstrumentor().instrument()


def instrument_app(app) -> None:
    """Call after app = FastAPI() to instrument all routes."""
    FastAPIInstrumentor.instrument_app(app)
