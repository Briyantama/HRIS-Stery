package sharednats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// AckMessage acknowledges a core NATS message and logs failures.
func AckMessage(logger *zap.Logger, msg *nats.Msg) {
	if err := msg.Ack(); err != nil {
		logger.Error("nats ack failed", zap.Error(err), zap.String("subject", msg.Subject))
	}
}

// NakMessage negatively acknowledges a core NATS message and logs failures.
func NakMessage(logger *zap.Logger, msg *nats.Msg) {
	if err := msg.Nak(); err != nil {
		logger.Error("nats nak failed", zap.Error(err), zap.String("subject", msg.Subject))
	}
}

// AckJetStream acknowledges a JetStream message and logs failures.
func AckJetStream(logger *zap.Logger, msg jetstream.Msg) {
	if err := msg.Ack(); err != nil {
		subject := ""
		if msg != nil {
			subject = msg.Subject()
		}
		logger.Error("jetstream ack failed", zap.Error(err), zap.String("subject", subject))
	}
}

// NakJetStream negatively acknowledges a JetStream message and logs failures.
func NakJetStream(logger *zap.Logger, msg jetstream.Msg) {
	if err := msg.Nak(); err != nil {
		subject := ""
		if msg != nil {
			subject = msg.Subject()
		}
		logger.Error("jetstream nak failed", zap.Error(err), zap.String("subject", subject))
	}
}
