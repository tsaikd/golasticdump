package esdump

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	isatty "github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	pb "gopkg.in/cheggaaa/pb.v1"
	elastic "gopkg.in/olivere/elastic.v6"
)

type elasticMessage struct {
	elastic.SearchHit
}

type AuthOptions struct {
	BasicUsername string
	BasicPassword string
}

// Options used with ElasticDump
type Options struct {
	Debug              bool
	InputElasticURL    string
	InputBasicAuth     AuthOptions
	InputElasticSniff  bool
	OutputElasticURL   string
	OutputBasicAuth    AuthOptions
	OutputElasticSniff bool
	ScrollSize         int
	BulkActions        int
	BulkSize           int
	BulkFlushInterval  time.Duration
	Delete             bool
	Compress           bool
	MaxRows            int64
}

// ElasticDump dump elastic data with Options
func ElasticDump(opt Options) (err error) {
	if opt.Debug {
		logger.Level = logrus.DebugLevel
	}

	inputElasticURL, inputElasticIndexName, isInputFile, err := parseElasticURL(opt.InputElasticURL)
	if err != nil {
		return
	}
	if isInputFile {
		return fmt.Errorf(`file input not suported`)
	}
	inputClient, err := elastic.NewClient(
		elastic.SetURL(inputElasticURL),
		elastic.SetSniff(opt.InputElasticSniff),
		elastic.SetBasicAuth(opt.InputBasicAuth.BasicUsername, opt.InputBasicAuth.BasicPassword),
	)
	if err != nil {
		return
	}

	outputElasticURL, outputElasticIndexName, isOutputFile, err := parseElasticURL(opt.OutputElasticURL)
	if err != nil {
		return
	}

	var outputClient *elastic.Client

	if !isOutputFile {
		outputClient, err = elastic.NewClient(
			elastic.SetURL(outputElasticURL),
			elastic.SetSniff(opt.OutputElasticSniff),
			elastic.SetBasicAuth(opt.OutputBasicAuth.BasicUsername, opt.OutputBasicAuth.BasicPassword),
		)
		if err != nil {
			return
		}
	}

	ctx := context.Background()
	ctx = contextWithOSSignal(ctx, os.Interrupt, os.Kill)
	g, ctx := errgroup.WithContext(ctx)

	logger.Debug("start")

	totalDoc, err := inputClient.Count(inputElasticIndexName).Do(ctx)
	if err != nil {
		return
	}
	var bar *pb.ProgressBar
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bar = pb.New64(totalDoc).Start()
	}
	defer func() {
		if bar != nil {
			bar.Finish()
		}
	}()

	hits := make(chan elasticMessage, opt.ScrollSize)
	g.Go(func() error {
		defer close(hits)

		// Initialize scroller. Just don't call Do yet.
		scroll := inputClient.Scroll(inputElasticIndexName).Size(opt.ScrollSize)

		return getData(ctx, hits, scroll)
	})

	savedHits := make(chan elasticMessage, opt.ScrollSize)
	g.Go(func() error {
		defer close(savedHits)

		if isOutputFile {
			return setDataFile(ctx, hits, savedHits, outputElasticIndexName, opt.Compress, opt.MaxRows)
		}

		outputBulkProcess, err2 := outputClient.BulkProcessor().
			Name("golasticdump-output").
			BulkActions(opt.BulkActions).
			BulkSize(opt.BulkSize).
			FlushInterval(opt.BulkFlushInterval).
			Do(ctx)
		if err2 != nil {
			return err2
		}

		if err2 := setData(ctx, hits, savedHits, outputBulkProcess, outputElasticIndexName); err2 != nil {
			return err2
		}

		logger.Debug("closing output bulk process")
		return outputBulkProcess.Close()
	})

	g.Go(func() error {
		inputBulkProcess, err2 := inputClient.BulkProcessor().
			Name("golasticdump-input").
			BulkActions(opt.BulkActions).
			BulkSize(opt.BulkSize).
			FlushInterval(opt.BulkFlushInterval).
			Do(ctx)
		if err2 != nil {
			return err2
		}

		if err2 := delData(ctx, savedHits, inputBulkProcess, opt.Delete, bar); err2 != nil {
			return err2
		}

		logger.Debug("closing input bulk process")
		return inputBulkProcess.Close()
	})

	// Check whether any goroutines failed.
	if err = g.Wait(); err != nil {
		return
	}

	return
}

