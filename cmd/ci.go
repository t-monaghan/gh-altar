// Package cmd contains the command logic for gh-altar.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cli/go-gh/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/t-monaghan/altar/examples/pipelinewatcher"
)

const loopDelay = time.Second * 5

func init() {
	rootCmd.AddCommand(ciCmd)
}

var ciCmd = &cobra.Command{
	Use:     "watch-ci",
	Short:   "watches the required checks for the given working directory's PR and sends the action information to altar",
	Aliases: []string{"ci", "checks"},
	Run:     ci,
}

func ci(_ *cobra.Command, _ []string) {
	var failedActions []string

	completedActions := -1
	totalActions := 0

	for completedActions != totalActions {
		response, err := queryGH()
		if err != nil {
			slog.Error("failed to query github", "error", err)

			return
		}

		totalActions, completedActions, failedActions = getCheckInfo(response)
		status := pipelinewatcher.PullRequestActionsStatus{
			CompletedActions: completedActions,
			TotalActions:     totalActions,
			FailedActions:    failedActions,
		}

		err = sendRequest(status)
		if err != nil {
			slog.Error("error sending request to altar admin", "error", err)

			return
		}

		time.Sleep(loopDelay)
	}

	slog.Info("checks complete")
}

func sendRequest(status pipelinewatcher.PullRequestActionsStatus) error {
	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal json data into PullRequestActionStatus: %w", err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	address := viper.GetString("broker.address")

	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, address+"/api/pipeline-watcher", bufferedJSON)

	if err != nil {
		return fmt.Errorf("failed to marshal github pr status into json: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request to altar admin: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("altar admin responded to request with non-200 status: %v", resp.Status) //nolint:err113
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.Error("failed to close body of github state request", "error", closeErr)
		}
	}()

	return nil
}

// ErrGHCLIQuery represents an error querying GitHub CLI.
var ErrGHCLIQuery = errors.New("error encountered querying gh cli for pr check information")

func queryGH() ([]ChecksResult, error) {
	out, stderr, err := gh.Exec("pr", "checks", "--json", "name,bucket")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGHCLIQuery, stderr.String())
	}

	var response []ChecksResult
	err = json.Unmarshal(out.Bytes(), &response)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal github cli response into StatusCheckGroup: %w", err)
	}

	return response, nil
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
			completedActions++
		}
	}

	return totalActions, completedActions, failedActions
}

// ChecksResult represents the schema returned by the `gh pr checks` execution.
type ChecksResult struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
}
