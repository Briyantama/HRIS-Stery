package grpc

import (
	"context"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/document/v1"
	"github.com/hris-stery/hris-stery/services/document-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/document-service/internal/application/queries"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DocumentServiceServer struct {
	pb.UnimplementedDocumentServiceServer

	uploadHandler   *commands.UploadDocumentHandler
	deleteHandler   *commands.DeleteDocumentHandler
	versionHandler  *commands.CreateDocumentVersionHandler
	listHandler     *queries.ListDocumentsHandler
	metadataHandler *queries.GetDocumentMetadataHandler
	logger          *zap.Logger
}

func NewDocumentServiceServer(
	uploadHandler *commands.UploadDocumentHandler,
	deleteHandler *commands.DeleteDocumentHandler,
	versionHandler *commands.CreateDocumentVersionHandler,
	listHandler *queries.ListDocumentsHandler,
	metadataHandler *queries.GetDocumentMetadataHandler,
	logger *zap.Logger,
) *DocumentServiceServer {
	return &DocumentServiceServer{
		uploadHandler:   uploadHandler,
		deleteHandler:   deleteHandler,
		versionHandler:  versionHandler,
		listHandler:     listHandler,
		metadataHandler: metadataHandler,
		logger:          logger,
	}
}

// UploadDocument initiates file upload
func (s *DocumentServiceServer) UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.FileName == "" {
		return nil, status.Error(codes.InvalidArgument, "file_name is required")
	}

	uploadedBy, err := actorIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	entityType, err := entityTypeFromPB(req.EntityType)
	if err != nil {
		return nil, err
	}
	classification, err := classificationFromPB(req.Classification)
	if err != nil {
		return nil, err
	}

	result, err := s.uploadHandler.Handle(ctx, commands.UploadDocumentCommand{
		TenantID:       req.TenantId,
		UploadedBy:     uploadedBy,
		EntityType:     entityType,
		EntityID:       req.EntityId,
		FileName:       req.FileName,
		MimeType:       req.MimeType,
		SizeBytes:      req.SizeBytes,
		Classification: classification,
		RetentionDays:  req.RetentionDays,
	})

	if err != nil {
		s.logger.Error("upload failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UploadDocumentResponse{
		FileId:          result.DocumentID,
		UploadSessionId: result.UploadSessionID,
		PresignedUrl:    result.PresignedURL,
		ExpiresAt:       timestamppb.Now(),
	}, nil
}

// DownloadDocument retrieves a file
func (s *DocumentServiceServer) DownloadDocument(ctx context.Context, req *pb.DownloadDocumentRequest) (*pb.DownloadDocumentResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	// Placeholder - retrieve document and return presigned URL
	return &pb.DownloadDocumentResponse{
		PresignedUrl: "",
		ExpiresAt:    timestamppb.Now(),
	}, nil
}

// ListDocuments lists files by entity
func (s *DocumentServiceServer) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	entityType := ""
	if req.EntityType != pb.EntityType_ENTITY_TYPE_UNSPECIFIED {
		var mapErr error
		entityType, mapErr = entityTypeFromPB(req.EntityType)
		if mapErr != nil {
			return nil, mapErr
		}
	}

	result, err := s.listHandler.Handle(ctx, queries.ListDocumentsQuery{
		TenantID:   req.TenantId,
		EntityType: entityType,
		EntityID:   req.EntityId,
		Search:     req.Search,
		PageSize:   req.PageSize,
		Offset:     offsetFromPageToken(req.PageToken),
	})

	if err != nil {
		s.logger.Error("list failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	docs := make([]*pb.DocumentFile, len(result.Documents))
	for i, dto := range result.Documents {
		docs[i] = &pb.DocumentFile{
			FileId:         dto.FileID,
			TenantId:       dto.TenantID,
			UploadedBy:     dto.UploadedBy,
			FileName:       dto.FileName,
			MimeType:       dto.MimeType,
			SizeBytes:      dto.SizeBytes,
			CurrentVersion: dto.CurrentVersion,
			TotalVersions:  dto.TotalVersions,
		}
	}

	return &pb.ListDocumentsResponse{
		Documents:     docs,
		TotalCount:    result.TotalCount,
		NextPageToken: "",
	}, nil
}

// GetDocumentMetadata retrieves complete metadata
func (s *DocumentServiceServer) GetDocumentMetadata(ctx context.Context, req *pb.GetDocumentMetadataRequest) (*pb.GetDocumentMetadataResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	result, err := s.metadataHandler.Handle(ctx, queries.GetDocumentMetadataQuery{
		TenantID:        req.TenantId,
		DocumentID:      req.FileId,
		IncludeVersions: req.IncludeVersions,
	})

	if err != nil {
		s.logger.Error("metadata retrieval failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetDocumentMetadataResponse{
		File: &pb.DocumentFile{
			FileId:    result.Document.FileID,
			TenantId:  result.Document.TenantID,
			FileName:  result.Document.FileName,
			MimeType:  result.Document.MimeType,
			SizeBytes: result.Document.SizeBytes,
		},
	}, nil
}

// DeleteDocument soft-deletes a file
func (s *DocumentServiceServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.FileId == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	result, err := s.deleteHandler.Handle(ctx, commands.DeleteDocumentCommand{
		TenantID:   req.TenantId,
		DocumentID: req.FileId,
		Reason:     req.Reason,
	})

	if err != nil {
		s.logger.Error("delete failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteDocumentResponse{
		Success:   result.Success,
		DeletedAt: timestamppb.Now(),
	}, nil
}

// Placeholder methods for remaining RPCs
func (s *DocumentServiceServer) GetPresignedUrl(ctx context.Context, req *pb.GetPresignedUrlRequest) (*pb.GetPresignedUrlResponse, error) {
	return &pb.GetPresignedUrlResponse{}, nil
}

func (s *DocumentServiceServer) RollbackDocumentVersion(ctx context.Context, req *pb.RollbackDocumentVersionRequest) (*pb.RollbackDocumentVersionResponse, error) {
	return &pb.RollbackDocumentVersionResponse{}, nil
}
