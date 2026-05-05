package main

import (
	"testing"

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
