package commands

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	tokenSrc  oauth2.TokenSource
	authToken string
}

func NewGitHubClient(authToken string) (*GitHubClient, error) {
	return &GitHubClient{
		tokenSrc: oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: authToken},
		),
		authToken: authToken,
	}, nil
}

func (gh *GitHubClient) ListSAMLUsers(ctx context.Context, orgName string) (map[string]string, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, gh.tokenSrc))

	var q struct {
		Organization struct {
			SamlIdentityProvider struct {
				ExternalIdentities struct {
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
					Edges []struct {
						Node struct {
							SamlIdentity struct {
								NameID string
							}
							User struct {
								Login string
							}
						}
					}
				} `graphql:"externalIdentities(first: 100, after: $cursor)"`
			}
		} `graphql:"organization(login: $orgName)"`
	}

	vars := map[string]any{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil), // Null after argument to get first page
	}

	result := map[string]string{}
	for {
		if err := client.Query(ctx, &q, vars); err != nil {
			return nil, fmt.Errorf("failed to query SAML users: %w", err)
		}

		for _, v := range q.Organization.SamlIdentityProvider.ExternalIdentities.Edges {
			if v.Node.User.Login != "" {
				result[v.Node.User.Login] = v.Node.SamlIdentity.NameID
			}
		}

		if !q.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = q.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo.EndCursor
	}

	return result, nil
}

func (gh *GitHubClient) ListUserRoles(ctx context.Context, orgName string) (map[string]string, error) {
	client := githubv4.NewClient(oauth2.NewClient(ctx, gh.tokenSrc))

	var q struct {
		Organization struct {
			MembersWithRole struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Node struct {
						Login string
					}
					Role string
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
	}

	vars := map[string]any{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil), // Null after argument to get first page
	}

	result := map[string]string{}
	for {
		if err := client.Query(ctx, &q, vars); err != nil {
			return nil, fmt.Errorf("failed to query user roles: %w", err)
		}

		for _, v := range q.Organization.MembersWithRole.Edges {
			if v.Node.Login != "" {
				result[v.Node.Login] = v.Role
			}
		}

		if !q.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = q.Organization.MembersWithRole.PageInfo.EndCursor
	}

	return result, nil
}
