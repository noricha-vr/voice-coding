package trace

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func TestTimelineStepAndEventLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	now := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	tl := NewWith(logger, clock, true, "test", now)
	tl.Eventf("hotkey pressed")

	now = now.Add(150 * time.Millisecond)
	done := tl.Step("recorder.Start")
	now = now.Add(20 * time.Millisecond)
	done(nil)

	tl.Finishf("ok")

	out := buf.String()
	if !strings.Contains(out, "hotkey pressed") {
		t.Fatalf("expected event log, got: %q", out)
	}
	if !strings.Contains(out, "recorder.Start dur=20ms ok") {
		t.Fatalf("expected step duration log, got: %q", out)
	}
	if !strings.Contains(out, "FINISH ok") {
		t.Fatalf("expected finish log, got: %q", out)
	}
}

func TestTimingDisabledIsNoop(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	tl := NewWith(logger, clock, false, "test", now)
	tl.Eventf("nope")
	done := tl.Step("nope.step")
	done(nil)
	tl.Finishf("nope.finish")

	if buf.Len() != 0 {
		t.Fatalf("expected no output when disabled, got: %q", buf.String())
	}
}
