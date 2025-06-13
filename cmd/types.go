package cmd

import (
	"time"
)

type StatusResponse struct {
	CreatedBy   []StatusCheckGroup `json:"createdBy"`
	NeedsReview []interface{}      `json:"needsReview"`
}

type StatusCheckGroup struct {
	StatusCheckRollup []CheckRun `json:"statusCheckRollup"`
}

type CheckRun struct {
	TypeName     string    `json:"__typename"`
	CompletedAt  time.Time `json:"completedAt"`
	Conclusion   string    `json:"conclusion"`
	DetailsURL   string    `json:"detailsUrl"`
	Name         string    `json:"name"`
	StartedAt    time.Time `json:"startedAt"`
	Status       string    `json:"status"`
	WorkflowName string    `json:"workflowName"`
}
