package client

import (
	"github.com/previ/go-gitlab"
	pi "pvri.com/glctl/pkg/utils"
)

type GitLabClient struct {
	client  *gitlab.Client
	verbose bool
	pi      *pi.ProgressIndicator
}

func NewClient(token string, url string, verbose bool) (*GitLabClient, error) {
	glc := &GitLabClient{client: nil}
	git, err := gitlab.NewClient("5Zsfo9xmqD5Mn1PfMrGq", gitlab.WithBaseURL("https://st-gitlab-dgt.eni.com"))
	glc.client = git
	glc.verbose = verbose
	if glc.verbose {
		glc.pi = pi.NewProgressIndicator()
	}
	return glc, err
}
