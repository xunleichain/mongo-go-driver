// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package indexopt

import (
	"testing"

	"reflect"

	"go.mongodb.org/mongo-driver/core/option"
	"go.mongodb.org/mongo-driver/internal/testutil/helpers"
)

func createNestedBundle1Drop(t *testing.T) *DropBundle {
	nested := BundleDrop(MaxTime(10))
	testhelpers.RequireNotNil(t, nested, "nested bundle was nil")

	outer := BundleDrop(MaxTime(10), MaxTime(100), nested)
	testhelpers.RequireNotNil(t, nested, "nested bundle was nil")

	return outer
}

func createNestedBundle2Drop(t *testing.T) *DropBundle {
	b1 := BundleDrop(MaxTime(200))
	testhelpers.RequireNotNil(t, b1, "b1 was nil")

	b2 := BundleDrop(MaxTime(15), b1)
	testhelpers.RequireNotNil(t, b2, "b2 was nil")

	outer := BundleDrop(MaxTime(100), MaxTime(20), b2)
	testhelpers.RequireNotNil(t, outer, "outer was nil")

	return outer
}

func createNestedBundle3Drop(t *testing.T) *DropBundle {
	b1 := BundleDrop(MaxTime(100))
	testhelpers.RequireNotNil(t, b1, "b1 was nil")

	b2 := BundleDrop(MaxTime(10), b1)
	testhelpers.RequireNotNil(t, b2, "b2 was nil")

	b3 := BundleDrop(MaxTime(11))
	testhelpers.RequireNotNil(t, b3, "b3 was nil")

	b4 := BundleDrop(MaxTime(200), b3)
	testhelpers.RequireNotNil(t, b4, "b4 was nil")

	outer := BundleDrop(b4, MaxTime(1), b2)
	testhelpers.RequireNotNil(t, outer, "outer was nil")

	return outer
}

func TestDropOpt(t *testing.T) {
	nilBundle := BundleDrop()
	var nilOpts []option.DropIndexesOptioner

	var bundle1 *DropBundle
	bundle1 = bundle1.MaxTime(100).MaxTime(100)
	testhelpers.RequireNotNil(t, bundle1, "created bundle was nil")
	bundle1Opts := []option.DropIndexesOptioner{
		MaxTime(100).ConvertDropOption(),
		MaxTime(100).ConvertDropOption(),
	}
	bundle1OptsDedup := []option.DropIndexesOptioner{
		MaxTime(100).ConvertDropOption(),
	}

	bundle2 := BundleDrop(MaxTime(100))
	bundle2Opts := []option.DropIndexesOptioner{
		MaxTime(100).ConvertDropOption(),
	}

	nested1 := createNestedBundle1Drop(t)
	nestedOpts := []option.DropIndexesOptioner{
		MaxTime(10).ConvertDropOption(),
		MaxTime(100).ConvertDropOption(),
		MaxTime(10).ConvertDropOption(),
	}
	nestedOptsDedup := []option.DropIndexesOptioner{
		MaxTime(10).ConvertDropOption(),
	}

	nested2 := createNestedBundle2Drop(t)
	nested2Opts := []option.DropIndexesOptioner{
		MaxTime(100).ConvertDropOption(),
		MaxTime(20).ConvertDropOption(),
		MaxTime(15).ConvertDropOption(),
		MaxTime(200).ConvertDropOption(),
	}
	nested2OptsDedup := []option.DropIndexesOptioner{
		MaxTime(200).ConvertDropOption(),
	}

	nested3 := createNestedBundle3Drop(t)
	nested3Opts := []option.DropIndexesOptioner{
		MaxTime(200).ConvertDropOption(),
		MaxTime(11).ConvertDropOption(),
		MaxTime(1).ConvertDropOption(),
		MaxTime(10).ConvertDropOption(),
		MaxTime(100).ConvertDropOption(),
	}
	nested3OptsDedup := []option.DropIndexesOptioner{
		MaxTime(100).ConvertDropOption(),
	}

	t.Run("TestAll", func(t *testing.T) {
		opts := []DropOption{
			MaxTime(5000),
		}
		params := make([]Drop, len(opts))
		for i := range opts {
			params[i] = opts[i]
		}
		bundle := BundleDrop(params...)

		deleteOpts, _, err := bundle.Unbundle(true)
		testhelpers.RequireNil(t, err, "got non-nill error from unbundle: %s", err)

		if len(deleteOpts) != len(opts) {
			t.Errorf("expected unbundled opts len %d. got %d", len(opts), len(deleteOpts))
		}

		for i, opt := range opts {
			if !reflect.DeepEqual(opt.ConvertDropOption(), deleteOpts[i]) {
				t.Errorf("opt mismatch. expected %#v, got %#v", opt, deleteOpts[i])
			}
		}
	})

	t.Run("Nil Option Bundle", func(t *testing.T) {
		sess := IndexSessionOpt{}
		opts, _, err := BundleDrop(MaxTime(1), BundleDrop(nil), sess, nil).unbundle()
		testhelpers.RequireNil(t, err, "got non-nil error from unbundle: %s", err)

		if len(opts) != 1 {
			t.Errorf("expected bundle length 1. got: %d", len(opts))
		}

		opts, _, err = BundleDrop(nil, sess, BundleDrop(nil), MaxTime(1)).unbundle()
		testhelpers.RequireNil(t, err, "got non-nil error from unbundle: %s", err)

		if len(opts) != 1 {
			t.Errorf("expected bundle length 1. got: %d", len(opts))
		}
	})

	t.Run("Unbundle", func(t *testing.T) {
		var cases = []struct {
			name         string
			bundle       *DropBundle
			dedup        bool
			expectedOpts []option.DropIndexesOptioner
		}{
			{"NilBundle", nilBundle, false, nilOpts},
			{"Bundle1", bundle1, false, bundle1Opts},
			{"Bundle1Dedup", bundle1, true, bundle1OptsDedup},
			{"Bundle2", bundle2, false, bundle2Opts},
			{"Bundle2", bundle2, true, bundle2Opts},
			{"Nested1", nested1, false, nestedOpts},
			{"Nested1Dedup", nested1, true, nestedOptsDedup},
			{"Nested2", nested2, false, nested2Opts},
			{"Nested2Dedup", nested2, true, nested2OptsDedup},
			{"Nested3", nested3, false, nested3Opts},
			{"Nested3Dedup", nested3, true, nested3OptsDedup},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				opts, _, err := tc.bundle.Unbundle(tc.dedup)
				testhelpers.RequireNil(t, err, "err unbundling db: %s", err)

				if len(opts) != len(tc.expectedOpts) {
					t.Errorf("expectedOpts len mismatch. expected %d, got %d", len(tc.expectedOpts), len(opts))
				}

				for i, opt := range opts {
					if !reflect.DeepEqual(opt, tc.expectedOpts[i]) {
						t.Errorf("expected: %s\nreceived: %s", opt, tc.expectedOpts[i])
					}
				}
			})
		}
	})
}
