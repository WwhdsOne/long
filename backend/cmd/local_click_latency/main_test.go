package main

import (
	"testing"
	"time"

	"github.com/hertz-contrib/websocket"
	"google.golang.org/protobuf/proto"

	"long/internal/realtimepb"
)

func TestClassifyServerFrameReturnsClickAck(t *testing.T) {
	body, err := proto.Marshal(&realtimepb.ClickAck{Delta: 3})
	if err != nil {
		t.Fatalf("marshal click ack: %v", err)
	}
	frame := append([]byte{realtimeBinaryTypeClickAck}, body...)

	ack, ok, err := classifyServerFrame(websocket.BinaryMessage, frame)
	if err != nil {
		t.Fatalf("classify click ack: %v", err)
	}
	if !ok {
		t.Fatal("expected binary click_ack to be recognized")
	}
	if ack.GetDelta() != 3 {
		t.Fatalf("expected delta 3, got %d", ack.GetDelta())
	}
}

func TestClassifyServerFrameIgnoresOnlineCountText(t *testing.T) {
	ack, ok, err := classifyServerFrame(websocket.TextMessage, []byte(`{"type":"online_count","payload":{"count":2}}`))
	if err != nil {
		t.Fatalf("expected online_count to be ignored, got %v", err)
	}
	if ok {
		t.Fatal("expected online_count not to be treated as click_ack")
	}
	if ack != nil {
		t.Fatalf("expected nil ack for online_count, got %+v", ack)
	}
}

func TestClassifyServerFrameReturnsErrorForProtocolErrorText(t *testing.T) {
	_, ok, err := classifyServerFrame(websocket.TextMessage, []byte(`{"type":"error","code":"BOSS_PART_NOT_FOUND","message":"Boss 部位不存在或当前不可攻击。"}`))
	if err == nil {
		t.Fatal("expected text error frame to return error")
	}
	if ok {
		t.Fatal("expected text error not to be treated as click_ack")
	}
}

func TestValidateConfigRejectsNonPositiveConnections(t *testing.T) {
	err := validateConfig(config{
		baseURL:     "https://www.wclick.top",
		nickname:    "Wwhds",
		password:    "123456",
		slug:        "boss-part:0-0",
		count:       10,
		connections: 0,
		timeout:     10 * time.Second,
	})
	if err == nil {
		t.Fatal("expected validateConfig to reject non-positive connections")
	}
}

func TestBuildRunSummaryAggregatesAllConnections(t *testing.T) {
	summary := buildRunSummary([][]time.Duration{
		{10 * time.Millisecond, 20 * time.Millisecond},
		{30 * time.Millisecond},
	}, time.Second)

	if summary.totalSamples != 3 {
		t.Fatalf("expected 3 total samples, got %d", summary.totalSamples)
	}
	if len(summary.perConnection) != 2 {
		t.Fatalf("expected 2 per-connection summaries, got %d", len(summary.perConnection))
	}
	if summary.overall.Min != 10*time.Millisecond {
		t.Fatalf("expected overall min 10ms, got %s", summary.overall.Min)
	}
	if summary.overall.Max != 30*time.Millisecond {
		t.Fatalf("expected overall max 30ms, got %s", summary.overall.Max)
	}
	if summary.overall.P50 != 20*time.Millisecond {
		t.Fatalf("expected overall p50 20ms, got %s", summary.overall.P50)
	}
	if summary.overall.P95 != 30*time.Millisecond {
		t.Fatalf("expected overall p95 30ms, got %s", summary.overall.P95)
	}
	if summary.overall.QPS != 3 {
		t.Fatalf("expected overall qps 3, got %.2f", summary.overall.QPS)
	}
}
