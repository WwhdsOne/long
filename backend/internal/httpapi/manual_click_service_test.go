package httpapi

import (
	"context"
	"testing"
	"time"

	"long/internal/vote"
)

func TestManualClickServiceAcceptsFreshTicketOnce(t *testing.T) {
	now := time.Unix(1710000000, 0)
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
		},
		result: vote.ClickResult{
			Button: vote.Button{Key: "feel", Label: "有感觉吗", Count: 4, Enabled: true},
			Delta:  1,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 4,
			},
		},
	}
	service := NewManualClickService(ManualClickServiceOptions{
		Store: store,
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	result, err := service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior:         validClickBehavior(),
	})
	if err != nil {
		t.Fatalf("consume ticket: %v", err)
	}
	if result.Button.Key != "feel" {
		t.Fatalf("expected click result for feel, got %+v", result)
	}
	if store.lastClickNickname != "阿明" {
		t.Fatalf("expected manual click to use 阿明, got %q", store.lastClickNickname)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: "bad-proof",
		Behavior:         validClickBehavior(),
	})
	if !manualClickRequiresRetry(err) {
		t.Fatalf("expected reused ticket to require retry, got %v", err)
	}
}

func TestManualClickServiceRejectsExpiredOrMismatchedTicket(t *testing.T) {
	now := time.Unix(1710000000, 0)
	service := NewManualClickService(ManualClickServiceOptions{
		Store: &mockStore{
			state: vote.State{
				Buttons: []vote.Button{
					{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
				},
			},
		},
		Config: ManualClickConfig{
			TicketTTL:             time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "other",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior:         validClickBehavior(),
	})
	if !manualClickRequiresRetry(err) {
		t.Fatalf("expected mismatched slug to require retry, got %v", err)
	}

	now = now.Add(3 * time.Second)
	expiredTicket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue expired ticket: %v", err)
	}

	now = now.Add(2 * time.Second)
	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           expiredTicket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", expiredTicket.Value, expiredTicket.ChallengeNonce),
		Behavior:         validClickBehavior(),
	})
	if !manualClickRequiresRetry(err) {
		t.Fatalf("expected expired ticket to require retry, got %v", err)
	}
}

func TestManualClickServiceBansRepeatedAbuseTemporarily(t *testing.T) {
	now := time.Unix(1710000000, 0)
	service := NewManualClickService(ManualClickServiceOptions{
		Store: &mockStore{
			state: vote.State{
				Buttons: []vote.Button{
					{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
				},
			},
		},
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   1,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         2,
			BanDuration:           2 * time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	if _, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	}); err != nil {
		t.Fatalf("first issue ticket: %v", err)
	}

	if _, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	}); !manualClickTooFrequent(err) {
		t.Fatalf("expected second issue to be throttled, got %v", err)
	}

	_, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if !manualClickTooFrequent(err) {
		t.Fatalf("expected third issue to be blocked, got %v", err)
	}
	if retryAfter := manualClickRetryAfter(err); retryAfter < time.Minute {
		t.Fatalf("expected ban retry-after to be at least one minute, got %s", retryAfter)
	}

	now = now.Add(3 * time.Minute)
	if _, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	}); err != nil {
		t.Fatalf("expected issue to recover after ban window, got %v", err)
	}
}

func TestManualClickServiceRejectsMissingFingerprintOrBehavior(t *testing.T) {
	now := time.Unix(1710000000, 0)
	service := NewManualClickService(ManualClickServiceOptions{
		Store: &mockStore{
			state: vote.State{
				Buttons: []vote.Button{
					{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
				},
			},
		},
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	if _, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname: "阿明",
		Slug:     "feel",
		ClientID: "127.0.0.1",
	}); !manualClickRequiresRetry(err) {
		t.Fatalf("expected missing fingerprint on issue to require retry, got %v", err)
	}

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior: ClickBehavior{
			PressDurationMS: 5,
			Trajectory: []ClickPointerSample{
				{X: 1, Y: 1, T: 0},
				{X: 8, Y: 4, T: 2},
				{X: 15, Y: 10, T: 5},
			},
		},
	})
	if !manualClickRequiresRetry(err) {
		t.Fatalf("expected invalid behavior to require retry, got %v", err)
	}
}