func parseElasticURL(esurl string) (entrypoint string, indexName string, isFile bool, err error) {
	u, err := url.Parse(esurl)
	if err != nil {
		return
	}
	if u.Scheme == "" {
		return u.String(), esurl, true, nil
	} else if u.Scheme == "file" {
		return u.String(), u.Path, true, nil
	}

	indexName = strings.TrimLeft(u.Path, "/")
	u.Path = ""

	return u.String(), indexName, false, nil
}

func getData(
	ctx context.Context,
	hits chan<- elasticMessage,
	scroll *elastic.ScrollService,
) (err error) {
	defer func() {
		logger.Debug("getData defer with err: ", err)
	}()
	for {
		results, err := scroll.Do(ctx)
		if err != nil {
			if err == io.EOF {
				return nil // all results retrieved
			}
			return err // something went wrong
		}

		// Send the hits to the hits channel
		for _, hit := range results.Hits.Hits {
			if hit != nil {
				hits <- elasticMessage{*hit}
			}

			// Check if we need to terminate early
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}
}

func setDataFile(
	ctx context.Context,
	hits <-chan elasticMessage,
	savedHits chan<- elasticMessage,
	outputFileName string,
	compress bool,
	maxRows int64,
) (err error) {
	defer func() {
		if err == nil {
			logger.Debug("setDataFile finish")
		} else {
			logger.Debug("setDataFile err: ", err)
		}
	}()
	var split int
	var outputFile *os.File
	getFileName := func(n int) string {
		fileName := outputFileName
		if maxRows > 0 {
			ext := filepath.Ext(fileName)
			fileName = fmt.Sprintf("%s.split-%d%s", outputFileName[:len(outputFileName)-len(ext)],
				n, ext)
		}
		if compress {
			fileName += ".gz"
		}
		return fileName
	}
	outputFile, err = os.Create(getFileName(split))
	if err != nil {
		return err
	}
	w := gzip.NewWriter(outputFile)
	defer func() {
		if w != nil {
			w.Close()
		}
		if outputFile != nil {
			outputFile.Close()
		}
	}()
	encoder := json.NewEncoder(w)
	var rows int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case hit, ok := <-hits:
			if !ok {
				return nil
			}

			index := hit.Index
			hit.Sort = nil

			logger.Debugf("setDataFile index:%q type:%q id:%q", index, hit.Type, hit.Id)

			if err = encoder.Encode(hit); err != nil {
				return err
			}
			if maxRows > 0 {
				rows++
				if rows >= maxRows {
					rows = 0
					split++

					w.Close()
					w = nil
					outputFile.Close()
					outputFile, err = os.Create(getFileName(split))
					if err != nil {
						return err
					}
					w = gzip.NewWriter(outputFile)
					encoder = json.NewEncoder(w)
				}
			}

			savedHits <- hit
		}
	}
	return nil
}

func setData(
	ctx context.Context,
	hits <-chan elasticMessage,
	savedHits chan<- elasticMessage,
	outputBulkProcess *elastic.BulkProcessor,
	outputElasticIndexName string,
) (err error) {
	defer func() {
		logger.Debug("setData defer with err: ", err)
	}()
	for hit := range hits {
		index := hit.Index
		if outputElasticIndexName != "" {
			index = outputElasticIndexName
		}

		logger.Debugf("setData index:%q type:%q id:%q", index, hit.Type, hit.Id)
		indexRequest := elastic.NewBulkIndexRequest().Index(index).Type(hit.Type).Id(hit.Id).Doc(hit.Source)
		outputBulkProcess.Add(indexRequest)

		// Check if we need to terminate early
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			savedHits <- hit
		}
	}
	logger.Debug("setData finish")
	return nil
}

func delData(
	ctx context.Context,
	savedHits <-chan elasticMessage,
	inputBulkProcess *elastic.BulkProcessor,
	delete bool,
	bar *pb.ProgressBar,
) (err error) {
	defer func() {
		logger.Debug("delData defer with err: ", err)
	}()
	for hit := range savedHits {
		if delete {
			logger.Debugf("delData index:%q type:%q id:%q", hit.Index, hit.Type, hit.Id)
			deleteRequest := elastic.NewBulkDeleteRequest().Index(hit.Index).Type(hit.Type).Id(hit.Id)
			inputBulkProcess.Add(deleteRequest)
		}

		// Check if we need to terminate early
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if bar != nil {
				bar.Increment()
			}
		}
	}
	logger.Debug("delData finish")
	return nil
}
