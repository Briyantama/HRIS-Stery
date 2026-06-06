package grpc

import (
	"context"
	"strconv"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/document/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func actorIDFromContext(ctx context.Context) (string, error) {
	actorID, ok := ctx.Value("actor_id").(string)
	if !ok || actorID == "" {
		return "", status.Error(codes.Unauthenticated, "actor_id not found in context")
	}
	return actorID, nil
}

func entityTypeFromPB(et pb.EntityType) (string, error) {
	switch et {
	case pb.EntityType_ENTITY_TYPE_EMPLOYEE:
		return "EMPLOYEE", nil
	case pb.EntityType_ENTITY_TYPE_LEAVE:
		return "LEAVE", nil
	case pb.EntityType_ENTITY_TYPE_GENERAL:
		return "GENERAL", nil
	default:
		return "", status.Error(codes.InvalidArgument, "entity_type is required")
	}
}

func classificationFromPB(c pb.FileClassification) (string, error) {
	switch c {
	case pb.FileClassification_FILE_CLASSIFICATION_PUBLIC:
		return "PUBLIC", nil
	case pb.FileClassification_FILE_CLASSIFICATION_CONFIDENTIAL:
		return "CONFIDENTIAL", nil
	case pb.FileClassification_FILE_CLASSIFICATION_SECRET:
		return "SECRET", nil
	default:
		return "", status.Error(codes.InvalidArgument, "classification is required")
	}
}

func offsetFromPageToken(pageToken string) int32 {
	if pageToken == "" {
		return 0
	}
	offset, err := strconv.Atoi(pageToken)
	if err != nil || offset < 0 {
		return 0
	}
	return int32(offset)
}
