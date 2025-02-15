package shared

import (
	"fmt"
	"testing"
	"time"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/cmd/factory"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/cli/v2/pkg/search"
	"github.com/stretchr/testify/assert"
)

func TestSearcher(t *testing.T) {
	f := factory.New("1")
	f.Config = func() (config.Config, error) {
		return config.NewBlankConfig(), nil
	}
	_, err := Searcher(f)
	assert.NoError(t, err)
}

func TestSearchIssues(t *testing.T) {
	query := search.Query{
		Keywords: []string{"keyword"},
		Kind:     "issues",
		Limit:    30,
		Qualifiers: search.Qualifiers{
			Language: "go",
			Type:     "issue",
			Is:       []string{"public", "locked"},
		},
	}

	var updatedAt = time.Date(2021, 2, 28, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		errMsg     string
		name       string
		opts       *IssuesOptions
		tty        bool
		wantErr    bool
		wantStderr string
		wantStdout string
	}{
		{
			name: "displays results tty",
			opts: &IssuesOptions{
				Entity: Issues,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{
							IncompleteResults: false,
							Items: []search.Issue{
								{RepositoryURL: "github.com/test/cli", Number: 123, State: "open", Title: "something broken", Labels: []search.Label{{Name: "bug"}, {Name: "p1"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/what/what", Number: 456, State: "closed", Title: "feature request", Labels: []search.Label{{Name: "enhancement"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/blah/test", Number: 789, State: "open", Title: "some title", UpdatedAt: updatedAt},
							},
							Total: 300,
						}, nil
					},
				},
			},
			tty:        true,
			wantStdout: "\nShowing 3 of 300 issues\n\ntest/cli   #123  something broken  bug, p1      about 1 year ago\nwhat/what  #456  feature request   enhancement  about 1 year ago\nblah/test  #789  some title                     about 1 year ago\n",
		},
		{
			name: "displays issues and pull requests tty",
			opts: &IssuesOptions{
				Entity: Both,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{
							IncompleteResults: false,
							Items: []search.Issue{
								{RepositoryURL: "github.com/test/cli", Number: 123, State: "open", Title: "bug", Labels: []search.Label{{Name: "bug"}, {Name: "p1"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/what/what", Number: 456, State: "open", Title: "fix bug", Labels: []search.Label{{Name: "fix"}}, PullRequestLinks: search.PullRequestLinks{URL: "someurl"}, UpdatedAt: updatedAt},
							},
							Total: 300,
						}, nil
					},
				},
			},
			tty:        true,
			wantStdout: "\nShowing 2 of 300 issues and pull requests\n\nissue  test/cli   #123  bug      bug, p1  about 1 year ago\npr     what/what  #456  fix bug  fix      about 1 year ago\n",
		},
		{
			name: "displays no results tty",
			opts: &IssuesOptions{
				Entity: Issues,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{}, nil
					},
				},
			},
			tty:        true,
			wantStdout: "\nNo issues matched your search\n",
		},
		{
			name: "displays results notty",
			opts: &IssuesOptions{
				Entity: Issues,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{
							IncompleteResults: false,
							Items: []search.Issue{
								{RepositoryURL: "github.com/test/cli", Number: 123, State: "open", Title: "something broken", Labels: []search.Label{{Name: "bug"}, {Name: "p1"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/what/what", Number: 456, State: "closed", Title: "feature request", Labels: []search.Label{{Name: "enhancement"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/blah/test", Number: 789, State: "open", Title: "some title", UpdatedAt: updatedAt},
							},
							Total: 300,
						}, nil
					},
				},
			},
			wantStdout: "test/cli\t123\topen\tsomething broken\tbug, p1\t2021-02-28 12:30:00 +0000 UTC\nwhat/what\t456\tclosed\tfeature request\tenhancement\t2021-02-28 12:30:00 +0000 UTC\nblah/test\t789\topen\tsome title\t\t2021-02-28 12:30:00 +0000 UTC\n",
		},
		{
			name: "displays issues and pull requests notty",
			opts: &IssuesOptions{
				Entity: Both,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{
							IncompleteResults: false,
							Items: []search.Issue{
								{RepositoryURL: "github.com/test/cli", Number: 123, State: "open", Title: "bug", Labels: []search.Label{{Name: "bug"}, {Name: "p1"}}, UpdatedAt: updatedAt},
								{RepositoryURL: "github.com/what/what", Number: 456, State: "open", Title: "fix bug", Labels: []search.Label{{Name: "fix"}}, PullRequestLinks: search.PullRequestLinks{URL: "someurl"}, UpdatedAt: updatedAt},
							},
							Total: 300,
						}, nil
					},
				},
			},
			wantStdout: "issue\ttest/cli\t123\topen\tbug\tbug, p1\t2021-02-28 12:30:00 +0000 UTC\npr\twhat/what\t456\topen\tfix bug\tfix\t2021-02-28 12:30:00 +0000 UTC\n",
		},
		{
			name: "displays no results notty",
			opts: &IssuesOptions{
				Entity: Issues,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{}, nil
					},
				},
			},
		},
		{
			name: "displays search error",
			opts: &IssuesOptions{
				Entity: Issues,
				Query:  query,
				Searcher: &search.SearcherMock{
					IssuesFunc: func(query search.Query) (search.IssuesResult, error) {
						return search.IssuesResult{}, fmt.Errorf("error with query")
					},
				},
			},
			errMsg:  "error with query",
			wantErr: true,
		},
		{
			name: "opens browser for web mode tty",
			opts: &IssuesOptions{
				Browser: &cmdutil.TestBrowser{},
				Entity:  Issues,
				Query:   query,
				Searcher: &search.SearcherMock{
					URLFunc: func(query search.Query) string {
						return "https://github.com/search?type=issues&q=cli"
					},
				},
				WebMode: true,
			},
			tty:        true,
			wantStderr: "Opening github.com/search in your browser.\n",
		},
		{
			name: "opens browser for web mode notty",
			opts: &IssuesOptions{
				Browser: &cmdutil.TestBrowser{},
				Entity:  Issues,
				Query:   query,
				Searcher: &search.SearcherMock{
					URLFunc: func(query search.Query) string {
						return "https://github.com/search?type=issues&q=cli"
					},
				},
				WebMode: true,
			},
		},
	}
	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdinTTY(tt.tty)
		io.SetStdoutTTY(tt.tty)
		io.SetStderrTTY(tt.tty)
		tt.opts.IO = io
		t.Run(tt.name, func(t *testing.T) {
			err := SearchIssues(tt.opts)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
				return
			} else if err != nil {
				t.Fatalf("SearchIssues unexpected error: %v", err)
			}
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}
