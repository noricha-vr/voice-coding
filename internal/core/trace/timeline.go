package trace

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const EnableTimingLogsEnvVar = "VOICECODE_ENABLE_TIMING_LOGS"

var globalRunSeq uint64

type Logger interface {
	Printf(format string, v ...any)
}

// Timeline is a lightweight per-run timing logger.
// It is designed for "press hotkey -> result" end-to-end measurement.
type Timeline struct {
	runID   uint64
	name    string
	start   time.Time
	enabled bool
	logger  Logger
	clock   func() time.Time
}

func New(name string) *Timeline {
	return NewWithStart(name, time.Now())
}

func NewWithStart(name string, start time.Time) *Timeline {
	return NewWith(log.Default(), time.Now, timingEnabledFromEnv(), name, start)
}

func NewWith(logger Logger, clock func() time.Time, enabled bool, name string, start time.Time) *Timeline {
	if logger == nil {
		logger = log.Default()
	}
	if clock == nil {
		clock = time.Now
	}

	return &Timeline{
		runID:   atomic.AddUint64(&globalRunSeq, 1),
		name:    name,
		start:   start,
		enabled: enabled,
		logger:  logger,
		clock:   clock,
	}
}

func (t *Timeline) Enabled() bool {
	if t == nil {
		return false
	}
	return t.enabled
}

func (t *Timeline) RunID() uint64 {
	if t == nil {
		return 0
	}
	return t.runID
}

func (t *Timeline) StartTime() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.start
}

func (t *Timeline) SinceStart() time.Duration {
	if t == nil || !t.enabled {
		return 0
	}
	return t.clock().Sub(t.start)
}

func (t *Timeline) Eventf(format string, args ...any) {
	if t == nil || !t.enabled {
		return
	}

	now := t.clock()
	since := now.Sub(t.start)
	msg := fmt.Sprintf(format, args...)
	t.logger.Printf("[Timing run=%d +%s] %s", t.runID, fmtDur(since), msg)
}

// Step returns a closure that logs step duration when called.
// Usage:
//
//	done := tl.Step("recorder.Start")
//	err := rec.Start()
//	done(err)
func (t *Timeline) Step(name string) func(err error) {
	if t == nil || !t.enabled {
		return func(error) {}
	}

	stepStart := t.clock()
	return func(err error) {
		now := t.clock()
		stepDur := now.Sub(stepStart)
		since := now.Sub(t.start)

		status := "ok"
		if err != nil {
			status = "err"
		}

		if err != nil {
			t.logger.Printf("[Timing run=%d +%s] %s dur=%s %s: %v", t.runID, fmtDur(since), name, fmtDur(stepDur), status, err)
			return
		}
		t.logger.Printf("[Timing run=%d +%s] %s dur=%s %s", t.runID, fmtDur(since), name, fmtDur(stepDur), status)
	}
}

func (t *Timeline) Finishf(format string, args ...any) {
	if t == nil || !t.enabled {
		return
	}

	now := t.clock()
	since := now.Sub(t.start)
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	if msg != "" {
		msg = " " + msg
	}
	t.logger.Printf("[Timing run=%d +%s] FINISH%s", t.runID, fmtDur(since), msg)
}

func fmtDur(d time.Duration) string {
	return d.Truncate(time.Microsecond).String()
}

func timingEnabledFromEnv() bool {
	// Default: enabled (the user explicitly asked for E2E timing logs).
	v, ok := os.LookupEnv(EnableTimingLogsEnvVar)
	if !ok {
		return true
	}
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return true
	}
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return false
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	// Unknown value -> keep enabled (fail loud via log line).
	log.Printf("[Timing] invalid %s=%q (expected true/false). timing logs are enabled.", EnableTimingLogsEnvVar, v)
	return true
}

type timelineCtxKey struct{}

func WithTimeline(ctx context.Context, t *Timeline) context.Context {
	if t == nil {
		return ctx
	}
	return context.WithValue(ctx, timelineCtxKey{}, t)
}

func FromContext(ctx context.Context) *Timeline {
	if ctx == nil {
		return nil
	}
	t, _ := ctx.Value(timelineCtxKey{}).(*Timeline)
	return t
}
