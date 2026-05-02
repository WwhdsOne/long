package core

import (
	"context"
	"errors"
	"testing"
)

type fakeLuaRunner struct {
	loadCalls int
	evalCalls []string
	loadSHA   string
	evalValue any
	evalErrs  []error
}

func (f *fakeLuaRunner) EvalSha(_ context.Context, sha string, _ []string, _ ...any) (any, error) {
	f.evalCalls = append(f.evalCalls, sha)
	if len(f.evalErrs) > 0 {
		err := f.evalErrs[0]
		f.evalErrs = f.evalErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	return f.evalValue, nil
}

func (f *fakeLuaRunner) ScriptLoad(_ context.Context, _ string) (string, error) {
	f.loadCalls++
	return f.loadSHA, nil
}

func TestCachedLuaScriptLoadsShaOnceAndReusesIt(t *testing.T) {
	runner := &fakeLuaRunner{
		loadSHA:   "sha-1",
		evalValue: []any{int64(1)},
	}
	cache := newLuaScriptCache()
	script := newCachedLuaScript("boss-click", "return 1", cache)

	ctx := context.Background()
	if _, err := script.Run(ctx, runner, []string{"k1"}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if _, err := script.Run(ctx, runner, []string{"k1"}); err != nil {
		t.Fatalf("second run: %v", err)
	}

	if runner.loadCalls != 1 {
		t.Fatalf("expected one script load, got %d", runner.loadCalls)
	}
	if len(runner.evalCalls) != 2 || runner.evalCalls[0] != "sha-1" || runner.evalCalls[1] != "sha-1" {
		t.Fatalf("expected both evals to use cached sha, got %+v", runner.evalCalls)
	}
}

func TestCachedLuaScriptReloadsOnNoScript(t *testing.T) {
	runner := &fakeLuaRunner{
		loadSHA:   "sha-2",
		evalValue: []any{int64(1)},
		evalErrs: []error{
			errors.New("NOSCRIPT No matching script. Please use EVAL."),
			nil,
		},
	}
	cache := newLuaScriptCache()
	script := newCachedLuaScript("boss-click", "return 1", cache)
	cache.set("boss-click", "stale-sha")

	ctx := context.Background()
	if _, err := script.Run(ctx, runner, []string{"k1"}); err != nil {
		t.Fatalf("run with noscript reload: %v", err)
	}

	if runner.loadCalls != 1 {
		t.Fatalf("expected one reload after NOSCRIPT, got %d", runner.loadCalls)
	}
	if len(runner.evalCalls) != 2 || runner.evalCalls[0] != "stale-sha" || runner.evalCalls[1] != "sha-2" {
		t.Fatalf("unexpected eval sha sequence: %+v", runner.evalCalls)
	}
}
