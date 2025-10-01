package sqs

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// arrakis contains the state and algorithm implementation for adaptive SQS polling.
// It uses EWMA (Exponentially Weighted Moving Average) to track message volume patterns
// and automatically adjusts polling intervals to optimize API usage and responsiveness.
//
// The algorithm works by:
// 1. Tracking message counts from each polling operation
// 2. Calculating EWMA to smooth out volume fluctuations
// 3. Classifying current volume into categories (idle, low, medium, high, very high)
// 4. Selecting appropriate wait times based on volume classification
// 5. Implementing decay mechanisms for idle periods
// 6. Detecting volume drops and resetting when appropriate
type arrakis struct {
	// mu protects concurrent access to the EWMA calculation and state updates
	mu sync.RWMutex

	// Atomic counters for thread-safe message tracking
	messageCount    int64 // Current message count from last polling operation
	lastUpdate      int64 // Unix timestamp of last update
	messageSum      int64 // Cumulative sum of messages (used for statistics)
	messageCounts   int64 // Number of polling operations performed
	windowStartTime int64 // Start time of current measurement window

	// EWMA calculation state (protected by mutex)
	average          float64   // Current EWMA average of message volume
	lowVolumeCycle   int       // Counter of consecutive low-volume cycles
	lastReceiveEmpty time.Time // Timestamp of last empty response
	lastReset        time.Time // Timestamp of last EWMA reset

	// Algorithm configuration (set during initialization)
	dropDetectionThreshold   int64   // Threshold for detecting volume drops
	ewmaAlpha                float64 // EWMA smoothing factor
	consecutiveEmptyMessages int64   // Counter of consecutive empty responses
}

// updateMessageCount processes a new message count observation and updates the EWMA algorithm state.
// This method is called after each polling operation to incorporate the new message count
// into the adaptive polling algorithm's calculations.
//
// The function performs several key operations:
// 1. Updates atomic counters for thread-safe statistics tracking
// 2. Recalculates the EWMA average with the new observation
// 3. Tracks low-volume cycles for drop detection
// 4. Triggers EWMA reset if volume drop conditions are met
//
// Parameters:
//   - messageCount: Number of messages received in the current polling operation
func (s *SQS) updateMessageCount(messageCount int) {
	now := time.Now().Unix()

	// Update atomic counters for thread-safe access
	atomic.StoreInt64(&s.config.arrakis.messageCount, int64(messageCount))
	atomic.StoreInt64(&s.config.arrakis.lastUpdate, now)
	atomic.StoreInt64(&s.config.arrakis.messageSum, int64(messageCount))
	atomic.StoreInt64(&s.config.arrakis.messageCounts, 1)

	// Protect EWMA calculation with mutex
	defer s.config.arrakis.mu.Unlock()
	s.config.arrakis.mu.Lock()

	// Update EWMA with new observation
	s.config.arrakis.average = s.calculateAverage(messageCount)

	// Track low-volume cycles for drop detection
	if messageCount < _lowVolumeMessageThreshold {
		s.config.arrakis.lowVolumeCycle++
		// Check if we should reset EWMA due to sustained low volume
		if s.shouldResetEWMA() {
			s.resetEWMA()
		}
	} else {
		// Reset low-volume cycle counter on higher volume
		s.config.arrakis.lowVolumeCycle = 0
	}
}

// calculateAverage computes the new EWMA (Exponentially Weighted Moving Average) value
// incorporating the latest message count observation. The algorithm includes spike protection
// to prevent sudden large increases from dramatically skewing the average.
//
// The EWMA formula used is: new_average = α * current_value + (1-α) * old_average
// where α (alpha) is the smoothing factor controlling responsiveness vs stability.
//
// Spike protection limits the maximum change in a single update to prevent outliers
// from causing dramatic shifts in the polling behavior.
//
// Parameters:
//   - messageCount: The current message count observation
//
// Returns:
//   - float64: The updated EWMA average
func (s *SQS) calculateAverage(messageCount int) float64 {
	count := float64(messageCount)

	// Apply spike protection if we have an existing average
	if s.config.arrakis.average > 0 {
		delta := count - s.config.arrakis.average
		maxDelta := s.config.arrakis.average * 2 // Allow maximum 200% increase per update
		if delta > maxDelta {
			count = s.config.arrakis.average + maxDelta
		}
	}

	// Calculate EWMA: α * current + (1-α) * previous
	s.config.arrakis.average = s.config.arrakis.ewmaAlpha*count + (1.0-s.config.arrakis.ewmaAlpha)*s.config.arrakis.average

	return s.config.arrakis.average
}

