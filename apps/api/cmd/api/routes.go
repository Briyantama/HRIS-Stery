package main

import (
	"fmt"
	"net/http"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/Briyantama/stery-kit/go/health"
	"github.com/Briyantama/stery-kit/go/objstore"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Briyantama/HRIS-Stery/internal/config"
	"github.com/Briyantama/HRIS-Stery/services/attendance"
	"github.com/Briyantama/HRIS-Stery/services/audit"
	"github.com/Briyantama/HRIS-Stery/services/auth"
	"github.com/Briyantama/HRIS-Stery/services/document"
	"github.com/Briyantama/HRIS-Stery/services/employee"
	"github.com/Briyantama/HRIS-Stery/services/leave"
	"github.com/Briyantama/HRIS-Stery/services/notification"
)

type registrar = interface{ Register(*http.ServeMux) }

// build constructs every service: their HTTP handlers and their stream consumers keyed by group.
func build(cfg config.Config, pool *pgxpool.Pool, rdb *redis.Client, store objstore.Store) ([]registrar, map[string]map[string]events.Handler, error) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM(cfg.AuthPublicKeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("auth public key: %w", err)
	}
	verifier := authn.NewRSAVerifier(pub, nil)
	publisher := events.NewPublisher(rdb, eventStream)

	employeeDeps := employee.Deps{Pool: pool, Events: publisher, Verifier: verifier}
	leaveDeps := leave.Deps{Pool: pool, Events: publisher, Verifier: verifier}
	auditDeps := audit.Deps{Pool: pool, Verifier: verifier}
	notificationDeps := notification.Deps{Pool: pool, Verifier: verifier}

	var handlers []registrar
	for _, b := range []func() (registrar, error){
		func() (registrar, error) {
			return auth.Build(auth.Deps{Config: cfg, Pool: pool, Redis: rdb, Events: publisher, Verifier: verifier})
		},
		func() (registrar, error) { return employee.Build(employeeDeps) },
		func() (registrar, error) { return leave.Build(leaveDeps) },
		func() (registrar, error) {
			return attendance.Build(attendance.Deps{Pool: pool, Events: publisher, Verifier: verifier})
		},
		func() (registrar, error) { return audit.Build(auditDeps) },
		func() (registrar, error) {
			return document.Build(document.Deps{Config: cfg, Pool: pool, Verifier: verifier, Store: store})
		},
		func() (registrar, error) { return notification.Build(notificationDeps) },
	} {
		h, err := b()
		if err != nil {
			return nil, nil, err
		}
		handlers = append(handlers, h)
	}

	consumers := map[string]map[string]events.Handler{
		employee.ConsumerGroup:     employee.Consumers(employeeDeps),
		leave.ConsumerGroup:        leave.Consumers(leaveDeps),
		audit.ConsumerGroup:        audit.Consumers(auditDeps),
		notification.ConsumerGroup: notification.Consumers(notificationDeps),
	}
	return handlers, consumers, nil
}

// newMux registers liveness and every service's routes.
func newMux(handlers ...registrar) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", health.Handler())
	for _, h := range handlers {
		h.Register(mux)
	}
	return mux
}
