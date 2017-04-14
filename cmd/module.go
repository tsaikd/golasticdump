package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/golasticdump/esdump"
)

// command line flags
var (
	flagDebug = &cobrather.BoolFlag{
		Name:    "debug",
		Default: false,
		Usage:   "Enable debug logging",
		EnvVar:  "GOESDUMP_DEBUG",
	}
	flagInput = &cobrather.StringFlag{
		Name:   "input",
		Usage:  "Source elastic URL, e.g. http://localhost:9200 or http://localhost:9200/index-*",
		EnvVar: "GOESDUMP_INPUT",
	}
	flagOutput = &cobrather.StringFlag{
		Name:   "output",
		Usage:  "Destination elastic URL, e.g. http://localhost:9200 or http://localhost:9200/copy-index",
		EnvVar: "GOESDUMP_OUTPUT",
	}
	flagScroll = &cobrather.Int64Flag{
		Name:    "scroll",
		Default: 100,
		Usage:   "Load number per operation",
		EnvVar:  "GOESDUMP_SCROLL",
	}
	flagDelete = &cobrather.BoolFlag{
		Name:    "delete",
		Default: false,
		Usage:   "Delete source data after copy",
		EnvVar:  "GOESDUMP_DELETE",
	}
)

// errors
var (
	ErrEmptyConfig1 = errutil.NewFactory("empty config %q")
)

// Module info
var Module = &cobrather.Module{
	Use:   "gogstash",
	Short: "Logstash like, written in golang",
	Commands: []*cobrather.Module{
		cobrather.VersionModule,
	},
	Flags: []cobrather.Flag{
		flagDebug,
		flagInput,
		flagOutput,
		flagScroll,
		flagDelete,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		inputElasticURL := flagInput.String()
		if inputElasticURL == "" {
			return ErrEmptyConfig1.New(nil, flagInput.Name)
		}
		outputElasticURL := flagOutput.String()
		if outputElasticURL == "" {
			return ErrEmptyConfig1.New(nil, flagOutput.Name)
		}

		return esdump.ElasticDump(esdump.Options{
			Debug:              flagDebug.Bool(),
			InputElasticURL:    inputElasticURL,
			InputElasticSniff:  false,
			OutputElasticURL:   outputElasticURL,
			OutputElasticSniff: false,
			ScrollSize:         int(flagScroll.Int64()),
			BulkActions:        1000,
			BulkSize:           2 << 20, // 2 MB
			BulkFlushInterval:  30 * time.Second,
			Delete:             flagDelete.Bool(),
		})
	},
}
