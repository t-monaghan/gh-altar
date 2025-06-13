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

var brokerAddress string

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.Flags().StringVarP(&brokerAddress, "broker-address", "a", "http://127.0.0.1:25827/api/pipeline-watcher", "IP Address of your altar broker admin server")
}

var ciCmd = &cobra.Command{
	Use:     "watch-ci",
	Short:   "tell the altar broker to watch the github actions for this PR",
	Aliases: []string{"ci"},
	Run:     ci,
}

func ci(cmd *cobra.Command, args []string) {
	out, stderr, err := gh.Exec("pr", "status", "--json", "statusCheckRollup")
	if err != nil {
		slog.Error("error encountered querying gh cli for pr status", "error", err, "stderr", stderr.String())
		return
	}
	var response StatusCheckGroup
	err = json.Unmarshal(out.Bytes(), &response)
	if err != nil {
		slog.Error("failed to unmarshal github cli response into StatusCheckGroup", "error", err)
		return
	}
	// query how many actions there are
	totalActions := len(response.StatusCheckRollup)
	completedActions := 0
	failedActions := []string{}
	// query how many actions are in progress
	for _, action := range response.StatusCheckRollup {
		if action.Status == "COMPLETED" {
			completedActions += 1
			// query if any actions have failed
			if action.Conclusion == "FAILURE" {
				failedActions = append(failedActions, action.Name)
			}
		}
	}
	status := pipelinewatcher.PullRequestActionsStatus{
		CompletedActions: completedActions,
		TotalActions:     totalActions,
		FailedActions:    failedActions}

	jsonData, err := json.Marshal(status)
	if err != nil {
		slog.Error("failed to marshal github pr status into json", "error", err)
		return
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), "POST", brokerAddress, bufferedJSON)
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
	return
}
