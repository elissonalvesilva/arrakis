package sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Basic SQS configuration tests
func TestNewSQS(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	if client == nil {
		t.Error("Expected non-nil SQS client")
	}

	if client.client == nil {
		t.Error("Expected non-nil internal client")
	}
}

func TestEnableArrakis(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	client.EnableArrakis()

	if !client.IsArrakisEnabled() {
		t.Error("Expected Arrakis to be enabled")
	}
}

func TestDisableArrakis(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	// Enable first
	client.EnableArrakis()
	if !client.IsArrakisEnabled() {
		t.Error("Setup failed: Expected Arrakis to be enabled")
	}

	// Then disable
	client.DisableArrakis()
	if client.IsArrakisEnabled() {
		t.Error("Expected Arrakis to be disabled")
	}
}

func TestIsArrakisEnabled(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	// Default should be disabled
	if client.IsArrakisEnabled() {
		t.Error("Expected Arrakis to be disabled by default")
	}

	// Enable and test
	client.EnableArrakis()
	if !client.IsArrakisEnabled() {
		t.Error("Expected Arrakis to be enabled after EnableArrakis()")
	}

	// Disable and test again
	client.DisableArrakis()
	if client.IsArrakisEnabled() {
		t.Error("Expected Arrakis to be disabled after DisableArrakis()")
	}
}

// Test constants and default values
func TestConstants(t *testing.T) {
	// Test that default values are reasonable
	if _defaultNumberOfMessages < 1 || _defaultNumberOfMessages > 10 {
		t.Errorf("_defaultNumberOfMessages should be between 1 and 10, got %d", _defaultNumberOfMessages)
	}

	// Verify the expected default value
	if _defaultNumberOfMessages != 10 {
		t.Errorf("Expected _defaultNumberOfMessages to be 10, got %d", _defaultNumberOfMessages)
	}
}

// Test basic functionality without mocks - integration style tests
func TestSQSStructure(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	// Test that the struct is properly initialized
	if client == nil {
		t.Fatal("NewSQS returned nil")
	}

	// Test that methods don't panic when called
	t.Run("EnableArrakis doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("EnableArrakis panicked: %v", r)
			}
		}()
		client.EnableArrakis()
	})

	t.Run("DisableArrakis doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DisableArrakis panicked: %v", r)
			}
		}()
		client.DisableArrakis()
	})

	t.Run("IsArrakisEnabled doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IsArrakisEnabled panicked: %v", r)
			}
		}()
		_ = client.IsArrakisEnabled()
	})
}

// Test ReceiveMessage method signature and parameter handling
func TestReceiveMessage_ParameterValidation(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	tests := []struct {
		name              string
		queueURL          string
		maxMsg            int32
		messageAttributes map[string]string
		shouldPanic       bool
		description       string
	}{
		{
			name:              "Valid parameters",
			queueURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			maxMsg:            5,
			messageAttributes: map[string]string{"Priority": "", "Author": ""},
			shouldPanic:       false,
			description:       "Should handle valid parameters without panic",
		},
		{
			name:              "Empty queue URL",
			queueURL:          "",
			maxMsg:            10,
			messageAttributes: nil,
			shouldPanic:       false,
			description:       "Should handle empty queue URL (will fail at AWS level)",
		},
		{
			name:              "Zero maxMsg",
			queueURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			maxMsg:            0,
			messageAttributes: nil,
			shouldPanic:       false,
			description:       "Should handle zero maxMsg (should use default)",
		},
		{
			name:              "Nil message attributes",
			queueURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			maxMsg:            1,
			messageAttributes: nil,
			shouldPanic:       false,
			description:       "Should handle nil message attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("ReceiveMessage panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("ReceiveMessage should have panicked but didn't")
				}
			}()

			// Note: This will likely fail with AWS errors since we don't have real credentials
			// But it tests that the method signature works and parameter handling doesn't panic
			_, _ = client.ReceiveMessage(context.Background(), tt.queueURL, tt.maxMsg, tt.messageAttributes)

			t.Logf("Test passed: %s", tt.description)
		})
	}
}

// Test DeleteMessage method signature and parameter handling
func TestDeleteMessage_ParameterValidation(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	tests := []struct {
		name          string
		queueURL      string
		receiptHandle string
		shouldPanic   bool
		description   string
	}{
		{
			name:          "Valid parameters",
			queueURL:      "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			receiptHandle: "AQEBwJnKyrHigUMZj6rYigCgxlaS3SLy0a",
			shouldPanic:   false,
			description:   "Should handle valid parameters without panic",
		},
		{
			name:          "Empty queue URL",
			queueURL:      "",
			receiptHandle: "AQEBwJnKyrHigUMZj6rYigCgxlaS3SLy0a",
			shouldPanic:   false,
			description:   "Should handle empty queue URL (will fail at AWS level)",
		},
		{
			name:          "Empty receipt handle",
			queueURL:      "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			receiptHandle: "",
			shouldPanic:   false,
			description:   "Should handle empty receipt handle (will fail at AWS level)",
		},
		{
			name:          "Both empty",
			queueURL:      "",
			receiptHandle: "",
			shouldPanic:   false,
			description:   "Should handle both empty parameters (will fail at AWS level)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("DeleteMessage panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("DeleteMessage should have panicked but didn't")
				}
			}()

			// Note: This will likely fail with AWS errors since we don't have real credentials
			// But it tests that the method signature works and parameter handling doesn't panic
			_, _ = client.DeleteMessage(context.Background(), tt.queueURL, tt.receiptHandle)

			t.Logf("Test passed: %s", tt.description)
		})
	}
}

// Test Arrakis behavior integration
func TestArrakisIntegration(t *testing.T) {
	config := &aws.Config{}
	client := NewSQS(config)

	// Test state transitions
	t.Run("Arrakis state management", func(t *testing.T) {
		// Default state
		if client.IsArrakisEnabled() {
			t.Error("Arrakis should be disabled by default")
		}

		// Enable Arrakis
		client.EnableArrakis()
		if !client.IsArrakisEnabled() {
			t.Error("Arrakis should be enabled after EnableArrakis()")
		}

		// Test that ReceiveMessage can be called with Arrakis enabled
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ReceiveMessage with Arrakis enabled panicked: %v", r)
			}
		}()

		// This will fail with AWS errors but shouldn't panic
		_, _ = client.ReceiveMessage(context.Background(), "test-queue", 1, nil)

		// Disable Arrakis
		client.DisableArrakis()
		if client.IsArrakisEnabled() {
			t.Error("Arrakis should be disabled after DisableArrakis()")
		}

		// Test that ReceiveMessage can be called with Arrakis disabled
		_, _ = client.ReceiveMessage(context.Background(), "test-queue", 1, nil)
	})
}
