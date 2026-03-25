package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-catalog-service/internal/catalog"
	pb "product-catalog-service/proto"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

func main() {
	// --- Tracing (OpenTelemetry → Grafana Tempo) ---
	// Set ENABLE_TRACING=1 and COLLECTOR_SERVICE_ADDR=<otel-collector:4317>
	if os.Getenv("ENABLE_TRACING") == "1" {
		if err := initTracing(); err != nil {
			log.Warnf("tracing init failed, continuing without it: %v", err)
		} else {
			log.Info("tracing enabled")
		}
	} else {
		log.Info("tracing disabled — set ENABLE_TRACING=1 to enable")
	}

	// --- Profiling (Pyroscope) ---
	// TODO: uncomment when wiring up the LGTM stack
	// if os.Getenv("ENABLE_PROFILING") == "1" {
	//     initPyroscope()
	// }

	// --- Extra latency injection for chaos/load testing ---
	var extraLatency time.Duration
	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("invalid EXTRA_LATENCY %q: %v", s, err)
		}
		extraLatency = v
		log.Infof("extra latency: %v", extraLatency)
	}

	// --- Hot-reload toggle via Unix signals ---
	// SIGUSR1 = enable reload on every request, SIGUSR2 = disable
	reloadCatalog := false
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGUSR1 {
				reloadCatalog = true
				log.Info("catalog reload enabled")
			} else {
				reloadCatalog = false
				log.Info("catalog reload disabled")
			}
		}
	}()

	port := "3550"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Infof("starting grpc server on :%s", port)
	if err := run(port, extraLatency, &reloadCatalog); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run(port string, extraLatency time.Duration, reloadFlag *bool) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	svc, err := catalog.NewProductCatalog(extraLatency, reloadFlag)
	if err != nil {
		return fmt.Errorf("failed to init product catalog: %w", err)
	}

	pb.RegisterProductCatalogServiceServer(srv, svc)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	log.Infof("listening on %s", listener.Addr().String())
	return srv.Serve(listener)
}

// initTracing wires OpenTelemetry to an OTLP gRPC collector.
// In the LGTM stack this points to Grafana Agent or OTel Collector → Tempo.
func initTracing() error {
	collectorAddr := os.Getenv("COLLECTOR_SERVICE_ADDR")
	if collectorAddr == "" {
		return fmt.Errorf("COLLECTOR_SERVICE_ADDR not set")
	}

	ctx := context.Background()

	conn, err := grpc.NewClient(
		collectorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to collector %s: %w", collectorAddr, err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return nil
}