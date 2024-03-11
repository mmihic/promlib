// Package querylog logs prometheus queries.
package querylog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// A LoggedQuery is an in-flight query.
type LoggedQuery interface {
	// QueryComplete is called when the query is complete.
	QueryComplete(result any)

	// QueryFailed is called if the query fails.
	QueryFailed(err error)
}

// A Logger logs queries.
type Logger interface {
	// BeginQuery starts a new query.
	BeginQuery(queryType string, fields ...zap.Field) LoggedQuery
}

// NewNop returns a Logger that does nothing.
func NewNop() Logger {
	return nopLogger{}
}

type nopLoggedQuery struct{}

func (q nopLoggedQuery) QueryComplete(_ any) {}
func (q nopLoggedQuery) QueryFailed(_ error) {}

type nopLogger struct{}

func (l nopLogger) BeginQuery(_ string, _ ...zap.Field) LoggedQuery {
	return nopLoggedQuery{}
}

// New returns a new Logger.
func New(log *zap.Logger, logResponses bool) Logger {
	return NewWithClock(log, logResponses, clockwork.NewRealClock())
}

// NewWithClock returns a new Logger using a custom Clock.
func NewWithClock(log *zap.Logger, logResponses bool, clock clockwork.Clock) Logger {
	return &logger{
		log:         log,
		logResponse: logResponses,
		clock:       clock,
	}
}

type logger struct {
	nextID      atomic.Uint64
	log         *zap.Logger
	logResponse bool
	clock       clockwork.Clock
}

func (log *logger) BeginQuery(queryType string, fields ...zap.Field) LoggedQuery {
	nextID := log.nextID.Add(1)
	startTime := log.clock.Now()

	if ce := log.log.Check(zap.InfoLevel, queryType); ce != nil {
		ce.Write(append([]zap.Field{
			zap.Uint64("query_id", nextID),
		}, fields...)...)

		return loggedQuery{
			id:        nextID,
			startTime: startTime,
			queryType: queryType,
			fields:    fields,
			logger:    log,
		}
	}

	return nopLoggedQuery{}
}

type loggedQuery struct {
	id        uint64
	startTime time.Time
	queryType string
	fields    []zap.Field
	logger    *logger
}

func (q loggedQuery) QueryComplete(result any) {
	if ce := q.logger.log.Check(zap.InfoLevel, q.queryType); ce != nil {
		fields := append([]zap.Field{
			zap.Uint64("query_id", q.id),
			zap.Duration("elapsed_time", q.logger.clock.Now().Sub(q.startTime)),
		}, q.fields...)

		if q.logger.logResponse {
			b, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fields = append(fields,
					zap.String("result", fmt.Sprintf("err marshalling: %s", err)))
			} else {
				fields = append(fields, zap.Any("result", json.RawMessage(b)))
			}
		}

		ce.Write(fields...)
	}
}

func (q loggedQuery) QueryFailed(err error) {
	if ce := q.logger.log.Check(zap.InfoLevel, q.queryType); ce != nil {
		ce.Write(zap.Uint64("query_id", q.id))
		ce.Write(q.fields...)
		ce.Write(zap.Error(err))
	}
}
