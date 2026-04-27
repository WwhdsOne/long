package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

type talentLearnRequest struct {
	TalentID string `json:"talentId"`
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
		talentGroup.POST("/learn", h.learn)
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

	allTrees := map[string][]vote.TalentDef{
		"normal": vote.GetTreeTalents(vote.TalentTreeNormal),
		"armor":  vote.GetTreeTalents(vote.TalentTreeArmor),
		"crit":   vote.GetTreeTalents(vote.TalentTreeCrit),
	}

	userState, err := h.store.GetUserState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "user state fetch failed"})
		return
	}

	c.JSON(consts.StatusOK, map[string]any{
		"trees":        allTrees,
		"talents":      state.Talents,
		"talentPoints": userState.TalentPoints,
	})
}

func (h *talentAPI) learn(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	var req talentLearnRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if err := h.store.LearnTalent(ctx, nickStr, req.TalentID); err != nil {
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
				"talents":               defsToMap(vote.GetTreeTalents(vote.TalentTreeNormal)),
				"tierCompletionBonuses": vote.TalentTierCompletionBonusLabels(vote.TalentTreeNormal),
			},
			"armor": map[string]any{
				"name":                  "碎盾攻坚",
				"talents":               defsToMap(vote.GetTreeTalents(vote.TalentTreeArmor)),
				"tierCompletionBonuses": vote.TalentTierCompletionBonusLabels(vote.TalentTreeArmor),
			},
			"crit": map[string]any{
				"name":                  "致命洞察",
				"talents":               defsToMap(vote.GetTreeTalents(vote.TalentTreeCrit)),
				"tierCompletionBonuses": vote.TalentTierCompletionBonusLabels(vote.TalentTreeCrit),
			},
		},
	}
	c.JSON(consts.StatusOK, result)
}

func defsToMap(defs []vote.TalentDef) []map[string]any {
	result := make([]map[string]any, 0, len(defs))
	for _, d := range defs {
		result = append(result, map[string]any{
			"id":                d.ID,
			"tree":              d.Tree,
			"tier":              d.Tier,
			"name":              d.Name,
			"effectType":        d.EffectType,
			"effectValue":       d.EffectValue,
			"effectDescription": vote.TalentEffectDescription(d),
			"cost":              d.Cost,
			"prerequisite":      d.Prerequisite,
			"prerequisiteName":  vote.TalentPrerequisiteName(d),
		})
	}
	return result
}

func talentErrorStatus(err error) int {
	switch {
	case errors.Is(err, vote.ErrTalentNotFound):
		return consts.StatusNotFound
	case errors.Is(err, vote.ErrTalentPointsInsufficient),
		errors.Is(err, vote.ErrTalentPrerequisite),
		errors.Is(err, vote.ErrTalentAlreadyLearned),
		errors.Is(err, vote.ErrTalentInvalidCost),
		errors.Is(err, vote.ErrInvalidTalentTree):
		return consts.StatusBadRequest
	default:
		return consts.StatusInternalServerError
	}
}

func talentErrorCode(err error) string {
	switch {
	case errors.Is(err, vote.ErrTalentPointsInsufficient):
		return "TALENT_POINTS_INSUFFICIENT"
	case errors.Is(err, vote.ErrTalentInvalidCost):
		return "TALENT_INVALID_COST"
	case errors.Is(err, vote.ErrTalentPrerequisite):
		return "TALENT_PREREQUISITE_NOT_MET"
	case errors.Is(err, vote.ErrTalentAlreadyLearned):
		return "TALENT_ALREADY_LEARNED"
	case errors.Is(err, vote.ErrInvalidTalentTree):
		return "INVALID_TALENT_TREE"
	case errors.Is(err, vote.ErrTalentNotFound):
		return "TALENT_NOT_FOUND"
	default:
		return "TALENT_OPERATION_FAILED"
	}
}

func talentErrorMessage(err error) string {
	switch {
	case errors.Is(err, vote.ErrTalentPointsInsufficient):
		return "天赋点不足，无法学习该节点。"
	case errors.Is(err, vote.ErrTalentInvalidCost):
		return "天赋配置异常，节点成本无效。"
	case errors.Is(err, vote.ErrTalentPrerequisite):
		return "前置节点尚未学习。"
	case errors.Is(err, vote.ErrTalentAlreadyLearned):
		return "该天赋已学习。"
	case errors.Is(err, vote.ErrInvalidTalentTree):
		return "天赋树选择无效。"
	case errors.Is(err, vote.ErrTalentNotFound):
		return "未找到该天赋节点。"
	default:
		return "天赋操作失败，请稍后重试。"
	}
}
