package domain

type FileStatus string
type FileClassification string
type EntityType string

const (
	// FileStatus values
	StatusUploaded   FileStatus = "UPLOADED"
	StatusScanned    FileStatus = "SCANNED"
	StatusScanFailed FileStatus = "SCAN_FAILED"
	StatusDeleted    FileStatus = "DELETED"
)

const (
	// FileClassification values
	ClassificationPublic       FileClassification = "PUBLIC"
	ClassificationConfidential FileClassification = "CONFIDENTIAL"
	ClassificationSecret       FileClassification = "SECRET"
)

const (
	// EntityType values
	EntityTypeEmployee EntityType = "EMPLOYEE"
	EntityTypeLeave    EntityType = "LEAVE"
	EntityTypeGeneral  EntityType = "GENERAL"
)

// ValidateFileStatus checks if a FileStatus is valid
func ValidateFileStatus(status FileStatus) bool {
	switch status {
	case StatusUploaded, StatusScanned, StatusScanFailed, StatusDeleted:
		return true
	default:
		return false
	}
}

// ValidateClassification checks if a FileClassification is valid
func ValidateClassification(classification FileClassification) bool {
	switch classification {
	case ClassificationPublic, ClassificationConfidential, ClassificationSecret:
		return true
	default:
		return false
	}
}

// ValidateEntityType checks if an EntityType is valid
func ValidateEntityType(entityType EntityType) bool {
	switch entityType {
	case EntityTypeEmployee, EntityTypeLeave, EntityTypeGeneral:
		return true
	default:
		return false
	}
}
