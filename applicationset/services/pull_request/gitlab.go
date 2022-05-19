package pull_request

import (
	"context"
	"fmt"
	"os"

	"github.com/gosimple/slug"
	"github.com/xanzy/go-gitlab"
)

type GitlabService struct {
	client *gitlab.Client
	repo   string
	labels []string
	State  string
	Token  string
}

var _ PullRequestService = (*GitlabService)(nil)

func NewGitlabService(ctx context.Context, token, url, repo string, labels []string) (PullRequestService, error) {

	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}
	var client *gitlab.Client
	if url == "" {
		var err error
		client, err = gitlab.NewClient(token)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(url))
		if err != nil {
			return nil, err
		}
	}

	return &GitlabService{
		client: client,
		repo:   repo,
		labels: labels,
		Token:  token,
		State:  "opened",
	}, nil
}

func (g *GitlabService) List(ctx context.Context) ([]*PullRequest, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:  &g.State,
		Labels: &gitlab.Labels{g.labels[0]},
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	pullRequests := []*PullRequest{}

	for {

		p, _, err := g.client.Projects.GetProject(g.repo, nil)

		fmt.Printf("g.client: %s\n\n", g.Token)
		pulls, resp, err := g.client.MergeRequests.ListProjectMergeRequests(p.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing pull requests for %s: %v", g.repo, err)
		}

		for i, pull := range pulls {
			slug.CustomSub = map[string]string{
				"_": "-",
			}
			source_branch_slug := slug.Make(pull.SourceBranch)

			fmt.Printf("\n\nPull #%d, branch name: %d\n\n", i, source_branch_slug)

			pullRequests = append(pullRequests, &PullRequest{
				Number:  pull.IID,
				Branch:  source_branch_slug,
				HeadSHA: pull.SHA,
			})
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return pullRequests, nil
}

// // containLabels returns true if gotLabels contains expectedLabels
// func containGitlabLabels(expectedLabels []string, gotLabels []*github.Label) bool {
// 	for _, expected := range expectedLabels {
// 		found := false
// 		for _, got := range gotLabels {
// 			if got.Name == nil {
// 				continue
// 			}
// 			if expected == *got.Name {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
// 	return true
// }
