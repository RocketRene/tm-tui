package data

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type StackStatus string

const (
	StatusAll       StackStatus = "all"
	StatusUnhealthy StackStatus = "unhealthy"
	StatusHealthy   StackStatus = "healthy"
	StatusDrifted   StackStatus = "drifted"
	StatusFailed    StackStatus = "failed"
	StatusOK        StackStatus = "ok"
)

type DeploymentStatus string

const (
	DeploymentCanceled DeploymentStatus = "canceled"
	DeploymentFailed   DeploymentStatus = "failed"
	DeploymentOK       DeploymentStatus = "ok"
	DeploymentPending  DeploymentStatus = "pending"
	DeploymentRunning  DeploymentStatus = "running"
)

type DriftStatus string

const (
	DriftOK      DriftStatus = "ok"
	DriftDrifted DriftStatus = "drifted"
	DriftFailed  DriftStatus = "failed"
)

type StackData struct {
	StackID          int              `json:"stack_id"`
	Repository       string           `json:"repository"`
	Path             string           `json:"path"`
	DefaultBranch    string           `json:"default_branch"`
	MetaID           string           `json:"meta_id"`
	MetaName         string           `json:"meta_name"`
	MetaDescription  string           `json:"meta_description"`
	MetaTags         []string         `json:"meta_tags"`
	Status           StackStatus      `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	SeenAt           time.Time        `json:"seen_at"`
	DeploymentStatus DeploymentStatus `json:"deployment_status"`
	DriftStatus      DriftStatus      `json:"drift_status"`
	Draft            bool             `json:"draft"`
}

type StackAPIResponse struct {
	Stacks          []StackData `json:"stacks"`
	PaginatedResult struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"paginated_result"`
}

func (data StackData) GetTitle() string {
	return data.MetaName
}

//	func (data IssueData) GetRepoNameWithOwner() string {
//		return data.Repository.NameWithOwner
//	}
func (data StackData) GetRepoName() string {
	return data.Repository
}

func (data StackData) GetNumber() int {
	return data.StackID
}

func (data StackData) GetUrl() string {
	url := fmt.Sprintf("https://cloud.terramate.io/o/RocketRene/stacks/%d", data.StackID)
	return url
}

func (data StackData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue %s sort:updated", query)
}

func FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	var err error
	client, err := gh.DefaultGraphQLClient()
	if err != nil {
		return IssuesResponse{}, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Issue IssueData `graphql:"... on Issue"`
			}
			IssueCount int
			PageInfo   PageInfo
		} `graphql:"search(type: ISSUE, first: $limit, after: $endCursor, query: $query)"`
	}
	var endCursor *string
	if pageInfo != nil {
		endCursor = &pageInfo.EndCursor
	}
	variables := map[string]interface{}{
		"query":     graphql.String(makeIssuesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching issues", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return IssuesResponse{}, err
	}
	log.Debug("Successfully fetched issues", "query", query, "count", queryResult.Search.IssueCount)

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.Issue.Repository.IsArchived {
			continue
		}
		issues = append(issues, node.Issue)
	}

	return IssuesResponse{
		Issues:     issues,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

/* func fetchStacks(client *http.Client, token string) ([]Stack, error) {
	request, err := http.NewRequest("GET", "http://api.terramate.io/v1/stacks/5fbadfe9-b35b-4352-aadf-b03ee7a0a0c0", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResponse StackAPIResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	sort.Slice(apiResponse.Stacks, func(i, j int) bool {
		return apiResponse.Stacks[i].UpdatedAt.After(apiResponse.Stacks[j].UpdatedAt)
	})

	return apiResponse.Stacks, nil

}


// Auth
type CredentialData struct {
	IDToken string `json:"id_token"`
}

func LoadCredentials(filepath string) (string, error) {
	var creds CredentialData
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}
	return creds.IDToken, nil
}







*/
