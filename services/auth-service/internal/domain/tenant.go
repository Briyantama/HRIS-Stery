package domain

import (
	"fmt"
	"regexp"
	"time"
)

// TenantSlug is a URL-safe identifier for a tenant.
type TenantSlug struct {
	value string
}

// NewTenantSlug validates and creates a TenantSlug.
// Slug must be 3-50 alphanumeric characters and hyphens, lowercase.
func NewTenantSlug(slug string) (TenantSlug, error) {
	if len(slug) < 3 || len(slug) > 50 {
		return TenantSlug{}, fmt.Errorf("slug must be 3-50 characters, got %d", len(slug))
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(slug) {
		return TenantSlug{}, fmt.Errorf("slug must contain only lowercase letters, digits, and hyphens")
	}
	return TenantSlug{value: slug}, nil
}

// MustNewTenantSlug panics if the slug is invalid.
func MustNewTenantSlug(slug string) TenantSlug {
	s, err := NewTenantSlug(slug)
	if err != nil {
		panic(err)
	}
	return s
}

// String returns the slug string.
func (s TenantSlug) String() string {
	return s.value
}

// Tenant is the aggregate root for a company/organization in the system.
type Tenant struct {
	id          TenantID
	slug        TenantSlug
	companyName string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewTenant creates a new Tenant aggregate.
func NewTenant(id TenantID, slug TenantSlug, companyName string) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		id:          id,
		slug:        slug,
		companyName: companyName,
		isActive:    true,
		createdAt:   now,
		updatedAt:   now,
	}
}

// ID returns the tenant's unique identifier.
func (t *Tenant) ID() TenantID {
	return t.id
}

// Slug returns the tenant's URL-safe slug.
func (t *Tenant) Slug() TenantSlug {
	return t.slug
}

// CompanyName returns the tenant's company name.
func (t *Tenant) CompanyName() string {
	return t.companyName
}

// IsActive returns whether the tenant is active.
func (t *Tenant) IsActive() bool {
	return t.isActive
}

// Deactivate marks the tenant as inactive. No users can log in once deactivated.
func (t *Tenant) Deactivate() {
	t.isActive = false
	t.updatedAt = time.Now().UTC()
}

// CreatedAt returns the creation timestamp.
func (t *Tenant) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns the last update timestamp.
func (t *Tenant) UpdatedAt() time.Time {
	return t.updatedAt
}
