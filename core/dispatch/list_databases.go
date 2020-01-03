// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package dispatch

import (
	"context"

	"go.mongodb.org/mongo-driver/core/command"
	"go.mongodb.org/mongo-driver/core/description"
	"go.mongodb.org/mongo-driver/core/result"
	"go.mongodb.org/mongo-driver/core/session"
	"go.mongodb.org/mongo-driver/core/topology"
	"go.mongodb.org/mongo-driver/core/uuid"
)

// ListDatabases handles the full cycle dispatch and execution of a listDatabases command against the provided
// topology.
func ListDatabases(
	ctx context.Context,
	cmd command.ListDatabases,
	topo *topology.Topology,
	selector description.ServerSelector,
	clientID uuid.UUID,
	pool *session.Pool,
) (result.ListDatabases, error) {

	ss, err := topo.SelectServer(ctx, selector)
	if err != nil {
		return result.ListDatabases{}, err
	}

	conn, err := ss.Connection(ctx)
	if err != nil {
		return result.ListDatabases{}, err
	}
	defer conn.Close()

	// If no explicit session and deployment supports sessions, start implicit session.
	if cmd.Session == nil && topo.SupportsSessions() {
		cmd.Session, err = session.NewClientSession(pool, clientID, session.Implicit)
		if err != nil {
			return result.ListDatabases{}, err
		}
		defer cmd.Session.EndSession()
	}

	return cmd.RoundTrip(ctx, ss.Description(), conn)
}
