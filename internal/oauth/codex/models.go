package codex

import "charm.land/catwalk/pkg/catwalk"

func Models() []catwalk.Model {
	return []catwalk.Model{
		model("gpt-5.5", "GPT-5.5"),
		model("gpt-5.5-fast", "GPT-5.5 Fast"),
		model("gpt-5.4", "GPT-5.4"),
		model("gpt-5.4-mini", "GPT-5.4 Mini"),
		model("gpt-5.3-codex", "GPT-5.3 Codex"),
		model("gpt-5.3-codex-spark", "GPT-5.3 Codex Spark"),
	}
}

func model(id, name string) catwalk.Model {
	return catwalk.Model{
		ID:                     id,
		Name:                   name,
		ContextWindow:          400000,
		DefaultMaxTokens:       128000,
		CanReason:              true,
		ReasoningLevels:        []string{"low", "medium", "high", "xhigh"},
		DefaultReasoningEffort: "medium",
	}
}
