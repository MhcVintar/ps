package server

import (
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
)

func (s *ServerNode) TailHandoff(stream grpc.BidiStreamingServer[api.TailHandoffRequest, api.TailHandoffResponse]) error {
	ctx := stream.Context()

	var lastLSN int64
	for entry := range s.db.YieldWAL(ctx, 0) {
		err := stream.Send(&api.TailHandoffResponse{
			Message: &api.TailHandoffResponse_Entry{
				Entry: database.ToApiWALEntry(entry),
			},
		})
		if err != nil {
			shared.Logger.ErrorContext(ctx, "failed to send WAL entry during dump")
			return err
		}

		if _, err := stream.Recv(); err != nil {
			shared.Logger.ErrorContext(ctx, "failed to receive WAL entry ACK during dump")
			return err
		}
		lastLSN = entry.ID
	}

	s.db.Freeze()
	defer s.db.Unfreeze()

	for entry := range s.db.YieldWAL(ctx, lastLSN+1) {
		err := stream.Send(&api.TailHandoffResponse{
			Message: &api.TailHandoffResponse_Entry{
				Entry: database.ToApiWALEntry(entry),
			},
		})
		if err != nil {
			shared.Logger.ErrorContext(ctx, "failed to send WAL entry during final transfer")
			return err
		}

		if _, err := stream.Recv(); err != nil {
			shared.Logger.ErrorContext(ctx, "failed to receive WAL entry ACK during final transfer")
			return err
		}
	}

	stream.Send(&api.TailHandoffResponse{
		Message: &api.TailHandoffResponse_Handoff{
			Handoff: true,
		},
	})

	rewireRequest, err := stream.Recv()
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to receive rewire request")
		return err
	}

	if _, err := s.Rewire(ctx, rewireRequest.RewireRequest); err != nil {
		shared.Logger.ErrorContext(ctx, "failed to rewire")
	}
	return err
}
