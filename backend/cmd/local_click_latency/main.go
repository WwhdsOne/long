package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/websocket"
	"google.golang.org/protobuf/proto"

	"long/internal/realtimepb"
)

const (
	realtimeBinaryTypeClickRequest byte = 1
	realtimeBinaryTypeClickAck     byte = 2
	playerSessionCookieName             = "long_player_session"
)

type config struct {
	baseURL        string
	nickname       string
	password       string
	slug           string
	count          int
	connections    int
	pause          time.Duration
	timeout        time.Duration
	handshakeWait  time.Duration
	insecureOrigin bool
}

type loginResponse struct {
	Authenticated bool   `json:"authenticated"`
	Nickname      string `json:"nickname"`
	Error         string `json:"error"`
	Message       string `json:"message"`
}

type realtimeTextEnvelope struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type latencyStats struct {
	Min     time.Duration
	Max     time.Duration
	Avg     time.Duration
	P50     time.Duration
	P95     time.Duration
	P99     time.Duration
	Elapsed time.Duration
	QPS     float64
}

type workerResult struct {
	connectionIndex int
	latencies       []time.Duration
	err             error
}

type runSummary struct {
	overall       latencyStats
	perConnection []latencyStats
	totalSamples  int
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.baseURL, "base", "", "站点地址，例如 https://example.com")
	flag.StringVar(&cfg.nickname, "nickname", "", "压测账号昵称")
	flag.StringVar(&cfg.password, "password", "", "压测账号密码")
	flag.StringVar(&cfg.slug, "slug", "", "Boss 部位 slug，例如 boss-part:1-0")
	flag.IntVar(&cfg.count, "count", 200, "每条连接发送点击次数")
	flag.IntVar(&cfg.connections, "connections", 1, "WebSocket 连接数")
	flag.DurationVar(&cfg.pause, "pause", 0, "每次点击之间的停顿，例如 5ms")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "HTTP 和单次读写超时")
	flag.DurationVar(&cfg.handshakeWait, "handshake-wait", 0, "建连后额外等待时间，例如 200ms")
	flag.BoolVar(&cfg.insecureOrigin, "insecure-origin", false, "附带 Origin: baseURL，便于某些网关校验")
	flag.Parse()

	return cfg
}

func run(cfg config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	cookie, nickname, err := login(cfg)
	if err != nil {
		return err
	}

	connections, err := connectWebSockets(cfg, cookie)
	if err != nil {
		return err
	}
	defer closeWebSockets(connections)

	if cfg.handshakeWait > 0 {
		time.Sleep(cfg.handshakeWait)
	}

	frame, err := packClickRequest(cfg.slug)
	if err != nil {
		return err
	}

	results := make(chan workerResult, cfg.connections)
	startCh := make(chan struct{})
	var wg sync.WaitGroup
	for index, conn := range connections {
		wg.Add(1)
		go func(connectionIndex int, conn *websocket.Conn) {
			defer wg.Done()
			results <- runConnectionWorker(connectionIndex, conn, frame, cfg, startCh)
		}(index+1, conn)
	}

	startAll := time.Now()
	close(startCh)
	go func() {
		wg.Wait()
		close(results)
	}()

	perConnectionLatencies := make([][]time.Duration, cfg.connections)
	var firstErr error
	for result := range results {
		if result.connectionIndex > 0 && result.connectionIndex <= len(perConnectionLatencies) {
			perConnectionLatencies[result.connectionIndex-1] = result.latencies
		}
		if result.err != nil && firstErr == nil {
			firstErr = result.err
		}
	}
	elapsed := time.Since(startAll)
	if firstErr != nil {
		return firstErr
	}

	summary := buildRunSummary(perConnectionLatencies, elapsed)
	printSummary(cfg, nickname, summary)
	return nil
}

func validateConfig(cfg config) error {
	switch {
	case strings.TrimSpace(cfg.baseURL) == "":
		return errors.New("缺少 -base")
	case strings.TrimSpace(cfg.nickname) == "":
		return errors.New("缺少 -nickname")
	case strings.TrimSpace(cfg.password) == "":
		return errors.New("缺少 -password")
	case strings.TrimSpace(cfg.slug) == "":
		return errors.New("缺少 -slug")
	case !strings.HasPrefix(strings.TrimSpace(cfg.slug), "boss-part:"):
		return errors.New("-slug 必须是 boss-part:x-y，例如 boss-part:1-0")
	case cfg.count <= 0:
		return errors.New("-count 必须大于 0")
	case cfg.connections <= 0:
		return errors.New("-connections 必须大于 0")
	case cfg.timeout <= 0:
		return errors.New("-timeout 必须大于 0")
	}
	return nil
}

