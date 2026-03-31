// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package instance

import (
	"context"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/start-codex/tookly/internal/testpg"
)

func TestGetConfig_NotFound(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	_, err := GetConfig(context.Background(), db, "nonexistent_key")
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("error = %v, want ErrConfigNotFound", err)
	}
}

func TestSetConfig_InsertAndUpdate(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	ctx := context.Background()
	key := "test_key_" + testpg.UniqueSuffix(t, db)

	// Insert
	if err := SetConfig(ctx, db, key, "value1"); err != nil {
		t.Fatalf("set config: %v", err)
	}
	val, err := GetConfig(ctx, db, key)
	if err != nil {
		t.Fatalf("get config after insert: %v", err)
	}
	if val != "value1" {
		t.Fatalf("value = %q, want %q", val, "value1")
	}

	// Update
	if err := SetConfig(ctx, db, key, "value2"); err != nil {
		t.Fatalf("update config: %v", err)
	}
	val, err = GetConfig(ctx, db, key)
	if err != nil {
		t.Fatalf("get config after update: %v", err)
	}
	if val != "value2" {
		t.Fatalf("value = %q, want %q", val, "value2")
	}

	// Cleanup
	db.ExecContext(ctx, `DELETE FROM instance_config WHERE key = $1`, key)
}

func TestIsInitialized_FreshDB(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	// Fresh DB has initialized=false from migration seed
	init, err := IsInitialized(context.Background(), db)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if init {
		t.Fatal("IsInitialized = true on fresh DB, want false")
	}
}

func TestIsInitialized_MissingRow(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	ctx := context.Background()
	// Delete the initialized row — should fail closed (error, not false)
	db.ExecContext(ctx, `DELETE FROM instance_config WHERE key = 'initialized'`)
	t.Cleanup(func() {
		db.ExecContext(ctx, `INSERT INTO instance_config (key, value) VALUES ('initialized', 'false') ON CONFLICT DO NOTHING`)
	})

	_, err := IsInitialized(ctx, db)
	if err == nil {
		t.Fatal("IsInitialized with missing row should return error, got nil")
	}
}

func TestIsInitialized_InvalidValue(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	ctx := context.Background()
	SetConfig(ctx, db, "initialized", "corrupted")
	t.Cleanup(func() {
		SetConfig(ctx, db, "initialized", "false")
	})

	_, err := IsInitialized(ctx, db)
	if err == nil {
		t.Fatal("IsInitialized with invalid value should return error, got nil")
	}
}

func TestIsInitialized_AfterSet(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	ctx := context.Background()

	if err := SetConfig(ctx, db, "initialized", "true"); err != nil {
		t.Fatalf("set initialized: %v", err)
	}
	t.Cleanup(func() {
		SetConfig(ctx, db, "initialized", "false")
	})

	init, err := IsInitialized(ctx, db)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !init {
		t.Fatal("IsInitialized = false after setting to true")
	}
}
