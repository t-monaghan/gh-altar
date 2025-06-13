// Package cmd contains the command logic for gh-altar.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
	"github.com/t-monaghan/altar/examples/pipelinewatcher"
)

var brokerAddress string //nolint:gochecknoglobals

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.Flags().StringVarP(&brokerAddress, "broker-address",
		"a", "http://127.0.0.1:25827/api/pipeline-watcher", "IP Address of your altar broker admin server")
}

var ciCmd = &cobra.Command{
	Use:     "watch-ci",
	Short:   "watches the required checks for the given working directory's PR and sends the action information to altar",
	Aliases: []string{"ci"},
	Run:     ci,
}

func ci(_ *cobra.Command, _ []string) {
	out, stderr, err := gh.Exec("pr", "checks", "--json", "name,bucket")
	if err != nil {
		slog.Error("error encountered querying gh cli for pr check information", "error", err, "stderr", stderr.String())

		return
	}

	var response []ChecksResult
	err = json.Unmarshal(out.Bytes(), &response)

	if err != nil {
		slog.Error("failed to unmarshal github cli response into StatusCheckGroup", "error", err)

		return
	}

	totalActions, completedActions, failedActions := getCheckInfo(response)
	status := pipelinewatcher.PullRequestActionsStatus{
		CompletedActions: completedActions,
		TotalActions:     totalActions,
		FailedActions:    failedActions,
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		slog.Error("failed to marshal github pr status into json", "error", err)

		return
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, brokerAddress, bufferedJSON)

	if err != nil {
		slog.Error("failed to marshal github pr status into json", "error", err)

		return
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to perform request to altar admin", "error", err)

		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("altar admin responded to request with non-200 status", "status", resp.Status)

		return
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.Error("failed to close body of github state request", "error", closeErr)
		}
	}()
}

func getCheckInfo(response []ChecksResult) (int, int, []string) {
	totalActions := len(response)
	completedActions := 0
	failedActions := []string{}

	for _, action := range response {
		switch action.Bucket {
		case "pass":
			completedActions++
		case "fail", "skipping", "cancel":
			failedActions = append(failedActions, action.Name)
		}
	}

	return totalActions, completedActions, failedActions
}

// ChecksResult represents the schema returned by the `gh pr checks` execution.
type ChecksResult struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
}