func login(cfg config) (*http.Cookie, string, error) {
	loginURL := strings.TrimRight(cfg.baseURL, "/") + "/api/player/auth/login"
	body, err := json.Marshal(map[string]string{
		"nickname": cfg.nickname,
		"password": cfg.password,
	})
	if err != nil {
		return nil, "", fmt.Errorf("编码登录请求失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("创建登录请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: cfg.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	var parsed loginResponse
	_ = json.Unmarshal(rawBody, &parsed)

	if resp.StatusCode != http.StatusOK {
		if parsed.Message != "" {
			return nil, "", fmt.Errorf("登录失败: %s (%s)", parsed.Message, parsed.Error)
		}
		return nil, "", fmt.Errorf("登录失败: HTTP %d", resp.StatusCode)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == playerSessionCookieName {
			return cookie, parsed.Nickname, nil
		}
	}
	return nil, "", errors.New("登录成功但未拿到 long_player_session cookie")
}

func connectWebSocket(cfg config, cookie *http.Cookie) (*websocket.Conn, error) {
	handshakeURL, err := toHandshakeURL(cfg.baseURL)
	if err != nil {
		return nil, err
	}
	handshakeURL = strings.TrimRight(handshakeURL, "/") + "/api/ws"

	cli, err := client.NewClient(client.WithDialer(standard.NewDialer()))
	if err != nil {
		return nil, fmt.Errorf("创建 hertz 客户端失败: %w", err)
	}

	req := protocol.AcquireRequest()
	resp := protocol.AcquireResponse()
	defer protocol.ReleaseRequest(req)

	req.SetRequestURI(handshakeURL)
	req.SetMethod(http.MethodGet)
	req.Header.Set("Cookie", cookie.Name+"="+cookie.Value)
	if cfg.insecureOrigin {
		req.Header.Set("Origin", strings.TrimRight(cfg.baseURL, "/"))
	}

	upgrader := websocket.ClientUpgrader{}
	upgrader.PrepareRequest(req)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	err = cli.Do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("建立 WebSocket 失败: %w", err)
	}

	conn, err := upgrader.UpgradeResponse(req, resp)
	if err != nil {
		return nil, fmt.Errorf("升级到 WebSocket 失败: %w", err)
	}
	return conn, nil
}

func connectWebSockets(cfg config, cookie *http.Cookie) ([]*websocket.Conn, error) {
	connections := make([]*websocket.Conn, 0, cfg.connections)
	for index := 0; index < cfg.connections; index++ {
		conn, err := connectWebSocket(cfg, cookie)
		if err != nil {
			closeWebSockets(connections)
			return nil, fmt.Errorf("建立第 %d 条 WebSocket 失败: %w", index+1, err)
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

func closeWebSockets(connections []*websocket.Conn) {
	for _, conn := range connections {
		if conn == nil {
			continue
		}
		_ = conn.Close()
	}
}

func toHandshakeURL(baseURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("解析 -base 失败: %w", err)
	}
	switch parsed.Scheme {
	case "ws":
		parsed.Scheme = "http"
	case "wss":
		parsed.Scheme = "https"
	case "http", "https":
	default:
		return "", fmt.Errorf("不支持的地址协议: %s", parsed.Scheme)
	}
	return parsed.String(), nil
}

func packClickRequest(slug string) ([]byte, error) {
	body, err := proto.Marshal(&realtimepb.ClickRequest{
		Slug: strings.TrimSpace(slug),
	})
	if err != nil {
		return nil, fmt.Errorf("编码点击请求失败: %w", err)
	}
	frame := make([]byte, 1+len(body))
	frame[0] = realtimeBinaryTypeClickRequest
	copy(frame[1:], body)
	return frame, nil
}

func unpackClickAck(frame []byte) (*realtimepb.ClickAck, error) {
	if len(frame) == 0 || frame[0] != realtimeBinaryTypeClickAck {
		return nil, errors.New("返回帧不是 click_ack")
	}
	message := &realtimepb.ClickAck{}
	if err := proto.Unmarshal(frame[1:], message); err != nil {
		return nil, err
	}
	return message, nil
}

func runConnectionWorker(connectionIndex int, conn *websocket.Conn, frame []byte, cfg config, startCh <-chan struct{}) workerResult {
	<-startCh

	latencies := make([]time.Duration, 0, cfg.count)
	for clickIndex := 0; clickIndex < cfg.count; clickIndex++ {
		if err := conn.SetWriteDeadline(time.Now().Add(cfg.timeout)); err != nil {
			return workerResult{
				connectionIndex: connectionIndex,
				latencies:       latencies,
				err:             fmt.Errorf("连接 %d 设置写超时失败: %w", connectionIndex, err),
			}
		}

		start := time.Now()
		if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
			return workerResult{
				connectionIndex: connectionIndex,
				latencies:       latencies,
				err:             fmt.Errorf("连接 %d 第 %d 次点击发送失败: %w", connectionIndex, clickIndex+1, err),
			}
		}

		latency, err := waitForClickAck(conn, cfg.timeout, start)
		if err != nil {
			return workerResult{
				connectionIndex: connectionIndex,
				latencies:       latencies,
				err:             fmt.Errorf("连接 %d 第 %d 次点击确认失败: %w", connectionIndex, clickIndex+1, err),
			}
		}
		latencies = append(latencies, latency)

		if cfg.pause > 0 {
			time.Sleep(cfg.pause)
		}
	}

	return workerResult{
		connectionIndex: connectionIndex,
		latencies:       latencies,
	}
}

func waitForClickAck(conn *websocket.Conn, timeout time.Duration, startedAt time.Time) (time.Duration, error) {
	deadline := time.Now().Add(timeout)
	for {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return 0, fmt.Errorf("设置读超时失败: %w", err)
		}
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		if _, ok, err := classifyServerFrame(messageType, payload); err != nil {
			return 0, err
		} else if ok {
			return time.Since(startedAt), nil
		}
	}
}

func classifyServerFrame(messageType int, payload []byte) (*realtimepb.ClickAck, bool, error) {
	switch messageType {
	case websocket.TextMessage:
		var message realtimeTextEnvelope
		if err := json.Unmarshal(payload, &message); err != nil {
			return nil, false, fmt.Errorf("解析文本推送失败: %w", err)
		}
		if strings.TrimSpace(message.Type) == "error" {
			if strings.TrimSpace(message.Message) != "" {
				return nil, false, errors.New(strings.TrimSpace(message.Message))
			}
			if strings.TrimSpace(message.Code) != "" {
				return nil, false, fmt.Errorf("服务端返回错误: %s", strings.TrimSpace(message.Code))
			}
			return nil, false, errors.New("服务端返回未知文本错误")
		}
		return nil, false, nil
	case websocket.BinaryMessage:
		if len(payload) == 0 || payload[0] != realtimeBinaryTypeClickAck {
			return nil, false, nil
		}
		ack, err := unpackClickAck(payload)
		if err != nil {
			return nil, false, fmt.Errorf("解析 click_ack 失败: %w", err)
		}
		return ack, true, nil
	default:
		return nil, false, fmt.Errorf("收到不支持的消息类型 %d", messageType)
	}
}

func summarizeLatencies(latencies []time.Duration, elapsed time.Duration) latencyStats {
	sorted := append([]time.Duration(nil), latencies...)
	slices.Sort(sorted)

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}

	stats := latencyStats{
		Min:     sorted[0],
		Max:     sorted[len(sorted)-1],
		Avg:     time.Duration(int64(total) / int64(len(sorted))),
		P50:     percentileDuration(sorted, 50),
		P95:     percentileDuration(sorted, 95),
		P99:     percentileDuration(sorted, 99),
		Elapsed: elapsed,
		QPS:     float64(len(sorted)) / elapsed.Seconds(),
	}
	return stats
}

