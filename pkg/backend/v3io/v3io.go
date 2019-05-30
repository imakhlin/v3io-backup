package v3io

import (
	"fmt"
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
	// TODO: find better and lighter test
	path := normalisePath(vds.cfg.BackupOptions.Paths[0])
	fullpath := fmt.Sprintf("%s/%s%s", vds.cfg.WebApiEndpoint, vds.cfg.Container, path)
	resp, err := vds.container.GetContainerContentsSync(&v3io.GetContainerContentsInput{Path: path})
	if err != nil {
		if v3ioUtils.IsNotExistsError(err) {
			return errors.Errorf("File found at container '%s' at path '%s'.", vds.cfg.WebApiEndpoint, path)
		} else {
			return errors.Wrapf(err, "Failed to read from '%s'.", fullpath)
		}
	}

	vds.logger.InfoWith("Connect Response", "Body", string(resp.Body()))

	return nil
}

func normalisePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func (vds *V3ioDataSource) Disconnect() error {
	return errors.Errorf("Not implemented: Disconnect")
}

func (vds *V3ioDataSource) ListDir(path string) (*FileInfoIterator, error) {
	return nil, errors.Errorf("Not implemented: ListDir")
}
func (vds *V3ioDataSource) Scan(path string, modifiedAfterTime time.Time) (*FileInfoIterator, error) {
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

	//err = ds.Connect()
	return &ds, err
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
