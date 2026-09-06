package breath

import (
	"math"
	"time"

	"github.com/sriharip316/brth/internal/preset"
)

// Phase denotes the active breathing phase.
type Phase int

const (
	PhasePrep Phase = iota
	PhaseInhale
	PhaseHoldIn
	PhaseExhale
	PhaseHoldOut
	PhaseDone
)

func (p Phase) String() string {
	switch p {
	case PhasePrep:
		return "PREPARE"
	case PhaseInhale:
		return "INHALE"
	case PhaseHoldIn:
		return "HOLD"
	case PhaseExhale:
		return "EXHALE"
	case PhaseHoldOut:
		return "HOLD (EMPTY)"
	case PhaseDone:
		return "COMPLETE"
	default:
		return "BREATHE"
	}
}

// Guidance returns gentle instruction text for the phase.
func (p Phase) Guidance() string {
	switch p {
	case PhasePrep:
		return "Get comfortable. Empty your lungs."
	case PhaseInhale:
		return "Breathe in deeply through your nose..."
	case PhaseHoldIn:
		return "Hold your breath gently..."
	case PhaseExhale:
		return "Release slowly through your mouth..."
	case PhaseHoldOut:
		return "Rest empty and still..."
	case PhaseDone:
		return "Session complete. Return to your natural breath."
	default:
		return ""
	}
}

// Engine manages the breathing cycle state machine.
type Engine struct {
	Preset       preset.Preset
	PrepDuration time.Duration
	TargetCycles int

	CurrentPhase Phase
	CurrentCycle int
	PhaseElapsed time.Duration
	TotalElapsed time.Duration

	LastTime  time.Time
	Paused    bool
	RateScale float64 // 1.0 = normal, < 1.0 = faster, > 1.0 = slower
}

// NewEngine creates a breathing engine initialized with the given preset.
func NewEngine(p preset.Preset, prepSeconds int, targetCycles int) *Engine {
	if targetCycles <= 0 && p.Cycles > 0 {
		targetCycles = p.Cycles
	}

	// Guard against degenerate presets: advancePhase's zero-duration skip
	// loop would never terminate if every phase had zero length.
	if p.Inhale <= 0 && p.HoldIn <= 0 && p.Exhale <= 0 && p.HoldOut <= 0 {
		p.Inhale = 4.0
		p.Exhale = 4.0
	}

	e := &Engine{
		Preset:       p,
		PrepDuration: time.Duration(prepSeconds) * time.Second,
		TargetCycles: targetCycles,
		CurrentCycle: 1,
		RateScale:    1.0,
		LastTime:     time.Now(),
	}

	if prepSeconds > 0 {
		e.CurrentPhase = PhasePrep
	} else {
		e.CurrentPhase = e.firstValidPhase()
	}

	return e
}

// SetPreset updates the active preset and recomputes the target cycle count.
// A positive cyclesOverride (from the CLI) takes precedence over the preset's
// own cycle count.
func (e *Engine) SetPreset(p preset.Preset, cyclesOverride int) {
	e.Preset = p
	e.TargetCycles = cyclesOverride
	if e.TargetCycles <= 0 {
		e.TargetCycles = p.Cycles
	}
}

// Restart resets the breathing session to cycle 1.
func (e *Engine) Restart() {
	e.CurrentCycle = 1
	e.PhaseElapsed = 0
	e.TotalElapsed = 0
	e.Paused = false
	e.LastTime = time.Now()

	if e.PrepDuration > 0 {
		e.CurrentPhase = PhasePrep
	} else {
		e.CurrentPhase = e.firstValidPhase()
	}
}

// TogglePause toggles the pause state.
func (e *Engine) TogglePause() {
	e.Paused = !e.Paused
	if !e.Paused {
		e.LastTime = time.Now()
	}
}

// AdjustRate increases or decreases the pace (RateScale is clamped between 0.5 and 2.0).
func (e *Engine) AdjustRate(delta float64) {
	e.RateScale += delta
	if e.RateScale < 0.5 {
		e.RateScale = 0.5
	} else if e.RateScale > 2.0 {
		e.RateScale = 2.0
	}
}

// PhaseDuration returns the effective scaled duration for the given phase.
func (e *Engine) PhaseDuration(phase Phase) time.Duration {
	scale := e.RateScale
	if scale <= 0 {
		scale = 1.0
	}

	var sec float64
	switch phase {
	case PhasePrep:
		return e.PrepDuration
	case PhaseInhale:
		sec = e.Preset.Inhale * scale
	case PhaseHoldIn:
		sec = e.Preset.HoldIn * scale
	case PhaseExhale:
		sec = e.Preset.Exhale * scale
	case PhaseHoldOut:
		sec = e.Preset.HoldOut * scale
	case PhaseDone:
		return 0
	default:
		return 0
	}

	return time.Duration(sec * float64(time.Second))
}

