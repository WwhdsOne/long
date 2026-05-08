package core

import (
	"github.com/bytedance/sonic"
	"github.com/vmihailenco/msgpack/v5"
)

func encodeBossParts(parts []BossPart) ([]byte, error) {
	if parts == nil {
		parts = []BossPart{}
	}
	return msgpack.Marshal(parts)
}

func decodeBossParts(raw []byte) ([]BossPart, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var parts []BossPart
	if err := msgpack.Unmarshal(raw, &parts); err == nil {
		return parts, nil
	}
	if err := sonic.Unmarshal(raw, &parts); err != nil {
		return nil, err
	}
	return parts, nil
}

func normalizeTalentCombatState(state *TalentCombatState) *TalentCombatState {
	if state == nil {
		return NewTalentCombatState()
	}
	if state.JudgmentDayUsed == nil {
		state.JudgmentDayUsed = make(map[string]int64)
	}
	if state.PartHeavyClickCount == nil {
		state.PartHeavyClickCount = make(map[string]int64)
	}
	if state.PartJudgmentDayCount == nil {
		state.PartJudgmentDayCount = make(map[string]int64)
	}
	if state.PartStormComboCount == nil {
		state.PartStormComboCount = make(map[string]int64)
	}
	if state.PartRetainedClicks == nil {
		state.PartRetainedClicks = make(map[string]int64)
	}
	if state.Bleeds == nil {
		state.Bleeds = make(map[string]TalentBleedState)
	}
	if state.SkinnerParts == nil {
		state.SkinnerParts = make(map[string]int64)
	}
	if state.SkinnerDurationByPart == nil {
		state.SkinnerDurationByPart = make(map[string]int64)
	}
	if state.DoomMarkCumDamage == nil {
		state.DoomMarkCumDamage = make(map[string]int64)
	}
	if state.PartMagicTriggerCount == nil {
		state.PartMagicTriggerCount = make(map[string]int64)
	}
	return state
}

func encodeTalentCombatState(state *TalentCombatState) ([]byte, error) {
	return msgpack.Marshal(normalizeTalentCombatState(state))
}

func decodeTalentCombatState(raw []byte) (*TalentCombatState, error) {
	state := NewTalentCombatState()
	if err := msgpack.Unmarshal(raw, state); err == nil {
		return normalizeTalentCombatState(state), nil
	}

	state = NewTalentCombatState()
	if err := sonic.Unmarshal(raw, state); err != nil {
		return nil, err
	}
	return normalizeTalentCombatState(state), nil
}
