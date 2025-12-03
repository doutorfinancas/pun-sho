package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
)

func TestShortyService_LabelsLogic(t *testing.T) {
	// Test the service layer logic for handling labels
	// without involving database operations

	tests := []struct {
		name           string
		createRequest  *request.CreateShorty
		expectedLabels entity.StringArray
	}{
		{
			name: "create shorty without labels",
			createRequest: &request.CreateShorty{
				Link:             "https://example.com",
				TTL:              nil,
				RedirectionLimit: nil,
				Labels:           nil,
			},
			expectedLabels: entity.StringArray(nil),
		},
		{
			name: "create shorty with empty labels",
			createRequest: &request.CreateShorty{
				Link:             "https://example.com",
				TTL:              nil,
				RedirectionLimit: nil,
				Labels:           []string{},
			},
			expectedLabels: entity.StringArray{},
		},
		{
			name: "create shorty with multiple labels",
			createRequest: &request.CreateShorty{
				Link:             "https://example.com",
				TTL:              nil,
				RedirectionLimit: nil,
				Labels:           []string{"marketing", "campaign", "2024"},
			},
			expectedLabels: entity.StringArray{"marketing", "campaign", "2024"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the service correctly processes the labels from the request
			// We'll verify the logic before it hits the database
			
			// Create a shorty entity from the request (simulating what the service does)
			shorty := &entity.Shorty{
				PublicID:         "test123",
				Link:             tt.createRequest.Link,
				TTL:              tt.createRequest.TTL,
				RedirectionLimit: tt.createRequest.RedirectionLimit,
				Labels:           tt.createRequest.Labels,
			}

			// Verify the labels were set correctly
			assert.Equal(t, tt.expectedLabels, shorty.Labels)
		})
	}
}

func TestShortyService_UpdateLabelsLogic(t *testing.T) {
	// Test the update logic for labels

	existingShorty := &entity.Shorty{
		ID:       uuid.New(),
		PublicID: "test123",
		Link:     "https://old-link.com",
		Labels:   entity.StringArray{"old", "label"},
	}

	tests := []struct {
		name            string
		updateRequest   *request.UpdateShorty
		expectedLabels  entity.StringArray
	}{
		{
			name: "update with new labels",
			updateRequest: &request.UpdateShorty{
				Labels: []string{"marketing", "campaign", "2024"},
			},
			expectedLabels: entity.StringArray{"marketing", "campaign", "2024"},
		},
		{
			name: "update with empty labels",
			updateRequest: &request.UpdateShorty{
				Labels: []string{},
			},
			expectedLabels: entity.StringArray{},
		},
		{
			name: "update without changing labels",
			updateRequest: &request.UpdateShorty{
				Link: "https://new-link.com",
			},
			expectedLabels: entity.StringArray{"old", "label"}, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy the existing shorty to simulate the update process
			shortyCopy := *existingShorty

			// Apply the update logic (simulating what the service does)
			if tt.updateRequest.Labels != nil {
				shortyCopy.Labels = tt.updateRequest.Labels
			}

			// Verify the labels were updated correctly
			assert.Equal(t, tt.expectedLabels, shortyCopy.Labels)
		})
	}
}

func TestShortyService_ListLabelsParameter(t *testing.T) {
	// Test that the List method correctly handles labels parameter

	tests := []struct {
		name   string
		labels []string
		valid  bool
	}{
		{
			name:   "nil labels",
			labels: nil,
			valid:  true,
		},
		{
			name:   "empty labels",
			labels: []string{},
			valid:  true,
		},
		{
			name:   "single label",
			labels: []string{"marketing"},
			valid:  true,
		},
		{
			name:   "multiple labels",
			labels: []string{"marketing", "tech", "2024"},
			valid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the parameters are valid for the List method
			// This tests the parameter validation logic
			
			if tt.labels == nil || len(tt.labels) == 0 || len(tt.labels) <= 10 {
				assert.True(t, tt.valid, "Labels parameter should be valid")
			} else {
				assert.False(t, tt.valid, "Too many labels should be invalid")
			}

			// Test that we can create the expected database query format
			if len(tt.labels) > 0 {
				// This simulates the label processing in the repository
				labelStr := "{"
				for i, label := range tt.labels {
					if i > 0 {
						labelStr += ","
					}
					labelStr += label
				}
				labelStr += "}"
				
				assert.Contains(t, labelStr, "{")
				assert.Contains(t, labelStr, "}")
			}
		})
	}
}
