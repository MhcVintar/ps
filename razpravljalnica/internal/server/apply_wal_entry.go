package server

import (
	"context"
	"encoding/json"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerNode) ApplyWALEntry(ctx context.Context, req *api.ApplyWALEntryRequest) (*emptypb.Empty, error) {
	lsn, err := s.db.LSN()
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to get LSN", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get LSN: %v", err)
	}

	if lsn+1 == req.Entry.Id {
		if err := s.applyWAL(ctx, req.Entry); err != nil {
			return nil, err
		}
	} else if lsn == req.Entry.Id {
		shared.Logger.InfoContext(ctx, "database is already up to date")
	} else {
		shared.Logger.ErrorContext(ctx, "database is more than 1 WAL behind")
	}

	if !s.isHead() {
		return s.downstreamClient.Internal.ApplyWALEntry(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (s *ServerNode) applyWAL(ctx context.Context, entry *api.WALEntry) error {
	var target any
	switch entry.Target {
	case api.WALEntry_TARGET_USER:
		target = &database.User{}
	case api.WALEntry_TARGET_TOPIC:
		target = &database.Topic{}
	case api.WALEntry_TARGET_MESSAGE:
		target = &database.Message{}
	case api.WALEntry_TARGET_LIKE:
		target = &database.Like{}
	}

	if err := json.Unmarshal(entry.Data, target); err != nil {
		return status.Errorf(codes.Internal, "failed to apply WAL: %v", err)
	}

	switch entry.Op {
	case api.WALEntry_OP_SAVE:
		s.db.Save(ctx, target)
	case api.WALEntry_OP_DELETE:
		s.db.Delete(ctx, target)
	}

	shared.Logger.InfoContext(ctx, "applied WAL", "lsn", entry.Id)
	return nil
}
