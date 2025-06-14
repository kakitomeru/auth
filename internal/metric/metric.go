package metric

import (
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter                    metric.Meter
	UserRegistrationsCounter metric.Int64Counter
	SuccessfulLoginsCounter  metric.Int64Counter
	FailedLoginsCounter      metric.Int64Counter
	ActiveSessionsCounter    metric.Int64UpDownCounter
)

func Init() {
	meter = otel.GetMeterProvider().Meter("nota.auth.service")
	var err error

	UserRegistrationsCounter, err = meter.Int64Counter(
		"auth.registrations.successful.total",
		metric.WithDescription("Total number of successful registrations"),
		metric.WithUnit("{registrations}"),
	)
	if err != nil {
		log.Fatalf("failed to create user registrations counter: %v", err)
	}

	SuccessfulLoginsCounter, err = meter.Int64Counter(
		"auth.logins.successful.total",
		metric.WithDescription("Total number of successful logins"),
		metric.WithUnit("{logins}"),
	)
	if err != nil {
		log.Fatalf("failed to create successful logins counter: %v", err)
	}

	FailedLoginsCounter, err = meter.Int64Counter(
		"auth.logins.failed.total",
		metric.WithDescription("Total number of failed logins"),
		metric.WithUnit("{logins}"),
	)
	if err != nil {
		log.Fatalf("failed to create failed logins counter: %v", err)
	}

	ActiveSessionsCounter, err = meter.Int64UpDownCounter(
		"auth.sessions.active.total",
		metric.WithDescription("Total number of active sessions"),
		metric.WithUnit("{sessions}"),
	)
	if err != nil {
		log.Fatalf("failed to create active sessions counter: %v", err)
	}
}
