// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrConfigNotFound = errors.New("config key not found")

func GetConfig(ctx context.Context, db *sqlx.DB, key string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if key == "" {
		return "", errors.New("key is required")
	}
	return getConfig(ctx, db, key)
}

func SetConfig(ctx context.Context, db *sqlx.DB, key, value string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if key == "" {
		return errors.New("key is required")
	}
	return setConfig(ctx, db, key, value)
}

func IsInitialized(ctx context.Context, db *sqlx.DB) (bool, error) {
	val, err := GetConfig(ctx, db, "initialized")
	if err != nil {
		return false, err
	}
	if val != "true" && val != "false" {
		return false, fmt.Errorf("invalid initialized value: %q", val)
	}
	return val == "true", nil
}
