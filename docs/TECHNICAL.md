# Arrakis Technical Documentation

## Algorithm Overview

The Arrakis library implements an adaptive polling algorithm for Amazon SQS using EWMA (Exponentially Weighted Moving Average) to automatically optimize polling intervals based on observed message volume.

## How It Works

### 1. Data Collection
- Each `ReceiveMessage` operation records how many messages were returned
- Empty and non-empty responses are handled differently
- Timestamps are maintained for temporal decay calculations

### 2. EWMA Calculation
```
new_value = α × current_observation + (1-α) × previous_value
```

Where:
- `α` (alpha) is the smoothing factor (0.0 to 1.0)
- Lower values = more stable, less responsive
- Higher values = more responsive, less stable

### 3. Volume Classification

| Category | Criteria (EWMA avg) | Default Wait Time |
|----------|-------------------|------------------|
| Idle | = 0 | 20 seconds |
| Low Volume | < 2 messages | 15 seconds |
| Medium Volume | 2-5 messages | 10 seconds |
| High Volume | 5-10 messages | 5 seconds |
| Very High Volume | > 10 messages | 1 second |

### 4. Protections and Optimizations

#### Spike Protection
- Limits sudden changes to at most 200% of current value
- Prevents outliers from dramatically distorting EWMA

#### Volume Drop Detection
- Monitors consecutive cycles of low volume
- Automatic EWMA reset when appropriate
- Prevents algorithm from "sticking" at high values

#### Temporal Decay
- Applies exponential decay during idle periods
- Configurable half-life (default: 30 seconds)
- Automatic reset of very small values

## Configuration Parameters

### Main Parameters
```go
type adaptivePolling struct {
    EnableAdaptivePolling bool     // Enable/disable the algorithm
    
    // Wait times by volume category
    IdleWaitTimeSeconds           int
    LowVolumeWaitTimeSeconds      int
    MediumVolumeWaitTimeSeconds   int
    HighVolumeWaitTimeSeconds     int
    VeryHighVolumeWaitTimeSeconds int
    
    // EWMA algorithm parameters
    EwmaAlpha              float64  // Smoothing factor (0.1-0.5 recommended)
    DropDetectionThreshold int      // Cycles before reset (5-15 recommended)
}
```

### Internal Constants
```go
// Classification thresholds
_lowVolumeThreshold    = 2   // Threshold between idle and low
_mediumVolumeThreshold = 5   // Threshold between low and medium
_highVolumeThreshold   = 10  // Threshold between medium and high

// Decay configuration
_halfLifeSeconds    = 30.0  // Half-life for exponential decay
_ewmaDecayThreshold = 0.2   // Threshold for automatic reset

// Drop detection
_lowVolumeMessageThreshold = 2    // Threshold for low volume cycle
_ewmaResetAverageThreshold = 1.0  // Threshold for reset eligibility
_consecutiveEmptyThreshold = 2    // Empty responses before decay
```

## Usage Scenarios

### 1. Constant High Volume
- EWMA converges to ~10+ messages
- Uses `VeryHighVolumeWaitTimeSeconds` (1s)
- Frequent polling for low latency

### 2. Variable Volume
- EWMA adjusts dynamically
- Smooth transitions between categories
- Balances cost vs latency

### 3. Idle Period
- Gradual decay reduces EWMA
- Transition to longer wait times
- Reduces cost of unnecessary API calls

### 4. Message Bursts
- Spike protection prevents over-reaction
- Quick but controlled adaptation
- Gradual return to normal volumes

## Metrics and Observability

### Exposed Internal State
```go
type arrakis struct {
    messageCount    int64    // Last message count
    lastUpdate      int64    // Timestamp of last update
    average         float64  // Current EWMA value
    lowVolumeCycle  int      // Low volume cycle counter
    consecutiveEmptyMessages int64  // Empty response counter
}
```

### Suggested Logs
- Current EWMA and calculated wait time
- Transitions between volume categories
- Reset and decay events
- Processed message counters

## Practical Examples

### Conservative Configuration (Stable)
```go
WithAdaptivePolling(20, 30, 15, 10, 5, 1, 0.2, 10)
// Low alpha = more stable
// Default thresholds
```

### Responsive Configuration (Agile)
```go
WithAdaptivePolling(15, 30, 10, 6, 3, 1, 0.4, 6)
// High alpha = more responsive
// Shorter wait times
// More frequent resets
```

### High Volume Configuration
```go
WithAdaptivePolling(10, 60, 8, 5, 2, 1, 0.3, 8)
// Wait times optimized for throughput
// Longer visibility timeout for processing
```

## Monitoring and Debugging

### Signs of Good Performance
- EWMA converges to stable values
- Smooth transitions between categories
- Few unnecessary resets
- Appropriate decay during idle periods

### Problem Indicators
- EWMA constantly oscillating
- Too frequent resets
- Inadequate wait times for traffic pattern
- Excessive delay in adapting to changes

### Recommended Adjustments
- **Unstable EWMA**: Reduce alpha
- **Slow adaptation**: Increase alpha
- **Frequent resets**: Increase DropDetectionThreshold  
- **Doesn't detect drops**: Decrease DropDetectionThreshold