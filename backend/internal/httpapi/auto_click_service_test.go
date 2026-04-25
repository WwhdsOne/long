package httpapi

import (
	"context"
	"testing"
	"time"

	"long/internal/vote"
)

type recordedChangePublisher struct {
	changes []vote.StateChange
}

func (p *recordedChangePublisher) PublishChange(_ context.Context, change vote.StateChange) error {
	p.changes = append(p.changes, change)
	return nil
}

func TestAutoClickServiceStartStatusStop(t *testing.T) {
	service := NewAutoClickService(AutoClickServiceOptions{
		Store: &mockStore{
			state: vote.State{
				Buttons: []vote.Button{
					{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
				},
			},
		},
		Interval:  time.Second / 3,
		AutoStart: false,
		Now: func() time.Time {
			return time.Unix(1710000000, 0)
		},
	})
	defer service.Close()

	status, err := service.Start(context.Background(), "阿明", "feel")
	if err != nil {
		t.Fatalf("start auto click: %v", err)
	}
	if !status.Active || status.ButtonKey != "feel" {
		t.Fatalf("expected active status for feel, got %+v", status)
	}

	status, err = service.Start(context.Background(), "阿明", "other")
	if err != nil {
		t.Fatalf("switch auto target: %v", err)
	}
	if status.ButtonKey != "other" {
		t.Fatalf("expected target switch to other, got %+v", status)
	}

	status = service.Status("阿明")
	if !status.Active || status.ButtonKey != "other" {
		t.Fatalf("expected persisted status, got %+v", status)
	}

	status = service.Stop("阿明")
	if status.Active {
		t.Fatalf("expected stopped status, got %+v", status)
	}
}

func TestAutoClickServiceRunOnceUsesSharedSettlementAndPublishesChange(t *testing.T) {
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
	publisher := &recordedChangePublisher{}
	service := NewAutoClickService(AutoClickServiceOptions{
		Store:           store,
		ChangePublisher: publisher,
		Interval:        time.Second / 3,
		AutoStart:       false,
		Now: func() time.Time {
			return time.Unix(1710000000, 0)
		},
	})
	defer service.Close()

	if _, err := service.Start(context.Background(), "阿明", "feel"); err != nil {
		t.Fatalf("start auto click: %v", err)
	}

	service.runOnce(context.Background())

	if store.lastAutoClickNickname != "阿明" {
		t.Fatalf("expected auto click to use official boss settlement without manual click count, got nickname %q", store.lastAutoClickNickname)
	}
	if len(publisher.changes) != 1 {
		t.Fatalf("expected one published change, got %d", len(publisher.changes))
	}
	if publisher.changes[0].Type != vote.StateChangeButtonClicked {
		t.Fatalf("expected button click change, got %+v", publisher.changes[0])
	}
}
