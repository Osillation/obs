"""
Flask OTel instrumentation.
Import and call configure_otel() before creating the Flask app:
  from instrumentation import configure_otel, instrument_app
  configure_otel()
  app = Flask(__name__)
  instrument_app(app)

Install dependencies:
  pip install \
    opentelemetry-distro \
    opentelemetry-exporter-otlp-proto-grpc \
    opentelemetry-instrumentation-flask \
    opentelemetry-instrumentation-sqlalchemy \
    opentelemetry-instrumentation-requests \
    opentelemetry-instrumentation-logging

Required env vars:
  OTEL_EXPORTER_OTLP_ENDPOINT=http://obs-otel-collector:4317
  OTEL_SERVICE_NAME=flask-service
  OTEL_SERVICE_NAMESPACE=client-name
"""

import os

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_NAMESPACE
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def configure_otel() -> None:
    """Initialize OTel SDK. Call before creating the Flask app instance."""

    resource = Resource.create({
        SERVICE_NAME: os.getenv("OTEL_SERVICE_NAME", "flask-service"),
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
    RequestsInstrumentor().instrument()


def instrument_app(app) -> None:
    """Call after app = Flask(__name__) to instrument all routes."""
    FlaskInstrumentor().instrument_app(app)
