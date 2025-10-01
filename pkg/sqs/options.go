package sqs

// config holds the complete configuration for the SQS client with adaptive polling capabilities.
type config struct {
	// VisibilityTimeout defines how long messages remain invisible after being received (in seconds).
	VisibilityTimeout int
	// AdaptivePolling contains all settings related to the Arrakis adaptive polling algorithm.
	AdaptivePolling adaptivePolling

	arrakis arrakis
}

// adaptivePolling contains configuration parameters for the adaptive polling algorithm.
// This algorithm dynamically adjusts polling intervals based on message volume using EWMA (Exponentially Weighted Moving Average).
type adaptivePolling struct {
	// EnableAdaptivePolling controls whether the Arrakis adaptive polling is active.
	EnableAdaptivePolling bool

	// IdleWaitTimeSeconds defines the wait time when no messages are being processed (idle state).
	IdleWaitTimeSeconds int
	// VisibilityTimeout for adaptive polling context (may differ from main config).
	VisibilityTimeout int

	// LowVolumeWaitTimeSeconds defines wait time when message volume is low (< 2 messages/poll).
	LowVolumeWaitTimeSeconds int
	// MediumVolumeWaitTimeSeconds defines wait time for medium volume (2-5 messages/poll).
	MediumVolumeWaitTimeSeconds int
	// HighVolumeWaitTimeSeconds defines wait time for high volume (5-10 messages/poll).
	HighVolumeWaitTimeSeconds int
	// VeryHighVolumeWaitTimeSeconds defines wait time for very high volume (>10 messages/poll).
	VeryHighVolumeWaitTimeSeconds int

	// EwmaAlpha is the smoothing factor for EWMA calculation (0.0 to 1.0).
	// Higher values give more weight to recent observations.
	EwmaAlpha float64
	// DropDetectionThreshold defines how many consecutive low-volume cycles trigger EWMA reset.
	DropDetectionThreshold int
}

// Option is a function type for configuring the SQS client with the functional options pattern.
type Option func(*config)

// WithAdaptivePolling configures all adaptive polling parameters in a single call.
// This is a convenience function for setting up the complete Arrakis adaptive polling configuration.
//
// Parameters:
//   - idleWaitTimeSeconds: Wait time when no messages are present (idle state)
//   - visibilityTimeout: Message visibility timeout for adaptive polling
//   - lowVolumeWaitTimeSeconds: Wait time for low message volume scenarios
//   - mediumVolumeWaitTimeSeconds: Wait time for medium message volume scenarios
//   - highVolumeWaitTimeSeconds: Wait time for high message volume scenarios
//   - veryHighVolumeWaitTimeSeconds: Wait time for very high message volume scenarios
//   - ewmaAlpha: EWMA smoothing factor (0.0-1.0, typically 0.1-0.5)
//   - dropDetectionThreshold: Number of low-volume cycles before EWMA reset
//
// Example:
//
//	option := WithAdaptivePolling(20, 30, 15, 10, 5, 1, 0.3, 10)
func WithAdaptivePolling(idleWaitTimeSeconds, visibilityTimeout, lowVolumeWaitTimeSeconds, mediumVolumeWaitTimeSeconds, highVolumeWaitTimeSeconds, veryHighVolumeWaitTimeSeconds int, ewmaAlpha float64, dropDetectionThreshold int) Option {
	return func(c *config) {
		c.AdaptivePolling.IdleWaitTimeSeconds = idleWaitTimeSeconds
		c.AdaptivePolling.VisibilityTimeout = visibilityTimeout
		c.AdaptivePolling.LowVolumeWaitTimeSeconds = lowVolumeWaitTimeSeconds
		c.AdaptivePolling.MediumVolumeWaitTimeSeconds = mediumVolumeWaitTimeSeconds
		c.AdaptivePolling.HighVolumeWaitTimeSeconds = highVolumeWaitTimeSeconds
		c.AdaptivePolling.VeryHighVolumeWaitTimeSeconds = veryHighVolumeWaitTimeSeconds
		c.AdaptivePolling.EwmaAlpha = ewmaAlpha
		c.AdaptivePolling.DropDetectionThreshold = dropDetectionThreshold
	}
}

// WithIdleWaitTimeSeconds sets the wait time when the queue is idle (no messages present).
// This is the longest wait time used when the system detects no message activity.
//
// Parameters:
//   - idleWaitTimeSeconds: Wait time in seconds (recommended: 15-20 seconds)
func WithIdleWaitTimeSeconds(idleWaitTimeSeconds int) Option {
	return func(c *config) {
		c.AdaptivePolling.IdleWaitTimeSeconds = idleWaitTimeSeconds
	}
}

// WithVisibilityTimeout sets the message visibility timeout for adaptive polling.
// This determines how long messages remain invisible after being received.
//
// Parameters:
//   - visibilityTimeout: Timeout in seconds (recommended: 30-300 seconds)
func WithVisibilityTimeout(visibilityTimeout int) Option {
	return func(c *config) {
		c.AdaptivePolling.VisibilityTimeout = visibilityTimeout
	}
}

// WithLowVolumeWaitTimeSeconds sets the wait time for low message volume scenarios.
// Used when the EWMA average is below the low volume threshold (< 2 messages).
//
// Parameters:
//   - lowVolumeWaitTimeSeconds: Wait time in seconds (recommended: 10-15 seconds)
func WithLowVolumeWaitTimeSeconds(lowVolumeWaitTimeSeconds int) Option {
	return func(c *config) {
		c.AdaptivePolling.LowVolumeWaitTimeSeconds = lowVolumeWaitTimeSeconds
	}
}

