package benchmarks

import (
	"context"
	"testing"

	"github.com/google/uuid"

	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

func BenchmarkBrokerPublish(b *testing.B) {
	broker := ssebroker.NewBroker(5000)
	defer broker.Close()

	ctx := context.Background()
	scope := ssebroker.Scope{Type: "tournament", ID: uuid.Nil}

	payload := []byte(`{"event":"test"}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.Publish(ctx, scope, "MockEvent", payload); err != nil {
			b.Fatalf("publish: %v", err)
		}
	}
}
