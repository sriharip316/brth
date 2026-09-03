package breath

import (
	"testing"
	"time"

	"github.com/srihari/brth/internal/preset"
)

func TestEngineBoxProgression(t *testing.T) {
	boxPreset := preset.Preset{
		Name:    "box",
		Inhale:  1.0,
		HoldIn:  1.0,
		Exhale:  1.0,
		HoldOut: 1.0,
	}

	engine := NewEngine(boxPreset, 0, 2)
	if engine.CurrentPhase != PhaseInhale {
		t.Fatalf("expected initial phase Inhale, got %v", engine.CurrentPhase)
	}

	now := time.Now()
	// Step 1: Inhale -> HoldIn
	now = now.Add(1100 * time.Millisecond)
	trans := engine.Update(now)
	if !trans || engine.CurrentPhase != PhaseHoldIn {
		t.Fatalf("expected transition to HoldIn, got trans=%v, phase=%v", trans, engine.CurrentPhase)
	}

	// Step 2: HoldIn -> Exhale
	now = now.Add(1100 * time.Millisecond)
	trans = engine.Update(now)
	if !trans || engine.CurrentPhase != PhaseExhale {
		t.Fatalf("expected transition to Exhale, got trans=%v, phase=%v", trans, engine.CurrentPhase)
	}

	// Step 3: Exhale -> HoldOut
	now = now.Add(1100 * time.Millisecond)
	trans = engine.Update(now)
	if !trans || engine.CurrentPhase != PhaseHoldOut {
		t.Fatalf("expected transition to HoldOut, got trans=%v, phase=%v", trans, engine.CurrentPhase)
	}

	// Step 4: HoldOut -> Cycle 2 Inhale
	now = now.Add(1100 * time.Millisecond)
	trans = engine.Update(now)
	if !trans || engine.CurrentPhase != PhaseInhale || engine.CurrentCycle != 2 {
		t.Fatalf("expected transition to Cycle 2 Inhale, got trans=%v, phase=%v, cycle=%d", trans, engine.CurrentPhase, engine.CurrentCycle)
	}

	// Complete Cycle 2
	now = now.Add(1100 * time.Millisecond) // HoldIn
	engine.Update(now)
	now = now.Add(1100 * time.Millisecond) // Exhale
	engine.Update(now)
	now = now.Add(1100 * time.Millisecond) // HoldOut
	engine.Update(now)
	now = now.Add(1100 * time.Millisecond) // Should complete target 2 cycles
	engine.Update(now)

	if engine.CurrentPhase != PhaseDone {
		t.Fatalf("expected PhaseDone after 2 cycles, got phase=%v", engine.CurrentPhase)
	}
}

func TestEngineSkipZeroPhases(t *testing.T) {
	// Coherent has 0 hold phases
	coherent := preset.Preset{
		Name:    "coherent",
		Inhale:  1.0,
		HoldIn:  0.0,
		Exhale:  1.0,
		HoldOut: 0.0,
	}

	engine := NewEngine(coherent, 0, 1)
	if engine.CurrentPhase != PhaseInhale {
		t.Fatalf("expected Inhale phase, got %v", engine.CurrentPhase)
	}

	now := time.Now().Add(1100 * time.Millisecond)
	engine.Update(now)
	if engine.CurrentPhase != PhaseExhale {
		t.Fatalf("expected skip to Exhale (skipping 0s HoldIn), got %v", engine.CurrentPhase)
	}

	// End of Exhale with 0s HoldOut should complete cycle 1 and finish
	now = now.Add(1100 * time.Millisecond)
	engine.Update(now)
	if engine.CurrentPhase != PhaseDone {
		t.Fatalf("expected completion, got phase %v", engine.CurrentPhase)
	}
}

func TestEnginePause(t *testing.T) {
	p := preset.Preset{Name: "test", Inhale: 2.0, Exhale: 2.0}
	engine := NewEngine(p, 0, 0)

	engine.TogglePause()
	if !engine.Paused {
		t.Fatalf("expected engine to be paused")
	}

	now := time.Now().Add(5 * time.Second)
	trans := engine.Update(now)
	if trans || engine.PhaseElapsed != 0 {
		t.Fatalf("engine should not advance while paused")
	}

	engine.TogglePause()
	if engine.Paused {
		t.Fatalf("expected engine to be unpaused")
	}
}

func TestCurrentRadiusFactor(t *testing.T) {
	p := preset.Preset{Name: "test", Inhale: 2.0, HoldIn: 2.0, Exhale: 2.0, HoldOut: 2.0}
	engine := NewEngine(p, 0, 0)

	// At start of inhale, factor should be ~0.0
	f0 := engine.CurrentRadiusFactor()
	if f0 < 0.0 || f0 > 0.05 {
		t.Errorf("start inhale factor expected near 0, got %f", f0)
	}

	// Mid inhale
	engine.PhaseElapsed = 1 * time.Second
	fMid := engine.CurrentRadiusFactor()
	if fMid < 0.4 || fMid > 0.6 {
		t.Errorf("mid inhale factor expected near 0.5, got %f", fMid)
	}

	// EaseInOutSine bounds
	for _, v := range []float64{-0.5, 0.0, 0.25, 0.5, 0.75, 1.0, 1.5} {
		res := EaseInOutSine(v)
		if res < 0.0 || res > 1.0 {
			t.Errorf("EaseInOutSine(%f) = %f out of [0, 1]", v, res)
		}
	}
}
