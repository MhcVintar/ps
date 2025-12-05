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
		return nil, status.Errorf(codes.Internal, "failed to get LSN: %v", err)
	}

	if lsn+1 == req.Entry.Id {
		var target any
		switch req.Entry.Target {
		case api.WALEntry_TARGET_USER:
			target = &database.User{}
		case api.WALEntry_TARGET_TOPIC:
			target = &database.Topic{}
		case api.WALEntry_TARGET_MESSAGE:
			target = &database.Message{}
		case api.WALEntry_TARGET_LIKE:
			target = &database.Like{}
		}

		if err := json.Unmarshal(req.Entry.Data, target); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to apply WAL: %v", err)
		}

		switch req.Entry.Op {
		case api.WALEntry_OP_SAVE:
			s.db.Save(ctx, target)
		case api.WALEntry_OP_DELETE:
			s.db.Delete(ctx, target)
		}

		shared.Logger.InfoContext(ctx, "applied WAL", "lsn", lsn)
	}

	return s.downstreamClient.Internal.ApplyWALEntry(ctx, req)
}
