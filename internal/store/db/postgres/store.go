//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/store/db/models"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/jackc/pgx/v5"
)

const (
	// constants relate to the postgres db table names
	storeTableName = "store"

	// constants relate to the storeObject fields
	appServiceKeyField = "appServiceKey"
)

// getFullStoreTableName returns the table name with schema as prefix
func (c Client) getFullStoreTableName() string {
	return fmt.Sprintf(`"%s".%s`, c.appServiceKey, storeTableName)
}

// Store persists a stored object to the store table and returns the assigned UUID
func (c Client) Store(o interfaces.StoredObject) (string, error) {
	err := o.ValidateContract(false)
	if err != nil {
		return "", err
	}

	var exists bool
	ctx := context.Background()

	fullTableName := c.getFullStoreTableName()
	err = c.connPool.QueryRow(ctx, sqlCheckExistsById(fullTableName), o.ID).Scan(&exists)
	if err != nil {
		return "", wrapDBError("failed to query app svc store record job by id", err)
	}

	var model models.StoredObject
	model.FromContract(o)

	jsonBytes, err := model.MarshalJSON()
	if err != nil {
		return "", err
	}

	_, err = c.connPool.Exec(
		ctx,
		sqlInsertContent(fullTableName),
		model.ID,
		jsonBytes,
	)
	if err != nil {
		return "", wrapDBError("failed to insert app svc store record", err)
	}

	return model.ID, nil
}

// RetrieveFromStore gets an object from the table with content column contains the appServiceKey
func (c Client) RetrieveFromStore(appServiceKey string) ([]interfaces.StoredObject, error) {
	// do not satisfy requests for a blank ASK
	if appServiceKey == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "no AppServiceKey provided", nil)
	}

	ctx := context.Background()

	queryObj := map[string]any{appServiceKeyField: appServiceKey}
	queryBytes, err := json.Marshal(queryObj)
	if err != nil {
		return nil, wrapDBError("failed to encode appServiceKey query obj", err)
	}

	// query from the store table with content contains {"appServiceKey": appServiceKey}
	rows, err := c.connPool.Query(ctx, sqlQueryContentByJSONField(c.getFullStoreTableName()), queryBytes)
	if err != nil {
		return nil, wrapDBError(fmt.Sprintf("failed to query app svc store record with key '%s'", appServiceKey), err)
	}

	objects, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (interfaces.StoredObject, error) {
		var obj models.StoredObject

		scanErr := row.Scan(&obj)
		if scanErr != nil {
			return interfaces.StoredObject{}, scanErr
		}

		return obj.ToContract(), nil
	})

	if err != nil {
		return nil, wrapDBError("failed to scan app svc store record to StoredObject", err)
	}

	return objects, nil
}

// Update replaces the data currently in the store table by the StoredObject ID
func (c Client) Update(o interfaces.StoredObject) error {
	err := o.ValidateContract(true)
	if err != nil {
		return err
	}

	var update models.StoredObject
	update.FromContract(o)
	json, err := update.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = c.connPool.Exec(
		context.Background(),
		sqlUpdateContentById(c.getFullStoreTableName()),
		json,
		o.ID,
	)
	if err != nil {
		return wrapDBError(fmt.Sprintf("failed to update app svc store record with id '%s'", o.ID), err)
	}

	return nil
}

// RemoveFromStore removes an object from the store table by StoredObject ID
func (c Client) RemoveFromStore(o interfaces.StoredObject) error {
	err := o.ValidateContract(true)
	if err != nil {
		return err
	}

	_, err = c.connPool.Exec(
		context.Background(),
		sqlDeleteById(c.getFullStoreTableName()),
		o.ID,
	)
	if err != nil {
		return wrapDBError(fmt.Sprintf("failed to delete app svc store record with id '%s'", o.ID), err)
	}

	return nil
}

// Disconnect ends the connection.
func (c Client) Disconnect() error {
	c.connPool.Close()
	return nil
}
