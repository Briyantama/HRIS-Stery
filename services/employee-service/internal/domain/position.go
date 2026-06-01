package domain

import (
	"fmt"
	"time"
)

// PositionLevel enum.
type PositionLevel string

const (
	LevelJunior   PositionLevel = "JUNIOR"
	LevelSenior   PositionLevel = "SENIOR"
	LevelLead     PositionLevel = "LEAD"
	LevelManager  PositionLevel = "MANAGER"
)

// IsValid returns true if the level is a valid PositionLevel.
func (l PositionLevel) IsValid() bool {
	switch l {
	case LevelJunior, LevelSenior, LevelLead, LevelManager:
		return true
	default:
		return false
	}
}

// Position is the aggregate root for job positions.
type Position struct {
	id          PositionID
	tenantID    TenantID
	title       string
	description string
	level       PositionLevel
	createdAt   time.Time
}

// NewPosition creates a new position.
func NewPosition(tenantID TenantID, title string, level PositionLevel) (*Position, error) {
	if title == "" {
		return nil, fmt.Errorf("position title is required")
	}
	if !level.IsValid() {
		return nil, fmt.Errorf("invalid position level: %s", level)
	}

	now := time.Now().UTC()
	return &Position{
		id:        GeneratePositionID(),
		tenantID:  tenantID,
		title:     title,
		level:     level,
		createdAt: now,
	}, nil
}

// RehydratePosition reconstructs a position from persisted data.
func RehydratePosition(
	id PositionID,
	tenantID TenantID,
	title, description string,
	level PositionLevel,
	createdAt time.Time,
) (*Position, error) {
	if title == "" {
		return nil, fmt.Errorf("position title is required")
	}
	if !level.IsValid() {
		return nil, fmt.Errorf("invalid position level: %s", level)
	}

	return &Position{
		id:          id,
		tenantID:    tenantID,
		title:       title,
		description: description,
		level:       level,
		createdAt:   createdAt,
	}, nil
}

// ID returns the position's unique identifier.
func (p *Position) ID() PositionID {
	return p.id
}

// TenantID returns the tenant this position belongs to.
func (p *Position) TenantID() TenantID {
	return p.tenantID
}

// Title returns the position title.
func (p *Position) Title() string {
	return p.title
}

// Description returns the position description.
func (p *Position) Description() string {
	return p.description
}

// Level returns the position level.
func (p *Position) Level() PositionLevel {
	return p.level
}

// CreatedAt returns the creation timestamp.
func (p *Position) CreatedAt() time.Time {
	return p.createdAt
}