// shouldResetEWMA determines whether the EWMA should be reset due to sustained low volume.
// This mechanism helps the algorithm adapt quickly when message volume drops significantly,
// preventing the EWMA from being "stuck" at high values during low-traffic periods.
//
// Reset conditions (all must be true):
// 1. Sufficient low-volume cycles have occurred (prevents premature resets)
// 2. Current EWMA average is below the reset threshold (confirms sustained low volume)
// 3. Minimum time has passed since last reset (prevents reset thrashing)
//
// Returns:
//   - bool: true if EWMA should be reset, false otherwise
func (s *SQS) shouldResetEWMA() bool {
	ewmaConfig := &s.config.arrakis

	hasEnoughLowVolumeCycles := ewmaConfig.lowVolumeCycle >= s.config.AdaptivePolling.DropDetectionThreshold
	isAverageBelowThreshold := ewmaConfig.average < _ewmaResetAverageThreshold
	hasMinimumTimePassed := time.Since(ewmaConfig.lastReset) > _minResetIntervalMinutes*time.Minute

	return hasEnoughLowVolumeCycles && isAverageBelowThreshold && hasMinimumTimePassed
}

// resetEWMA resets the EWMA algorithm state to handle volume drops.
// This allows the algorithm to quickly adapt to new, lower volume patterns
// instead of gradually decreasing from previously high averages.
//
// The reset operation:
// 1. Sets the EWMA average to zero (fresh start)
// 2. Resets the low-volume cycle counter
// 3. Records the reset timestamp to prevent frequent resets
func (s *SQS) resetEWMA() {
	s.config.arrakis.average = 0
	s.config.arrakis.lowVolumeCycle = 0
	s.config.arrakis.lastReset = time.Now()
}

// handleReceiveResponse processes the result of a ReceiveMessage operation and updates
// the adaptive polling algorithm state accordingly. This method distinguishes between
// empty and non-empty responses, applying different logic for each case.
//
// For empty responses: Tracks consecutive empty messages and applies EWMA decay
// For non-empty responses: Updates message counts and recalculates EWMA average
//
// Parameters:
//   - res: The SQS ReceiveMessage response to process
func (s *SQS) handleReceiveResponse(res *sqs.ReceiveMessageOutput) {
	if len(res.Messages) == 0 {
		s.handleEmptyResponse()
	} else if s.IsArrakisEnabled() {
		s.handleNonEmptyResponse(len(res.Messages))
	}
}

// handleEmptyResponse processes a polling operation that returned no messages.
// This method implements idle period handling by:
// 1. Tracking consecutive empty responses
// 2. Applying EWMA decay when appropriate to gradually reduce the average
// 3. Recording timestamps for idle period analysis
//
// Empty responses are important signals that help the algorithm detect when
// message volume has decreased and adjust polling intervals accordingly.
func (s *SQS) handleEmptyResponse() {
	if s.IsArrakisEnabled() {
		s.incrementConsecutiveEmptyMessages()

		// Apply EWMA decay if we've had enough consecutive empty responses
		if s.shouldDecayEWMA() {
			s.decayEWMA()
		}
	}
	s.config.arrakis.lastReceiveEmpty = time.Now()
}

// handleNonEmptyResponse processes a polling operation that returned messages.
// This method updates the algorithm state with the new message count and
// resets empty message tracking since we received actual messages.
//
// Parameters:
//   - messageCount: Number of messages received in this polling operation
func (s *SQS) handleNonEmptyResponse(messageCount int) {
	s.resetConsecutiveEmptyMessages()
	s.updateMessageCount(messageCount)
}

// incrementConsecutiveEmptyMessages safely increments the counter of consecutive
// empty polling responses. This counter is used to determine when EWMA decay
// should be applied during idle periods.
//
// Thread-safe operation using mutex protection.
func (s *SQS) incrementConsecutiveEmptyMessages() {
	s.config.arrakis.mu.Lock()
	s.config.arrakis.consecutiveEmptyMessages++
	s.config.arrakis.mu.Unlock()
}

