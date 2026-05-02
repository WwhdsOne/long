package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

type talentLearnRequest struct {
	TalentID    string `json:"talentId"`
	TargetLevel int    `json:"targetLevel"`
}

type talentAPI struct {
	store ButtonStore
	auth  PlayerAuthenticator
}

func registerTalentRoutes(router route.IRouter, options Options) {
	h := &talentAPI{
		store: options.Store,
		auth:  options.PlayerAuthenticator,
	}

	talentGroup := router.Group("/api/talents")
	talentGroup.Use(requireTalentAuth(h.auth))
	{
		talentGroup.GET("/state", h.getState)
		talentGroup.POST("/upgrade", h.upgrade)
		talentGroup.POST("/reset", h.reset)
		talentGroup.GET("/defs", h.getDefs)
	}
}

func requireTalentAuth(auth PlayerAuthenticator) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, auth)
		if !ok {
			c.Abort()
			return
		}
		c.Set("nickname", nickname)
	}
}

func (h *talentAPI) getState(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	state, err := h.store.GetTalentState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	allTrees := map[string][]core.TalentDef{
		"normal": core.GetTreeTalents(core.TalentTreeNormal),
		"armor":  core.GetTreeTalents(core.TalentTreeArmor),
		"crit":   core.GetTreeTalents(core.TalentTreeCrit),
	}

	userState, err := h.store.GetUserState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "user state fetch failed"})
		return
	}

	c.JSON(consts.StatusOK, map[string]any{
		"trees":              allTrees,
		"talents":            state.Talents,
		"talentPoints":       userState.TalentPoints,
		"effectLines":        core.BuildTalentEffectLineMap(state),
		"effectDescriptions": core.BuildTalentEffectDescriptionMap(state),
	})
}

func (h *talentAPI) upgrade(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	var req talentLearnRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	targetLevel := max(req.TargetLevel, 1)
	if err := h.store.UpgradeTalent(ctx, nickStr, req.TalentID, targetLevel); err != nil {
		c.JSON(talentErrorStatus(err), map[string]string{
			"error":   talentErrorCode(err),
			"message": talentErrorMessage(err),
		})
		return
	}

	state, err := h.store.GetTalentState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "talent state fetch failed"})
		return
	}
	userState, err := h.store.GetUserState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "user state fetch failed"})
		return
	}
	c.JSON(consts.StatusOK, map[string]any{
		"status":             "ok",
		"talents":            state.Talents,
		"talentPoints":       userState.TalentPoints,
		"effectLines":        core.BuildTalentEffectLineMap(state),
		"effectDescriptions": core.BuildTalentEffectDescriptionMap(state),
	})
}

func (h *talentAPI) reset(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	if err := h.store.ResetTalents(ctx, nickStr); err != nil {
		c.JSON(talentErrorStatus(err), map[string]string{
			"error":   talentErrorCode(err),
			"message": talentErrorMessage(err),
		})
		return
	}

	userState, err := h.store.GetUserState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "user state fetch failed"})
		return
	}
	c.JSON(consts.StatusOK, map[string]any{
		"status":       "ok",
		"talentPoints": userState.TalentPoints,
	})
}

func (h *talentAPI) getDefs(ctx context.Context, c *app.RequestContext) {
	result := map[string]any{
		"trees": map[string]any{
			"normal": map[string]any{
				"name":                  "均衡攻势",
				"talents":               defsToMap(core.GetTreeTalents(core.TalentTreeNormal)),
				"tierCompletionBonuses": core.TalentTierCompletionBonusLabels(core.TalentTreeNormal),
			},
			"armor": map[string]any{
				"name":                  "碎盾攻坚",
				"talents":               defsToMap(core.GetTreeTalents(core.TalentTreeArmor)),
				"tierCompletionBonuses": core.TalentTierCompletionBonusLabels(core.TalentTreeArmor),
			},
			"crit": map[string]any{
				"name":                  "致命洞察",
				"talents":               defsToMap(core.GetTreeTalents(core.TalentTreeCrit)),
				"tierCompletionBonuses": core.TalentTierCompletionBonusLabels(core.TalentTreeCrit),
			},
		},
	}
	c.JSON(consts.StatusOK, result)
}

func defsToMap(defs []core.TalentDef) []map[string]any {
	result := make([]map[string]any, 0, len(defs))
	for _, d := range defs {
		result = append(result, map[string]any{
			"id":                d.ID,
			"tree":              d.Tree,
			"tier":              d.Tier,
			"name":              d.Name,
			"effectType":        d.EffectType,
			"effectValue":       d.EffectValue,
			"effectDescription": core.TalentEffectDescription(d),
			"maxLevel":          d.MaxLevel,
			"cost":              d.Cost,
		})
	}
	return result
}

func talentErrorStatus(err error) int {
	switch {
	case errors.Is(err, core.ErrTalentNotFound):
		return consts.StatusNotFound
	case errors.Is(err, core.ErrTalentPointsInsufficient),
		errors.Is(err, core.ErrTalentTierLocked),
		errors.Is(err, core.ErrTalentAlreadyLearned),
		errors.Is(err, core.ErrTalentInvalidCost),
		errors.Is(err, core.ErrTalentMaxLevel),
		errors.Is(err, core.ErrTalentInvalidLevel),
		errors.Is(err, core.ErrInvalidTalentTree):
		return consts.StatusBadRequest
	default:
		return consts.StatusInternalServerError
	}
}

func talentErrorCode(err error) string {
	switch {
	case errors.Is(err, core.ErrTalentPointsInsufficient):
		return "TALENT_POINTS_INSUFFICIENT"
	case errors.Is(err, core.ErrTalentInvalidCost):
		return "TALENT_INVALID_COST"
	case errors.Is(err, core.ErrTalentTierLocked):
		return "TALENT_TIER_LOCKED"
	case errors.Is(err, core.ErrTalentAlreadyLearned):
		return "TALENT_ALREADY_LEARNED"
	case errors.Is(err, core.ErrInvalidTalentTree):
		return "INVALID_TALENT_TREE"
	case errors.Is(err, core.ErrTalentNotFound):
		return "TALENT_NOT_FOUND"
	case errors.Is(err, core.ErrTalentMaxLevel):
		return "TALENT_MAX_LEVEL"
	case errors.Is(err, core.ErrTalentInvalidLevel):
		return "TALENT_INVALID_LEVEL"
	default:
		return "TALENT_OPERATION_FAILED"
	}
}

func talentErrorMessage(err error) string {
	switch {
	case errors.Is(err, core.ErrTalentPointsInsufficient):
		return "天赋点不足，无法学习该节点。"
	case errors.Is(err, core.ErrTalentInvalidCost):
		return "天赋配置异常，节点成本无效。"
	case errors.Is(err, core.ErrTalentTierLocked):
		return "上一层天赋尚未点满。"
	case errors.Is(err, core.ErrTalentAlreadyLearned):
		return "该天赋已学习。"
	case errors.Is(err, core.ErrInvalidTalentTree):
		return "天赋树选择无效。"
	case errors.Is(err, core.ErrTalentNotFound):
		return "未找到该天赋节点。"
	case errors.Is(err, core.ErrTalentMaxLevel):
		return "已达到最大等级。"
	case errors.Is(err, core.ErrTalentInvalidLevel):
		return "无效的等级参数。"
	default:
		return "天赋操作失败，请稍后重试。"
	}
}
