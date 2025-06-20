package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/shurcooL/graphql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(contribCmd)
}

var contribCmd = &cobra.Command{
	Use:     "contributions",
	Short:   "sends the github contributions for the current user to altar",
	Aliases: []string{"grid"},
	Run:     grid,
}

type graphQlQuery struct {
	User struct {
		ContributionsCollection struct {
			ContributionCalendar struct {
				Weeks []struct {
					ContributionDays []struct {
						ContributionCount int
						Date              string
					}
				}
			}
		}
	} `graphql:"user(login: $userName)"`
}

func grid(_ *cobra.Command, _ []string) {
	var client *api.GraphQLClient

	var err error

	if GithubToken == "" {
		client, err = api.DefaultGraphQLClient()
	} else {
		client, err = api.NewGraphQLClient(api.ClientOptions{AuthToken: GithubToken})
	}

	if err != nil {
		slog.Error("failed to instantiate new graphql client", "error", err)

		return
	}

	userName, stderr, err := gh.Exec("api", "user", "--jq", ".login")
	if err != nil {
		slog.Error("failed to query github cli for current user", "error", stderr.String())

		return
	}

	variables := map[string]any{
		"userName": graphql.String(strings.TrimSpace(userName.String())),
	}

	var query graphQlQuery
	err = client.Query("foobarbaz", &query, variables)

	if err != nil {
		slog.Error("failed to create query for contributions", "error", err)
	}

	err = postToAltar(extractRawContributionCounts(query))
	if err != nil {
		slog.Error("error posting contribution counts to altar", "error", err)

		return
	}

	slog.Info("successfully posted grid to altar")
}

func extractRawContributionCounts(response graphQlQuery) []int {
	rawCount := []int{}

	for _, week := range response.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			rawCount = append(rawCount, day.ContributionCount)
		}
	}

	return rawCount
}

func postToAltar(contributions []int) error {
	jsonData, err := json.Marshal(contributions)
	if err != nil {
		return fmt.Errorf("failed to marshal json data into PullRequestActionStatus: %w", err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	address := viper.GetString("broker.address")

	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, address+"/api/contributions", bufferedJSON)

	if err != nil {
		return fmt.Errorf("failed to marshal list of ints as json: %w", err)
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
