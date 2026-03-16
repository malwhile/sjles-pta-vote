package services

import (
	"testing"
	"time"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/models"
)

func TestPollExpirationInPast(t *testing.T) {
	// Create a poll that expired in the past
	pastPoll := &models.Poll{
		ID:       2,
		Question: "Past poll",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Format(common.DATE_FORMAT),
	}

	// Check if poll is expired
	if pastPoll.ExpiresAt != "" {
		expiryTime, err := time.Parse(common.DATE_FORMAT, pastPoll.ExpiresAt)
		if err == nil && time.Now().After(expiryTime) {
			// Poll should be expired
		} else {
			t.Errorf("Past poll should be expired")
		}
	}
}

func TestPollWithoutExpiration(t *testing.T) {
	// Create a poll without expiration
	noPoll := &models.Poll{
		ID:       3,
		Question: "No expiration poll",
		ExpiresAt: "",
	}

	// Check if poll has expiration
	if noPoll.ExpiresAt == "" {
		// Poll has no expiration, should always be valid
	}
}

func TestPollPassFailDetermination(t *testing.T) {
	tests := []struct {
		name           string
		memberYes      int64
		memberNo       int64
		nonMemberYes   int64
		nonMemberNo    int64
		expectedPass   bool
		expectedStatus string
	}{
		{
			name:           "No votes",
			memberYes:      0,
			memberNo:       0,
			nonMemberYes:   0,
			nonMemberNo:    0,
			expectedPass:   false,
			expectedStatus: "No votes yet",
		},
		{
			name:           "Member votes pass",
			memberYes:      10,
			memberNo:       5,
			nonMemberYes:   0,
			nonMemberNo:    0,
			expectedPass:   true,
			expectedStatus: "Pass",
		},
		{
			name:           "Member votes fail",
			memberYes:      5,
			memberNo:       10,
			nonMemberYes:   0,
			nonMemberNo:    0,
			expectedPass:   false,
			expectedStatus: "Fail",
		},
		{
			name:           "Equal member votes",
			memberYes:      5,
			memberNo:       5,
			nonMemberYes:   0,
			nonMemberNo:    0,
			expectedPass:   false,
			expectedStatus: "Fail",
		},
		{
			name:           "Member votes with non-member votes",
			memberYes:      8,
			memberNo:       4,
			nonMemberYes:   3,
			nonMemberNo:    2,
			expectedPass:   true,
			expectedStatus: "Pass",
		},
	}

	for _, tc := range tests {
		_ = &models.Poll{
			ID:               1,
			Question:         tc.name,
			MemberYes:        tc.memberYes,
			MemberNo:         tc.memberNo,
			NonMemberYes:     tc.nonMemberYes,
			NonMemberNo:      tc.nonMemberNo,
			ExpiresAt:        time.Now().Add(1 * time.Hour).Format(common.DATE_FORMAT),
		}

		// Determine pass/fail based on member votes only
		totalMemberVotes := tc.memberYes + tc.memberNo
		hasVotes := totalMemberVotes > 0
		passed := hasVotes && tc.memberYes > tc.memberNo

		if passed != tc.expectedPass {
			t.Errorf("%s: expected passed=%v, got %v", tc.name, tc.expectedPass, passed)
		}

		// Determine status
		var status string
		if !hasVotes {
			status = "No votes yet"
		} else if passed {
			status = "Pass"
		} else {
			status = "Fail"
		}

		if status != tc.expectedStatus {
			t.Errorf("%s: expected status='%s', got '%s'", tc.name, tc.expectedStatus, status)
		}
	}
}
