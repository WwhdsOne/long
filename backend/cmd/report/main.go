package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/report"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("用法: go -C backend run ./cmd/report <daily|weekly|monthly|custom> [--from 2006-01-02] [--to 2006-01-02] [--out <dir>] [--no-md]")
	}

	sub := strings.ToLower(strings.TrimSpace(os.Args[1]))

	var fromStr, toStr, outDir string
	var noMd bool

	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.StringVar(&fromStr, "from", "", "起始日期 (2006-01-02)")
	fs.StringVar(&toStr, "to", "", "结束日期 (2006-01-02)")
	fs.StringVar(&outDir, "out", "reports", "输出目录")
	fs.BoolVar(&noMd, "no-md", false, "不写 .md 文件")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	now := time.Now()
	var from, to int64
	var title string

	switch sub {
	case "daily":
		from = dayStart(now.AddDate(0, 0, -1))
		to = dayStart(now)
		title = fmt.Sprintf("运营报表 — 日报 %s", time.Unix(from, 0).In(shanghaiLoc()).Format("2006-01-02"))
	case "weekly":
		from = weekStart(now.AddDate(0, 0, -7))
		to = weekStart(now)
		title = fmt.Sprintf("运营报表 — 周报 %s ~ %s",
			time.Unix(from, 0).In(shanghaiLoc()).Format("2006-01-02"),
			time.Unix(to, 0).In(shanghaiLoc()).AddDate(0, 0, -1).Format("2006-01-02"))
	case "monthly":
		thisMonth := monthStart(now)
		from = monthStart(time.Unix(thisMonth-1, 0).In(shanghaiLoc()))
		to = thisMonth
		title = fmt.Sprintf("运营报表 — 月报 %s", time.Unix(from, 0).In(shanghaiLoc()).Format("2006-01"))
	case "custom":
		if fromStr == "" || toStr == "" {
			return errors.New("custom 模式需要 --from 和 --to (格式 2006-01-02)")
		}
		var err error
		from, err = parseDate(fromStr)
		if err != nil {
			return fmt.Errorf("--from: %w", err)
		}
		to, err = parseDate(toStr)
		if err != nil {
			return fmt.Errorf("--to: %w", err)
		}
		if from >= to {
			return errors.New("--from 必须早于 --to")
		}
		title = fmt.Sprintf("运营报表 %s ~ %s", fromStr, toStr)
	default:
		return fmt.Errorf("未知子命令 %q，支持: daily, weekly, monthly, custom", sub)
	}

	// 从环境变量中读取配置
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE")
	mongoEnabled := os.Getenv("MONGO_ENABLED")

	if mongoEnabled != "true" {
		return errors.New("mongo.enabled=false，无法生成报表")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 连接 MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoURI).
		SetConnectTimeout(10*time.Second)) // 使用环境变量中的 URI 连接
	if err != nil {
		return fmt.Errorf("连接 Mongo: %w", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(mongoDatabase)

	fmt.Printf("正在查询 %s ~ %s ...\n\n",
		time.Unix(from, 0).In(shanghaiLoc()).Format("2006-01-02 15:04"),
		time.Unix(to, 0).In(shanghaiLoc()).Format("2006-01-02 15:04"))

	playerStats, err := report.QueryPlayerActivity(ctx, db, from, to)
	if err != nil {
		return err
	}

	bossStats, err := report.QueryBossStats(ctx, db, from, to)
	if err != nil {
		return err
	}

	economyStats, err := report.QueryEconomyStats(ctx, db, from, to)
	if err != nil {
		return err
	}

	summary := report.ReportSummary{
		Title:   title,
		From:    from,
		To:      to,
		Player:  playerStats,
		Boss:    bossStats,
		Economy: economyStats,
	}

	return report.Render(summary, outDir, noMd)
}

func shanghaiLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func dayStart(t time.Time) int64 {
	loc := shanghaiLoc()
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc).Unix()
}

func weekStart(t time.Time) int64 {
	loc := shanghaiLoc()
	tInLoc := t.In(loc)
	weekday := tInLoc.Weekday()
	daysSinceMonday := 0
	if weekday == time.Sunday {
		daysSinceMonday = 6
	} else {
		daysSinceMonday = int(weekday) - 1
	}
	mon := tInLoc.AddDate(0, 0, -daysSinceMonday)
	y, m, d := mon.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc).Unix()
}

func monthStart(t time.Time) int64 {
	loc := shanghaiLoc()
	y, m, _ := t.In(loc).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, loc).Unix()
}

func parseDate(s string) (int64, error) {
	loc := shanghaiLoc()
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(s), loc)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