// WithMediumVolumeWaitTimeSeconds sets the wait time for medium message volume scenarios.
// Used when the EWMA average is between low and medium thresholds (2-5 messages).
//
// Parameters:
//   - mediumVolumeWaitTimeSeconds: Wait time in seconds (recommended: 5-10 seconds)
func WithMediumVolumeWaitTimeSeconds(mediumVolumeWaitTimeSeconds int) Option {
	return func(c *config) {
		c.AdaptivePolling.MediumVolumeWaitTimeSeconds = mediumVolumeWaitTimeSeconds
	}
}

// WithHighVolumeWaitTimeSeconds sets the wait time for high message volume scenarios.
// Used when the EWMA average is between medium and high thresholds (5-10 messages).
//
// Parameters:
//   - highVolumeWaitTimeSeconds: Wait time in seconds (recommended: 2-5 seconds)
func WithHighVolumeWaitTimeSeconds(highVolumeWaitTimeSeconds int) Option {
	return func(c *config) {
		c.AdaptivePolling.HighVolumeWaitTimeSeconds = highVolumeWaitTimeSeconds
	}
}

// WithVeryHighVolumeWaitTimeSeconds sets the wait time for very high message volume scenarios.
// Used when the EWMA average exceeds the high volume threshold (>10 messages).
//
// Parameters:
//   - veryHighVolumeWaitTimeSeconds: Wait time in seconds (recommended: 1-2 seconds)
func WithVeryHighVolumeWaitTimeSeconds(veryHighVolumeWaitTimeSeconds int) Option {
	return func(c *config) {
		c.AdaptivePolling.VeryHighVolumeWaitTimeSeconds = veryHighVolumeWaitTimeSeconds
	}
}

// WithEwmaAlpha sets the smoothing factor for the EWMA (Exponentially Weighted Moving Average) calculation.
// This parameter controls how much weight is given to recent observations versus historical data.
//
// Parameters:
//   - ewmaAlpha: Smoothing factor between 0.0 and 1.0
//   - Lower values (0.1-0.3): More stable, slower to adapt to changes
//   - Higher values (0.4-0.7): More responsive, faster adaptation to changes
//   - Recommended: 0.2-0.4 for most use cases
func WithEwmaAlpha(ewmaAlpha float64) Option {
	return func(c *config) {
		c.AdaptivePolling.EwmaAlpha = ewmaAlpha
	}
}

// WithDropDetectionThreshold sets the threshold for detecting message volume drops.
// When the system experiences this many consecutive low-volume cycles, it resets the EWMA
// to adapt more quickly to reduced message volumes.
//
// Parameters:
//   - dropDetectionThreshold: Number of consecutive low-volume cycles (recommended: 5-15)
func WithDropDetectionThreshold(dropDetectionThreshold int) Option {
	return func(c *config) {
		c.AdaptivePolling.DropDetectionThreshold = dropDetectionThreshold
	}
}

// setDefaults initializes the configuration with sensible default values.
// This function ensures that all adaptive polling parameters have valid values
// even if they weren't explicitly configured by the user.
//
// Default values are based on AWS SQS best practices and performance testing:
// - Idle wait time: 20 seconds (maximum SQS long polling)
// - Visibility timeout: 30 seconds (sufficient for most processing tasks)
// - Volume-based wait times: Decrease as message volume increases
// - EWMA alpha: 0.3 (balanced responsiveness and stability)
// - Drop detection: 10 cycles (reasonable adaptation to volume changes)
func setDefaults(c *config) {
	// Set main VisibilityTimeout if not already set
	if c.VisibilityTimeout == 0 {
		c.VisibilityTimeout = _defaultVisibilityTimeout
	}

	if c.AdaptivePolling.IdleWaitTimeSeconds == 0 {
		c.AdaptivePolling.IdleWaitTimeSeconds = _defaultIdleWaitTimeSeconds
	}

	if c.AdaptivePolling.VisibilityTimeout == 0 {
		c.AdaptivePolling.VisibilityTimeout = _defaultVisibilityTimeout
	}

	if c.AdaptivePolling.LowVolumeWaitTimeSeconds == 0 {
		c.AdaptivePolling.LowVolumeWaitTimeSeconds = _defaultLowVolumeWaitTimeSeconds
	}

	if c.AdaptivePolling.MediumVolumeWaitTimeSeconds == 0 {
		c.AdaptivePolling.MediumVolumeWaitTimeSeconds = _defaultMediumVolumeWaitTimeSeconds
	}

	if c.AdaptivePolling.HighVolumeWaitTimeSeconds == 0 {
		c.AdaptivePolling.HighVolumeWaitTimeSeconds = _defaultHighVolumeWaitTimeSeconds
	}

	if c.AdaptivePolling.VeryHighVolumeWaitTimeSeconds == 0 {
		c.AdaptivePolling.VeryHighVolumeWaitTimeSeconds = _defaultVeryHighVolumeWaitTimeSeconds
	}

	if c.AdaptivePolling.EwmaAlpha == 0 {
		c.AdaptivePolling.EwmaAlpha = _defaultEwmaAlpha
	}

	if c.AdaptivePolling.DropDetectionThreshold == 0 {
		c.AdaptivePolling.DropDetectionThreshold = _defaultDropDetectionThreshold
	}
}
