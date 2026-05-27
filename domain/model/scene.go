package model

type RuleSpec struct {
	Role        string `json:"role"`
	RuleKey     string `json:"ruleKey"`
	RuleVersion string `json:"ruleVersion"`
	Description string `json:"description,omitempty"`
}

type SceneSpec struct {
	SceneKey     string          `json:"sceneKey"`
	TargetType   string          `json:"targetType"`
	GenerateRule RuleSpec        `json:"generateRule"`
	ReviewRule   RuleSpec        `json:"reviewRule"`
	Capability   SceneCapability `json:"capability"`
}

type SceneCapability struct {
	CanAutoContinue bool `json:"canAutoContinue"`
	CanManualEdit   bool `json:"canManualEdit"`
	CanReviewOnly   bool `json:"canReviewOnly"`
	CanAdopt        bool `json:"canAdopt"`
	DefaultMaxDepth int  `json:"defaultMaxDepth"`
}
