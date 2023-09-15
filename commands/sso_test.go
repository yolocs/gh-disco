package commands

import (
	"strings"
	"testing"
)

func TestPrintExceptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		userRoles map[string]string
		samlUsers map[string]string
		limit     int
		want      string
	}{{
		name: "success",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
			"user3": "MEMBER",
		},
		samlUsers: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
		want: `+-------------------+--------+
| LOGIN WITHOUT SSO |  ROLE  |
+-------------------+--------+
| user3             | MEMBER |
+-------------------+--------+
`,
	}, {
		name: "success_with_limit",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
			"user3": "MEMBER",
			"user4": "MEMBER",
		},
		samlUsers: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
		limit: 1,
		want: `+-------------------+--------+
| LOGIN WITHOUT SSO |  ROLE  |
+-------------------+--------+
| user3             | MEMBER |
+-------------------+--------+
`,
	}, {
		name: "empty",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
		},
		samlUsers: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
		limit: 1,
		want: `+-------------------+------+
| LOGIN WITHOUT SSO | ROLE |
+-------------------+------+
+-------------------+------+
`,
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			printExceptions(&buf, tc.userRoles, tc.samlUsers, tc.limit)

			if got, want := buf.String(), tc.want; got != want {
				t.Errorf("printExceptions got=%s want=%s", got, want)
			}
		})
	}
}

func TestPrintSAMLUsers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		userRoles map[string]string
		samlUsers map[string]string
		limit     int
		want      string
	}{{
		name: "success",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
			"user3": "MEMBER",
		},
		samlUsers: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
		want: `+-------+--------+-------------------+
| LOGIN |  ROLE  |   SSO IDENTITY    |
+-------+--------+-------------------+
| user1 | MEMBER | user1@example.com |
| user2 | ADMIN  | user2@example.com |
+-------+--------+-------------------+
`,
	}, {
		name: "success_with_limit",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
			"user3": "MEMBER",
			"user4": "MEMBER",
		},
		samlUsers: map[string]string{
			"user1": "user1@example.com",
			"user2": "user2@example.com",
		},
		limit: 1,
		want: `+-------+--------+-------------------+
| LOGIN |  ROLE  |   SSO IDENTITY    |
+-------+--------+-------------------+
| user1 | MEMBER | user1@example.com |
+-------+--------+-------------------+
`,
	}, {
		name: "empty",
		userRoles: map[string]string{
			"user1": "MEMBER",
			"user2": "ADMIN",
		},
		samlUsers: map[string]string{},
		limit:     1,
		want: `+-------+------+--------------+
| LOGIN | ROLE | SSO IDENTITY |
+-------+------+--------------+
+-------+------+--------------+
`,
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			printSAMLUsers(&buf, tc.userRoles, tc.samlUsers, tc.limit)

			if got, want := buf.String(), tc.want; got != want {
				t.Errorf("printSAMLUsers got=%s want=%s", got, want)
			}
		})
	}
}
