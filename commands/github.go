package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ghGraphQLEndpoint = "https://api.github.com/graphql"

	querySAMLUsers = `query {
  organization(login: "%s") {
    samlIdentityProvider(ssoUrl: "%s") {
      externalIdentities() {
        edges {
          node {
            guid
            samlIdentity {
              nameId
            }
            user {
              login
            }
          }
        }
      }
    }
  }
}`

	queryUserRoles = `query {
  organization(login: "%s") {
    membersWithRole() {
      edges {
        node {
          login
        }
        role
      }
    }
  }
}`
)

type GitHubClient struct {
	client    *http.Client
	authToken string
}

func NewGitHubClient(authToken string) (*GitHubClient, error) {
	return &GitHubClient{
		client:    http.DefaultClient,
		authToken: authToken,
	}, nil
}

func (gh *GitHubClient) ListSAMLUsers(ctx context.Context, orgName, samlProvider string) (map[string]string, error) {
	q := fmt.Sprintf(querySAMLUsers, orgName, samlProvider)
	req, err := http.NewRequest(http.MethodPost, ghGraphQLEndpoint, strings.NewReader(q))
	if err != nil {
		return nil, fmt.Errorf("failed to build GraphQL request: %w", err)
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "bearer "+gh.authToken)

	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub GraphQL API: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GitHub GraphQL response: %w", err)
	}

	return parseSAMLUsers(b)
}

func (gh *GitHubClient) ListUserRoles(ctx context.Context, orgName string) (map[string]string, error) {
	q := fmt.Sprintf(queryUserRoles, orgName)
	req, err := http.NewRequest(http.MethodPost, ghGraphQLEndpoint, strings.NewReader(q))
	if err != nil {
		return nil, fmt.Errorf("failed to build GraphQL request: %w", err)
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "bearer "+gh.authToken)

	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub GraphQL API: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GitHub GraphQL response: %w", err)
	}

	return parseSAMLUsers(b)
}

func parseUserRoles(b []byte) (map[string]string, error) {
	result := map[string]string{}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user roles: %w", err)
	}

	data, ok := raw["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid user roles:\n%s", string(b))
	}

	org, ok := data["organization"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("organization doesn't exist in user roles:\n%s", string(b))
	}

	membersWithRole, ok := org["membersWithRole"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("membersWithRole doesn't exist in user roles:\n%s", string(b))
	}

	edges, ok := membersWithRole["edges"].([]any)
	if !ok {
		return nil, fmt.Errorf("edges doesn't exist in user roles:\n%s", string(b))
	}

	for _, edge := range edges {
		e, ok := edge.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error edge type in user roles:\n%s", string(b))
		}

		node, ok := e["node"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("node doesn't exist in user roles:\n%s", string(b))
		}

		userLogin, ok := node["login"].(string)
		if !ok {
			return nil, fmt.Errorf("login doesn't exist in user roles:\n%s", string(b))
		}

		role, ok := e["role"].(string)
		if !ok {
			return nil, fmt.Errorf("role doesn't exist in user roles:\n%s", string(b))
		}

		result[userLogin] = role
	}

	return result, nil
}

// Ugly and hacky. Better to use some GraphQL lib that gives strong types.
func parseSAMLUsers(b []byte) (map[string]string, error) {
	result := map[string]string{}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SAML users: %w", err)
	}

	data, ok := raw["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid SAML users:\n%s", string(b))
	}

	org, ok := data["organization"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("organization doesn't exist in SAML users:\n%s", string(b))
	}

	samlIdentityProvider, ok := org["samlIdentityProvider"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("samlIdentityProvider doesn't exist in SAML users:\n%s", string(b))
	}

	externalIdentities, ok := samlIdentityProvider["externalIdentities"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("externalIdentities doesn't exist in SAML users:\n%s", string(b))
	}

	edges, ok := externalIdentities["edges"].([]any)
	if !ok {
		return nil, fmt.Errorf("edges doesn't exist in SAML users:\n%s", string(b))
	}

	for _, edge := range edges {
		e, ok := edge.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error edge type in SAML users:\n%s", string(b))
		}

		node, ok := e["node"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("node doesn't exist in SAML users:\n%s", string(b))
		}

		samlIdentity, ok := node["samlIdentity"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("samlIdentity doesn't exist in SAML users:\n%s", string(b))
		}
		user, ok := node["user"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("user doesn't exist in SAML users:\n%s", string(b))
		}

		samlNameID, ok := samlIdentity["nameId"].(string)
		if !ok {
			return nil, fmt.Errorf("samlIdentity.nameId doesn't exist in SAML users:\n%s", string(b))
		}

		userLogin, ok := user["login"].(string)
		if !ok {
			return nil, fmt.Errorf("user.login doesn't exist in SAML users:\n%s", string(b))
		}

		result[userLogin] = samlNameID
	}

	return result, nil
}
