package kbdscoring

import "../kbdlayout"

// Provides a score for a given keyboard layout
// The Init method will be called with the mapping to be used
// with the layouts to be scored. For performance reasons
// it is good practise to initialize the data structures
// so that they rely on the mapping indices instead of
// real characters.
type ScoringFunction interface {
	Init(mapping *kbdlayout.KeyboardMapping)

	// Return a score for the given layout. Higher score
	// translates to better keyboard layout.
	// Integers are used for performance reasons.
	CalculateScore(layout *kbdlayout.KeyboardLayout) uint64

	// This function is called when the keyboard score
	// will be presented to user.
	NormalizeScore(score uint64) float64
}
