/*
Copyright 2018 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package utils

import (
	//"github.com/imakhlin/v3io-backup/pkg/config"
	"time"
	"v3io-backup/pkg/config"

	"github.com/nuclio/logger"
	"github.com/nuclio/zap"
	"github.com/pkg/errors"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/http"
)

func NewLogger(level string) (logger.Logger, error) {
	var logLevel nucliozap.Level
	switch level {
	case "debug":
		logLevel = nucliozap.DebugLevel
	case "info":
		logLevel = nucliozap.InfoLevel
	case "warn":
		logLevel = nucliozap.WarnLevel
	case "error":
		logLevel = nucliozap.ErrorLevel
	default:
		logLevel = nucliozap.WarnLevel
	}

	log, err := nucliozap.NewNuclioZapCmd("v3io-prom", logLevel)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func CreateContainer(logger logger.Logger, cfg *config.Config, httpTimeout time.Duration) (v3io.Container, error) {
	newContextInput := &v3io.NewContextInput{
		ClusterEndpoints: []string{cfg.WebApiEndpoint},
		NumWorkers:       cfg.ScannerParallelism,
		DialTimeout:      httpTimeout,
	}
	context, err := v3iohttp.NewContext(logger, newContextInput)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a V3IO TSDB client.")
	}

	session, err := context.NewSession(&v3io.NewSessionInput{
		Username:  cfg.Username,
		Password:  cfg.Password,
		AccessKey: cfg.AccessKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create session.")
	}

	container, err := session.NewContainer(&v3io.NewContainerInput{ContainerName: cfg.Container})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create container.")
	}

	return container, nil
}