func TestManualClickServiceAcceptsLowMovementHumanClick(t *testing.T) {
	now := time.Unix(1710000000, 0)
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
		},
		result: vote.ClickResult{
			Button: vote.Button{Key: "feel", Label: "有感觉吗", Count: 4, Enabled: true},
			Delta:  1,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 4,
			},
		},
	}
	service := NewManualClickService(ManualClickServiceOptions{
		Store: store,
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior: ClickBehavior{
			PointerType:     "mouse",
			PressDurationMS: 120,
			Trajectory: []ClickPointerSample{
				{X: 100, Y: 100, T: 0},
				{X: 100, Y: 100, T: 120},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected low movement human click to pass, got %v", err)
	}
}

func TestManualClickServiceAcceptsShortLowMovementClick(t *testing.T) {
	now := time.Unix(1710000000, 0)
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
		},
		result: vote.ClickResult{
			Button: vote.Button{Key: "feel", Label: "有感觉吗", Count: 4, Enabled: true},
			Delta:  1,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 4,
			},
		},
	}
	service := NewManualClickService(ManualClickServiceOptions{
		Store: store,
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior: ClickBehavior{
			PointerType:     "mouse",
			PressDurationMS: 5,
			Trajectory: []ClickPointerSample{
				{X: 100, Y: 100, T: 0},
				{X: 100, Y: 100, T: 5},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected short low movement click to pass, got %v", err)
	}
}

func TestManualClickServiceAcceptsFastHumanClick(t *testing.T) {
	now := time.Unix(1710000000, 0)
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
		},
		result: vote.ClickResult{
			Button: vote.Button{Key: "feel", Label: "有感觉吗", Count: 4, Enabled: true},
			Delta:  1,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 4,
			},
		},
	}
	service := NewManualClickService(ManualClickServiceOptions{
		Store: store,
		Config: ManualClickConfig{
			TicketTTL:             2 * time.Second,
			IssueLimitPerSecond:   5,
			ConsumeLimitPerSecond: 5,
			RiskThreshold:         3,
			BanDuration:           time.Minute,
			MinPressDuration:      20 * time.Millisecond,
			MaxPressDuration:      2 * time.Second,
			MinTrajectoryPoints:   4,
			MaxTrajectoryPoints:   12,
			MinPathDistance:       10,
			MinDisplacement:       2,
			MinCurvature:          0.05,
			MinSpeedVariance:      0.01,
		},
		Now: func() time.Time {
			return now
		},
	})

	ticket, err := service.IssueTicket(context.Background(), TicketIssueRequest{
		Nickname:        "阿明",
		Slug:            "feel",
		ClientID:        "127.0.0.1",
		FingerprintHash: "fp-1",
	})
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	_, err = service.Click(context.Background(), ManualClickRequest{
		Nickname:         "阿明",
		Slug:             "feel",
		Ticket:           ticket.Value,
		ClientID:         "127.0.0.1",
		EntryType:        clickEntryHTTP,
		FingerprintHash:  "fp-1",
		FingerprintProof: fingerprintProof("fp-1", ticket.Value, ticket.ChallengeNonce),
		Behavior: ClickBehavior{
			PointerType:     "mouse",
			PressDurationMS: 6,
			Trajectory: []ClickPointerSample{
				{X: 100, Y: 100, T: 0},
				{X: 104, Y: 102, T: 2},
				{X: 109, Y: 107, T: 4},
				{X: 113, Y: 110, T: 6},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected fast human click to pass, got %v", err)
	}
}

func validClickBehavior() ClickBehavior {
	return ClickBehavior{
		PointerType:     "mouse",
		PressDurationMS: 120,
		Trajectory: []ClickPointerSample{
			{X: 10, Y: 10, T: 0},
			{X: 13, Y: 12, T: 30},
			{X: 17, Y: 18, T: 70},
			{X: 22, Y: 21, T: 120},
		},
	}
}
