package idempotency

import (
	"context"
	"strings"

	idempotencyuc "github.com/mamahoos/airbar-finance/internal/usecase/idempotency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// UnaryInterceptor deduplicates mutating gRPC commands via Redis + Postgres.
func UnaryInterceptor(guard *idempotencyuc.Guard) grpc.UnaryServerInterceptor {
	if guard == nil {
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		spec, ok := lookupMethod(info.FullMethod)
		if !ok {
			return handler(ctx, req)
		}

		md, _ := metadata.FromIncomingContext(ctx)
		metadataKey := firstMetadataValue(md, metadataKeyHeader)
		key := extractIdempotencyKey(metadataKey, bodyIdempotencyKey(req))

		replay, err := guard.Begin(ctx, key, spec.scope, spec.resourceType, spec.resourceID(req))
		if err != nil {
			return nil, MapDomainError(err)
		}
		if replay != nil {
			resp := spec.newResponse()
			if err := snapshotToResponse(info.FullMethod, replay, resp); err != nil {
				return nil, MapDomainError(err)
			}
			return resp, nil
		}

		resp, err := handler(ctx, req)
		if err != nil {
			_ = guard.Rollback(ctx, key)
			return nil, err
		}

		msg, ok := resp.(proto.Message)
		if !ok {
			_ = guard.Rollback(ctx, key)
			return resp, nil
		}

		snapshot, err := responseToSnapshot(info.FullMethod, msg)
		if err != nil {
			_ = guard.Rollback(ctx, key)
			return nil, err
		}
		if err := guard.Complete(ctx, key, snapshot); err != nil {
			return nil, MapDomainError(err)
		}
		return resp, nil
	}
}

func firstMetadataValue(md metadata.MD, key string) string {
	if md == nil {
		return ""
	}
	values := md.Get(strings.ToLower(key))
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