// CurrentPhaseDuration returns the duration of the current phase.
func (e *Engine) CurrentPhaseDuration() time.Duration {
	return e.PhaseDuration(e.CurrentPhase)
}

// PhaseProgress returns normalized progress [0.0, 1.0] within the current phase.
func (e *Engine) PhaseProgress() float64 {
	dur := e.CurrentPhaseDuration()
	if dur <= 0 {
		return 1.0
	}
	progress := float64(e.PhaseElapsed) / float64(dur)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// PhaseTimeRemaining returns seconds left in the current phase.
func (e *Engine) PhaseTimeRemaining() float64 {
	dur := e.CurrentPhaseDuration()
	rem := dur - e.PhaseElapsed
	if rem < 0 {
		rem = 0
	}
	return rem.Seconds()
}

// Update advances the engine clock by delta time. Returns true if phase changed.
func (e *Engine) Update(now time.Time) bool {
	if e.Paused || e.CurrentPhase == PhaseDone {
		e.LastTime = now
		return false
	}

	delta := now.Sub(e.LastTime)
	e.LastTime = now
	if delta <= 0 {
		return false
	}

	e.PhaseElapsed += delta
	e.TotalElapsed += delta

	currentDur := e.CurrentPhaseDuration()
	if e.PhaseElapsed >= currentDur {
		// Advance to next phase
		overflow := e.PhaseElapsed - currentDur
		e.advancePhase()
		e.PhaseElapsed = overflow
		return true
	}

	return false
}

// firstValidPhase returns the first phase in the preset with non-zero duration.
func (e *Engine) firstValidPhase() Phase {
	if e.Preset.Inhale > 0 {
		return PhaseInhale
	}
	if e.Preset.HoldIn > 0 {
		return PhaseHoldIn
	}
	if e.Preset.Exhale > 0 {
		return PhaseExhale
	}
	if e.Preset.HoldOut > 0 {
		return PhaseHoldOut
	}
	return PhaseInhale
}

// advancePhase moves to the next non-zero duration phase or increments cycle.
// The cycle boundary is handled exactly once, in the check below; the skip
// loop can never reach HoldOut from a phase this check has not already
// covered, so it must not repeat the cycle counting.
func (e *Engine) advancePhase() {
	// If we completed HoldOut (or completed Exhale when HoldOut is 0), cycle has ended
	if (e.CurrentPhase == PhaseHoldOut) || (e.CurrentPhase == PhaseExhale && e.PhaseDuration(PhaseHoldOut) <= 0) {
		if e.TargetCycles > 0 && e.CurrentCycle >= e.TargetCycles {
			e.CurrentPhase = PhaseDone
			return
		}
		e.CurrentCycle++
	}

	// Skip any phases with 0 duration
	next := e.nextPhase(e.CurrentPhase)
	for next != PhaseDone && e.PhaseDuration(next) <= 0 {
		next = e.nextPhase(next)
	}

	e.CurrentPhase = next
}

func (e *Engine) nextPhase(p Phase) Phase {
	switch p {
	case PhasePrep:
		return PhaseInhale
	case PhaseInhale:
		return PhaseHoldIn
	case PhaseHoldIn:
		return PhaseExhale
	case PhaseExhale:
		return PhaseHoldOut
	case PhaseHoldOut:
		return PhaseInhale
	default:
		return PhaseInhale
	}
}

// EaseInOutSine smooths progress to emulate natural breathing mechanics.
func EaseInOutSine(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1.0 {
		return 1.0
	}
	return -(math.Cos(math.Pi*t) - 1.0) / 2.0
}

// CurrentRadiusFactor returns a value in [0.0, 1.0] representing visual expansion size.
// 0.0 = fully contracted (lungs empty), 1.0 = fully expanded (lungs full).
func (e *Engine) CurrentRadiusFactor() float64 {
	p := e.PhaseProgress()

	switch e.CurrentPhase {
	case PhasePrep:
		// Gentle rest pulse before start
		return 0.15 + 0.05*math.Sin(p*2*math.Pi)
	case PhaseInhale:
		// Expands from 0.0 to 1.0
		return EaseInOutSine(p)
	case PhaseHoldIn:
		// Stays near 1.0 with a gentle micro-pulse
		return 0.98 + 0.02*math.Sin(p*math.Pi)
	case PhaseExhale:
		// Contracts from 1.0 to 0.0
		return 1.0 - EaseInOutSine(p)
	case PhaseHoldOut:
		// Stays near 0.0 with a gentle resting pulse
		return 0.02 * (1.0 + math.Sin(p*math.Pi))
	case PhaseDone:
		return 0.5
	default:
		return 0.0
	}
}
