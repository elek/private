// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package process

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	hw "github.com/jtolds/monkit-hw/v2"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/environment"
	"go.uber.org/zap"

	"storj.io/common/identity"
	"storj.io/common/telemetry"
	jaeger "storj.io/monkit-jaeger"
	"storj.io/private/version"
)

var (
	tracingSamplingRate = flag.Float64("tracing.sample", 1, "how frequently to send up telemetry")
	tracingAgent        = flag.String("tracing.agent-addr", flagDefault("127.0.0.1:5775", ""), "address for jaeger agent")
	tracingApp          = flag.String("tracing.app", filepath.Base(os.Args[0]), "application name for tracing identification")
	tracingAppSuffix    = flag.String("tracing.app-suffix", flagDefault("-dev", "-release"), "application suffix")
	tracingBuffer       = flag.Int("tracing.buffer-size", 1000, "buffer size for collector queue")
)

const (
	instanceIDKey = "instanceID"
	hostnameKey   = "hostname"
)

// InitTracing initializes distributed tracing with an instance ID.
func InitTracing(ctx context.Context, log *zap.Logger, r *monkit.Registry, instanceID string) error {
	return run(ctx, log, r, instanceID, []jaeger.Tag{})
}

// InitTracingWithCertPath initializes distributed tracing with certificate path.
func InitTracingWithCertPath(ctx context.Context, log *zap.Logger, r *monkit.Registry, certPath string) error {
	return run(ctx, log, r, nodeIDFromCertPath(ctx, log, certPath), []jaeger.Tag{})
}

// InitTracingWithHostname initializes distributed tracing with nodeID and hostname.
func InitTracingWithHostname(ctx context.Context, log *zap.Logger, r *monkit.Registry, certPath string) error {
	var processInfo []jaeger.Tag
	hostname, err := os.Hostname()
	if err != nil {
		log.Sugar().Errorf("Could not read hostname for tracing setup: %v", err)
	} else {
		processInfo = append(processInfo, jaeger.Tag{
			Key:   hostnameKey,
			Value: hostname,
		})
	}

	return run(ctx, log, r, nodeIDFromCertPath(ctx, log, certPath), processInfo)
}

func run(ctx context.Context, log *zap.Logger, r *monkit.Registry, instanceID string, processInfo []jaeger.Tag) (err error) {
	if r == nil {
		r = monkit.Default
	}
	environment.Register(r)
	hw.Register(r)
	r.ScopeNamed("env").Chain(monkit.StatSourceFunc(version.Build.Stats))

	log = log.Named("tracing")
	if *tracingAgent == "" || *tracingSamplingRate == 0 {
		log.Info("disabled")
		return nil
	}

	if len(instanceID) == 0 {
		instanceID = telemetry.DefaultInstanceID()
	}
	processInfo = append(processInfo, jaeger.Tag{
		Key:   instanceIDKey,
		Value: instanceID,
	})

	processName := *tracingApp + *tracingAppSuffix
	if len(processName) > maxInstanceLength {
		processName = processName[:maxInstanceLength]
	}
	collector, err := jaeger.NewUDPCollector(*tracingAgent, *tracingBuffer, processName, processInfo)
	if err != nil {
		return err
	}
	jaeger.RegisterJaeger(r, collector, jaeger.Options{Fraction: *tracingSamplingRate})
	return nil
}

func nodeIDFromCertPath(ctx context.Context, log *zap.Logger, certPath string) string {
	nodeID, err := identity.NodeIDFromCertPath(certPath)
	if err != nil {
		log.Sugar().Errorf("Could not read identity for tracing setup: %v", err)
		return ""
	}

	return nodeID.String()
}
