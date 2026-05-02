package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Render 将报表输出到终端和 .md 文件。
func Render(r ReportSummary, outDir string, noMd bool) error {
	md := buildMarkdown(r)

	fmt.Print(md)

	if noMd || outDir == "" {
		return nil
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录 %s: %w", outDir, err)
	}

	filename := buildFilename(r)
	path := filepath.Join(outDir, filename)
	if err := os.WriteFile(path, []byte(md), 0644); err != nil {
		return fmt.Errorf("写入 %s: %w", path, err)
	}

	fmt.Printf("\n报表已保存: %s\n", path)
	return nil
}

func buildFilename(r ReportSummary) string {
	t := time.Unix(r.From, 0).In(shanghaiLocation())
	return fmt.Sprintf("report-%s.md", t.Format("2006-01-02"))
}

func buildMarkdown(r ReportSummary) string {
	var sb strings.Builder

	// 标题
	title := r.Title
	if title == "" {
		fromTime := time.Unix(r.From, 0).In(shanghaiLocation())
		toTime := time.Unix(r.To, 0).In(shanghaiLocation())
		title = fmt.Sprintf("运营报表 %s ~ %s", fromTime.Format("2006-01-02"), toTime.Format("2006-01-02"))
	}
	sb.WriteString("# " + title + "\n\n")

	// 一、玩家活跃
	writeSection(&sb, "一、玩家活跃", func() {
		writeKVTable(&sb, [][2]string{
			{"独立访问用户数", formatNum(r.Player.UniqueIPs)},
			{"活跃玩家数", formatNum(r.Player.ActivePlayers)},
			{"新增玩家数", formatNum(r.Player.NewPlayers)},
			{"总请求量", formatNum(r.Player.TotalRequests)},
			{"P95 延迟", fmt.Sprintf("%.1f ms", r.Player.P95LatencyMs)},
		})

		if len(r.Player.TopPlayers) > 0 {
			sb.WriteString("\n**活跃玩家 Top 10**\n\n")
			sb.WriteString("| 排名 | 玩家 | 操作次数 |\n")
			sb.WriteString("|------|------|----------|\n")
			for i, p := range r.Player.TopPlayers {
				sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, p.Nickname, formatNum(p.EventCount)))
			}
			sb.WriteString("\n")
		}
	})

	// 二、Boss 战况
	writeSection(&sb, "二、Boss 战况", func() {
		writeKVTable(&sb, [][2]string{
			{"Boss 生成次数", formatNum(r.Boss.SpawnCount)},
			{"Boss 击杀次数", formatNum(r.Boss.KillCount)},
			{"击杀率", fmt.Sprintf("%.1f%%", r.Boss.KillRate)},
			{"总伤害量", formatNum(r.Boss.TotalDamage)},
			{"平均存活时间", fmt.Sprintf("%.0f 秒", r.Boss.AvgSurvivalSecs)},
		})

		if len(r.Boss.TopDamagers) > 0 {
			sb.WriteString("\n**伤害榜 Top 10**\n\n")
			sb.WriteString("| 排名 | 玩家 | 总伤害 |\n")
			sb.WriteString("|------|------|--------|\n")
			for i, d := range r.Boss.TopDamagers {
				sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, d.Nickname, formatNum(d.TotalDamage)))
			}
			sb.WriteString("\n")
		}

		if len(r.Boss.TopLoot) > 0 {
			sb.WriteString("\n**掉落 Top 5**\n\n")
			sb.WriteString("| 排名 | 装备 | 掉落次数 |\n")
			sb.WriteString("|------|------|----------|\n")
			for i, l := range r.Boss.TopLoot {
				sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, l.ItemName, formatNum(l.Count)))
			}
			sb.WriteString("\n")
		}
	})

	// 三、经济系统
	writeSection(&sb, "三、经济系统", func() {
		sb.WriteString("**商店**\n\n")
		writeKVTable(&sb, [][2]string{
			{"商店总销售额（金币）", formatNum(r.Economy.ShopTotalGold)},
			{"商店购买次数", formatNum(r.Economy.ShopPurchaseCnt)},
		})

		if len(r.Economy.TopShopItems) > 0 {
			sb.WriteString("\n**热销商品 Top 10**\n\n")
			sb.WriteString("| 排名 | 商品ID | 购买次数 |\n")
			sb.WriteString("|------|--------|----------|\n")
			for i, item := range r.Economy.TopShopItems {
				sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, item.ItemID, formatNum(item.Count)))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("**任务**\n\n")
		writeKVTable(&sb, [][2]string{
			{"任务完成数", formatNum(r.Economy.TaskClaimCnt)},
			{"任务参与人数", formatNum(r.Economy.TaskParticipants)},
			{"任务奖励金币", formatNum(r.Economy.TaskRewardGold)},
			{"任务奖励石头", formatNum(r.Economy.TaskRewardStones)},
			{"任务奖励天赋点", formatNum(r.Economy.TaskRewardTP)},
		})
	})

	return sb.String()
}

func writeSection(sb *strings.Builder, title string, body func()) {
	sb.WriteString("## " + title + "\n\n")
	body()
}

func writeKVTable(sb *strings.Builder, rows [][2]string) {
	if len(rows) == 0 {
		return
	}
	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("|------|------|\n")
	for _, row := range rows {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", row[0], row[1]))
	}
	sb.WriteString("\n")
}

// formatNum 千分位格式化数字。
func formatNum(n int64) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatNum(-n)
	}

	var parts []int64
	for n > 0 {
		parts = append([]int64{n % 1000}, parts...)
		n /= 1000
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d", parts[0]))
	for i := 1; i < len(parts); i++ {
		result.WriteString(fmt.Sprintf(",%03d", parts[i]))
	}
	return result.String()
}
