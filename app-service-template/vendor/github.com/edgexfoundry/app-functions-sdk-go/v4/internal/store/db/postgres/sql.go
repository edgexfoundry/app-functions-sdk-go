//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"
)

// Constants for column names in the database table
const (
	contentCol = "content"
	idCol      = "id"
)

// sqlInsertContent returns the SQL statement for inserting a new row with id, name, and content into the table
func sqlInsertContent(table string) string {
	return fmt.Sprintf("INSERT INTO %s(%s, %s) VALUES ($1, $2)", table, idCol, contentCol)
}

// sqlQueryCountByJSONField returns the SQL statement for selecting content column in the table by the given JSON query string ordered by created column
func sqlQueryContentByJSONField(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE content @> $1::jsonb ORDER BY created", table)
}

// sqlCheckExistsById returns the SQL statement for checking if a row exists in the table by id
func sqlCheckExistsById(table string) string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idCol)
}

// sqlUpdateContentById returns the SQL statement for updating the content of a row in the table by id
func sqlUpdateContentById(table string) string {
	return fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, contentCol, idCol)
}

// sqlDeleteById returns the SQL statement for deleting a row from the table by id
func sqlDeleteById(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, idCol)
}
