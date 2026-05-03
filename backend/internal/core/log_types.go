package core

// AdminAuditLog 后台操作审计日志。
type AdminAuditLog struct {
	Operator       string `json:"operator"`
	Action         string `json:"action"`
	RoomID         string `json:"roomId,omitempty"`
	QueueID        string `json:"queueId,omitempty"`
	TargetType     string `json:"targetType,omitempty"`
	TargetID       string `json:"targetId,omitempty"`
	RequestPath    string `json:"requestPath,omitempty"`
	RequestIP      string `json:"requestIp,omitempty"`
	PayloadSummary string `json:"payloadSummary,omitempty"`
	Result         string `json:"result"`
	ErrorCode      string `json:"errorCode,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
}

// DomainEvent 业务事件日志。
type DomainEvent struct {
	EventType string         `json:"eventType"`
	Nickname  string         `json:"nickname,omitempty"`
	BossID    string         `json:"bossId,omitempty"`
	RoomID    string         `json:"roomId,omitempty"`
	QueueID   string         `json:"queueId,omitempty"`
	ItemID    string         `json:"itemId,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt int64          `json:"createdAt"`
}
