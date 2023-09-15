package commands

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/sets"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type SSOCommand struct {
	cli.BaseCommand

	commonFlags *CommonFlags

	flagOrg            string
	flagSAMLProvider   string
	flagListExceptions bool
	flagLimit          int
}

func (c *SSOCommand) Desc() string {
	return "Query GitHub SSO status"
}

func (c *SSOCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Find SSO exceptions in the org:
	{{ COMMAND }} -org "my-org" -saml-provider="https://accounts.google.com/o/saml2/idp?idpid=example" -exceptions
`
}

func (c *SSOCommand) Flags() *cli.FlagSet {
	set := cli.NewFlagSet()

	c.commonFlags = &CommonFlags{}
	c.commonFlags.Register(set)

	f := set.NewSection("SSO FLAGS")

	f.StringVar(&cli.StringVar{
		Name:    "org",
		Target:  &c.flagOrg,
		Example: "my-org",
		EnvVar:  "GHDISCO_ORG",
		Usage:   "GitHub organization name.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "saml-provider",
		Target:  &c.flagSAMLProvider,
		Example: "https://accounts.google.com/o/saml2/idp?idpid=example",
		Usage:   "The SAML provider URL.",
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "exceptions",
		Target:  &c.flagListExceptions,
		Default: false,
		Usage:   "Whether to list SSO expcetions.",
	})

	f.IntVar(&cli.IntVar{
		Name:    "limit",
		Target:  &c.flagLimit,
		Aliases: []string{"n"},
		Default: 0,
		Usage:   "Limit the number of result. Setting to 0 will return all findings.",
	})

	set.AfterParse(func(existingErr error) error {
		err := existingErr
		if c.commonFlags.AuthToken == "" {
			err = errors.Join(existingErr, fmt.Errorf("missing -auth-token"))
		}
		if c.flagOrg == "" {
			err = errors.Join(existingErr, fmt.Errorf("missing -org"))
		}
		if c.flagSAMLProvider == "" {
			err = errors.Join(existingErr, fmt.Errorf("missing -saml-provider"))
		}
		if c.flagLimit < 0 {
			err = errors.Join(existingErr, fmt.Errorf("-limit must be equal or greater than 0"))
		}
		return err
	})

	return set
}

func (c *SSOCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %q", args)
	}

	gh, err := NewGitHubClient(c.commonFlags.AuthToken)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	samlUsers, err := gh.ListSAMLUsers(ctx, c.flagOrg, c.flagSAMLProvider)
	if err != nil {
		return fmt.Errorf("failed to list SAML users: %w", err)
	}

	userRoles, err := gh.ListUserRoles(ctx, c.flagOrg)
	if err != nil {
		return fmt.Errorf("failed to list user roles: %w", err)
	}

	if c.flagListExceptions {
		printExceptions(c.Stdout(), userRoles, samlUsers, c.flagLimit)
	} else {
		printSAMLUsers(c.Stdout(), userRoles, samlUsers, c.flagLimit)
	}

	return nil
}

func printExceptions(out io.Writer, userRoles, samlUsers map[string]string, limit int) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Login without SSO", "Role"})

	exceptions := sets.SubtractMapKeys(userRoles, samlUsers)
	keys := maps.Keys(exceptions)
	slices.Sort(keys)
	if limit > 0 && limit <= len(keys) {
		keys = keys[:limit]
	}
	for _, login := range keys {
		table.Append([]string{login, exceptions[login]})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func printSAMLUsers(out io.Writer, userRoles, samlUsers map[string]string, limit int) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Login", "Role", "SSO Identity"})

	keys := maps.Keys(samlUsers)
	slices.Sort(keys)
	if limit > 0 && limit <= len(keys) {
		keys = keys[:limit]
	}
	for _, login := range keys {
		table.Append([]string{login, userRoles[login], samlUsers[login]})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}
