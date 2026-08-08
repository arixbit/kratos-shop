package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func Init(addr string) error {
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		_ = http.ListenAndServe(addr, mux)
	}()
	return nil
}
