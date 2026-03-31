// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package instance

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func getConfig(ctx context.Context, db *sqlx.DB, key string) (string, error) {
	var value string
	err := db.GetContext(ctx, &value,
		`SELECT value FROM instance_config WHERE key = $1`,
		key,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrConfigNotFound
		}
		return "", fmt.Errorf("get config %q: %w", key, err)
	}
	return value, nil
}

func setConfig(ctx context.Context, db *sqlx.DB, key, value string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO instance_config (key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set config %q: %w", key, err)
	}
	return nil
}
