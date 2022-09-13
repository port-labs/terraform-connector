package port

import "time"

type (
	ActionBody struct {
		Action       string  `json:"action"`
		ResourceType string  `json:"resource_type"`
		Status       string  `json:"status"`
		Trigger      Trigger `json:"trigger"`
		Context      Context `json:"context"`
		Payload      Payload `json:"payload"`
	}

	Trigger struct {
		By     By        `json:"by"`
		At     time.Time `json:"at"`
		Origin string    `json:"origin"`
	}
	Context struct {
		Entity    Entity `json:"entity,omitempty"`
		Blueprint string `json:"blueprint"`
		RunID     string `json:"runId"`
	}
	Payload struct {
		Entity     Entity         `json:"entity,omitempty"`
		Action     Action         `json:"action,omitempty"`
		Properties map[string]any `json:"properties,omitempty"`
	}

	By struct {
		UserID string `json:"userId"`
		OrgID  string `json:"orgId"`
	}
	Entity struct {
		ID string `json:"id,omitempty"`
	}
	Action struct {
		ID               string     `json:"id"`
		Identifier       string     `json:"identifier"`
		Title            string     `json:"title"`
		UserInputs       UserInputs `json:"userInputs"`
		InvocationMethod string     `json:"invocationMethod"`
		Trigger          string     `json:"trigger"`
		Description      string     `json:"description"`
		Blueprint        string     `json:"blueprint"`
		CreatedAt        string     `json:"createdAt"`
		CreatedBy        string     `json:"createdBy"`
		UpdatedAt        string     `json:"updatedAt"`
		UpdatedBy        string     `json:"updatedBy"`
	}
	UserInputs struct {
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}
)
