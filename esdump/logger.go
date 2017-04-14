package esdump

import (
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/logutil"
)

var (
	// logger app logger
	logger = logutil.DefaultLogger
)

func init() {
	formatter := errutil.NewConsoleFormatter("; ")
	errutil.SetDefaultFormatter(formatter)
}