func buildRunSummary(perConnectionLatencies [][]time.Duration, elapsed time.Duration) runSummary {
	perConnection := make([]latencyStats, 0, len(perConnectionLatencies))
	allLatencies := make([]time.Duration, 0)
	totalSamples := 0
	for _, latencies := range perConnectionLatencies {
		if len(latencies) == 0 {
			continue
		}
		perConnection = append(perConnection, summarizeLatencies(latencies, elapsed))
		allLatencies = append(allLatencies, latencies...)
		totalSamples += len(latencies)
	}

	summary := runSummary{
		perConnection: perConnection,
		totalSamples:  totalSamples,
	}
	if len(allLatencies) > 0 {
		summary.overall = summarizeLatencies(allLatencies, elapsed)
	}
	return summary
}

func percentileDuration(sorted []time.Duration, percentile float64) time.Duration {
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := int(math.Ceil((percentile / 100.0) * float64(len(sorted))))
	if position <= 0 {
		position = 1
	}
	if position > len(sorted) {
		position = len(sorted)
	}
	return sorted[position-1]
}

func printSummary(cfg config, nickname string, summary runSummary) {
	fmt.Printf("账号: %s\n", nickname)
	fmt.Printf("按钮: %s\n", cfg.slug)
	fmt.Printf("连接数: %d\n", cfg.connections)
	fmt.Printf("每连接样本数: %d\n", cfg.count)
	fmt.Printf("总样本数: %d\n", summary.totalSamples)
	fmt.Printf("总耗时: %s\n", summary.overall.Elapsed)
	fmt.Printf("整体平均吞吐: %.2f 次/秒\n", summary.overall.QPS)
	fmt.Printf("整体最小延迟: %s\n", summary.overall.Min)
	fmt.Printf("整体平均延迟: %s\n", summary.overall.Avg)
	fmt.Printf("整体 p50 延迟: %s\n", summary.overall.P50)
	fmt.Printf("整体 p95 延迟: %s\n", summary.overall.P95)
	fmt.Printf("整体 p99 延迟: %s\n", summary.overall.P99)
	fmt.Printf("整体最大延迟: %s\n", summary.overall.Max)
	for index, stats := range summary.perConnection {
		fmt.Printf("连接 %d: qps=%.2f avg=%s p95=%s p99=%s\n", index+1, stats.QPS, stats.Avg, stats.P95, stats.P99)
	}
}
