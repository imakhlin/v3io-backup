package v3io

import (
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	"strings"
	"time"
	v3ioUtils "v3io-backup/pkg/backend/v3io/utils"
	"v3io-backup/pkg/config"
	containerUtils "v3io-backup/pkg/utils"
)

const defaultHttpTimeout = 30 * time.Second

type V3ioDataSource struct {
	logger      logger.Logger
	container   v3io.Container
	HttpTimeout time.Duration
	cfg         *config.Config
}

func (vds *V3ioDataSource) Connect() error {
	path := "/"
	response, err := vds.container.GetContainerContentsSync(&v3io.GetContainerContentsInput{Path: path})
	defer releaseResponse(response)

	if err != nil {
		if v3ioUtils.IsNotExistsError(err) {
			return errors.Errorf("File found at container '%s' at path '%s'.", vds.cfg.WebApiEndpoint, path)
		} else {
			return errors.Wrapf(err, "Failed to read from '%s/%s%s'.", vds.cfg.WebApiEndpoint, vds.cfg.Container, path)
		}
	}

	vds.logger.Info("Connected to container '%s' at '%s'", vds.cfg.Container, vds.cfg.WebApiEndpoint)
	return nil
}

func releaseResponse(response *v3io.Response) {
	if response != nil {
		response.Release()
	}
}

func normalisePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func (vds *V3ioDataSource) Disconnect() error {
	vds.logger.Info("Not implemented: Disconnect")
	return nil
}

func (vds *V3ioDataSource) ListDir(paths []string) (*FileInfoIterator, error) {
	if vds.cfg.BackupOptions.Paths == nil {
		return nil, errors.Errorf("Backup cannot continue without path. Path(s) not set.")
	}

	// TODO: Implement with async iterator
	/* Example
	path := normalisePath(vds.cfg.BackupOptions.Paths[0])
	resp, err := vds.container.GetContainerContentsSync(&v3io.GetContainerContentsInput{Path: path})
	defer releaseResponse(resp)

	if err != nil {
		if v3ioUtils.IsNotExistsError(err) {
			return errors.Errorf("File found at container '%s' at path '%s'.", vds.cfg.WebApiEndpoint, path)
		} else {
			return errors.Wrapf(err, "Failed to read from '%s/%s%s'.", vds.cfg.WebApiEndpoint, vds.cfg.Container, path)
		}
	}

	result := ListBucketResult{}
	xml.Unmarshal(resp.Body(), &result)

	vds.logger.InfoWith("Connect Response", "Result", result, "Rows count", len(result.Contents))
	*/

	return nil, errors.Errorf("Not implemented: ListDir")
}

func (vds *V3ioDataSource) Scan(paths []string, modifiedAfterTime time.Time) (*FileInfoIterator, error) {
	// TODO: Implement with async iterator
	return nil, errors.Errorf("Not implemented: Scan")
}

func NewDataSource(cfg *config.Config) (*V3ioDataSource, error) {
	return newV3ioDataSource(cfg, nil, nil)
}

func newV3ioDataSource(cfg *config.Config, container v3io.Container, logger logger.Logger) (*V3ioDataSource, error) {
	var err error
	ds := V3ioDataSource{}
	ds.cfg = cfg
	if logger != nil {
		ds.logger = logger
	} else {
		ds.logger, err = containerUtils.NewLogger(cfg.LogLevel)
		if err != nil {
			return nil, err
		}
	}

	ds.HttpTimeout = parseHttpTimeout(cfg, logger)

	if container != nil {
		ds.container = container
	} else {
		ds.container, err = containerUtils.CreateContainer(ds.logger, cfg, ds.HttpTimeout)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create V3IO data container")
		}
	}

	// Test connection
	err = ds.Connect()

	if err != nil {
		return nil, err
	}

	return &ds, nil
}

func parseHttpTimeout(cfg *config.Config, logger logger.Logger) time.Duration {
	if cfg.HttpTimeout == "" {
		return defaultHttpTimeout
	} else {
		timeout, err := time.ParseDuration(cfg.HttpTimeout)
		if err != nil {
			logger.Warn("Failed to parse httpTimeout '%s'. Defaulting to %d millis.", cfg.HttpTimeout, defaultHttpTimeout/time.Millisecond)
			return defaultHttpTimeout
		} else {
			return timeout
		}
	}
}
