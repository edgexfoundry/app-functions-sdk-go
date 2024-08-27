//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDBName = "edgex_db"

var once sync.Once
var dc *Client

type Client struct {
	connPool      *pgxpool.Pool
	loggingClient logger.LoggingClient
}

// NewClient returns a pointer to the Postgres client
func NewClient(ctx context.Context, config bootstrapConfig.Database, credentials bootstrapConfig.Credentials, baseScriptPath, extScriptPath string, lc logger.LoggingClient) (*Client, errors.EdgeX) {
	// Get the database name from the environment variable
	databaseName := os.Getenv("EDGEX_DBNAME")
	if databaseName == "" {
		databaseName = defaultDBName
	}

	var edgeXerr errors.EdgeX
	once.Do(func() {
		// use url encode to prevent special characters in the connection string
		connectionStr := "postgres://" + fmt.Sprintf("%s:%s@%s:%d/%s", url.PathEscape(credentials.Username), url.PathEscape(credentials.Password), url.PathEscape(config.Host), config.Port, url.PathEscape(databaseName))
		dbPool, err := pgxpool.New(ctx, connectionStr)
		if err != nil {
			edgeXerr = wrapDBError("fail to create pg connection pool", err)
		}

		dc = &Client{
			connPool:      dbPool,
			loggingClient: lc,
		}
	})
	if edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to create a Postgres client", edgeXerr)
	}

	// invoke Ping to test the connectivity to the DB
	if err := dc.connPool.Ping(ctx); err != nil {
		return nil, wrapDBError("failed to acquire a connection from database connection pool", err)
	}

	lc.Info("Successfully connect to Postgres database")

	// execute base DB scripts
	if edgeXerr = executeDBScripts(ctx, dc.connPool, baseScriptPath); edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute Postgres base DB scripts", edgeXerr)
	}
	if baseScriptPath != "" {
		lc.Info("Successfully execute Postgres base DB scripts")
	}

	// execute extension DB scripts
	if edgeXerr = executeDBScripts(ctx, dc.connPool, extScriptPath); edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute Postgres extension DB scripts", edgeXerr)
	}

	return dc, nil
}

// CloseSession closes the connections to postgres
func (c Client) CloseSession() {
	c.connPool.Close()
}
