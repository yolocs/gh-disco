package commands

import (
	"testing"

	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestParseSAMLUsers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string
	}{{
		name: "success",
		input: `{
  "data": {
    "organization": {
      "samlIdentityProvider": {
        "ssoUrl": "https://accounts.google.com/o/saml2/idp?idpid=example",
        "externalIdentities": {
          "edges": [
            {
              "node": {
                "samlIdentity": {
                  "nameId": "user1@example.com"
                },
                "user": {
                  "login": "user1"
                }
              }
            },
            {
              "node": {
                "samlIdentity": {
                  "nameId": "user2@example.com"
                },
                "user": {
                  "login": "user2"
                }
              }
            }
          ]
        }
      }
    }
  }
}`,
		want: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseSAMLUsers([]byte(tc.input))
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SAML users (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestParseUserRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string
	}{{
		name: "success",
		input: `{
  "data": {
    "organization": {
      "membersWithRole": {
        "edges": [
          {
            "node": {
              "login": "user1"
            },
            "role": "MEMBER"
          },
          {
            "node": {
              "login": "user2"
            },
            "role": "ADMIN"
          }
        ]
      }
    }
  }
}`,
		want: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
		},
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseUserRoles([]byte(tc.input))
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("User roles (-want,+got):\n%s", diff)
			}
		})
	}
}