// resetConsecutiveEmptyMessages resets the counter of consecutive empty responses
// to zero. This is called when messages are received, indicating that the queue
// is no longer idle and EWMA decay should be suspended.
//
// Thread-safe operation using mutex protection.
func (s *SQS) resetConsecutiveEmptyMessages() {
	s.config.arrakis.mu.Lock()
	s.config.arrakis.consecutiveEmptyMessages = 0
	s.config.arrakis.mu.Unlock()
}

// shouldDecayEWMA determines whether EWMA decay should be applied based on
// the number of consecutive empty responses. Decay is triggered when we've
// had enough empty responses to indicate a sustained idle period.
//
// Returns:
//   - bool: true if EWMA decay should be applied, false otherwise
func (s *SQS) shouldDecayEWMA() bool {
	return s.config.arrakis.consecutiveEmptyMessages >= _consecutiveEmptyThreshold
}

// calculateWaitTime determines the optimal SQS long polling wait time based on
// the current EWMA average message volume. The algorithm classifies volume into
// discrete categories and selects appropriate wait times for each category.
//
// Volume Classification:
// - Idle (avg = 0): No recent messages → longest wait time
// - Low (avg < 2): Very few messages → long wait time
// - Medium (avg 2-5): Moderate messages → medium wait time
// - High (avg 5-10): Many messages → short wait time
// - Very High (avg > 10): Constant messages → shortest wait time
//
// This classification optimizes the trade-off between API call frequency and
// message processing latency based on observed traffic patterns.
//
// Returns:
//   - int64: Optimal wait time in seconds for the next SQS ReceiveMessage call
func (s *SQS) calculateWaitTime() int64 {
	s.config.arrakis.mu.Lock()
	defer s.config.arrakis.mu.Unlock()

	avg := s.config.arrakis.average

	var waitTime int64

	switch {
	case avg == 0:
		// Idle: No recent messages, use maximum wait time
		waitTime = int64(s.config.AdaptivePolling.IdleWaitTimeSeconds)
	case avg < _lowVolumeThreshold:
		// Low volume: Few messages, use long wait time
		waitTime = int64(s.config.AdaptivePolling.LowVolumeWaitTimeSeconds)
	case avg < _mediumVolumeThreshold:
		// Medium volume: Moderate messages, use medium wait time
		waitTime = int64(s.config.AdaptivePolling.MediumVolumeWaitTimeSeconds)
	case avg < _highVolumeThreshold:
		// High volume: Many messages, use short wait time
		waitTime = int64(s.config.AdaptivePolling.HighVolumeWaitTimeSeconds)
	default:
		// Very high volume: Constant messages, use shortest wait time
		waitTime = int64(s.config.AdaptivePolling.VeryHighVolumeWaitTimeSeconds)
	}

	return waitTime
}

// decayEWMA applies exponential decay to the EWMA average during idle periods.
// This mechanism gradually reduces the average when no messages are being received,
// allowing the algorithm to adapt to decreased message volume without waiting
// for explicit volume drop detection.
//
// The decay uses a half-life approach: after each half-life period, the average
// is reduced by 50%. This provides smooth, predictable decay behavior that
// prevents the EWMA from remaining artificially high during extended idle periods.
//
// Decay conditions:
// 1. There must be a previous update (lastUpdate > 0)
// 2. Minimum time gap must have passed (prevents excessive decay)
// 3. Calculated decay factor is applied to current average
// 4. Very small averages are reset to zero (cleanup threshold)
func (s *SQS) decayEWMA() {
	// Get the last update timestamp atomically
	last := atomic.LoadInt64(&s.config.arrakis.lastUpdate)
	if last == 0 {
		// No previous updates, nothing to decay
		return
	}

	timeSinceLastUpdate := time.Since(time.Unix(last, 0))
	if timeSinceLastUpdate < _minDecayGapSeconds*time.Second {
		// Not enough time has passed, skip decay
		return
	}

	// Calculate exponential decay: decay = 0.5^(time_elapsed / half_life)
	decay := math.Pow(0.5, timeSinceLastUpdate.Seconds()/_halfLifeSeconds)

	s.config.arrakis.mu.Lock()
	defer s.config.arrakis.mu.Unlock()

	// Apply decay to current average
	s.config.arrakis.average *= decay

	// Reset very small averages to zero for cleaner behavior
	if s.config.arrakis.average < _ewmaDecayThreshold {
		s.config.arrakis.average = 0
	}
}
