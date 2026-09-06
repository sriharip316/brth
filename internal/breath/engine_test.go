package breath

import (
	"testing"
	"time"

	"github.com/sriharip316/brth/internal/preset"
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

func TestEngineNoHoldOutExactCycleCount(t *testing.T) {
	// Presets without a hold-out (4-7-8, coherent, triangle, ...) must count
	// each breath exactly once: the cycle boundary fires when leaving Exhale.
	p := preset.Preset{Name: "fast", Inhale: 1.0, Exhale: 1.0, HoldOut: 0.0}

	engine := NewEngine(p, 0, 3)
	now := time.Now()
	for i := 0; i < 100 && engine.CurrentPhase != PhaseDone; i++ {
		now = now.Add(600 * time.Millisecond)
		engine.Update(now)
	}

	if engine.CurrentPhase != PhaseDone {
		t.Fatalf("expected session to complete, got phase %v", engine.CurrentPhase)
	}
	if engine.CurrentCycle != 3 {
		t.Fatalf("expected exactly 3 cycles, engine counted %d", engine.CurrentCycle)
	}
}

func TestSetPresetRecomputesTargetCycles(t *testing.T) {
	continuous := preset.Preset{Name: "cont", Inhale: 1.0, Exhale: 1.0}
	limited := preset.Preset{Name: "limited", Inhale: 1.0, Exhale: 1.0, Cycles: 4}

	engine := NewEngine(continuous, 0, 0)
	if engine.TargetCycles != 0 {
		t.Fatalf("expected continuous session, got target cycles %d", engine.TargetCycles)
	}

	engine.SetPreset(limited, 0)
	if engine.TargetCycles != 4 {
		t.Fatalf("expected 4 target cycles after switching, got %d", engine.TargetCycles)
	}

	// Switching back to a continuous preset must not inherit the old limit.
	engine.SetPreset(continuous, 0)
	if engine.TargetCycles != 0 {
		t.Fatalf("expected continuous after switching back, got %d", engine.TargetCycles)
	}

	// An explicit CLI override outranks the preset's own cycle count.
	engine.SetPreset(continuous, 7)
	if engine.TargetCycles != 7 {
		t.Fatalf("expected override of 7 cycles, got %d", engine.TargetCycles)
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
