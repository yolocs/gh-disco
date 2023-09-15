package commands

import (
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
)

type CommonFlags struct {
	AuthToken string
	LogLevel  string
}

func (f *CommonFlags) Register(fs *cli.FlagSet) {
	sec := fs.NewSection("COMMON FLAGS")

	sec.StringVar(&cli.StringVar{
		Name:   "auth-token",
		Usage:  "GitHub PAT",
		EnvVar: "GHDISCO_AUTH_TOKEN",
		Target: &f.AuthToken,
	})

	sec.LogLevelVar(&cli.LogLevelVar{
		Logger: logging.DefaultLogger(),
	})
}
