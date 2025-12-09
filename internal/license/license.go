package license

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("license not found")
	ErrApplicationNotFound = errors.New("application not found")
	ErrNoAccess         = errors.New("user does not have application access")
	ErrLicenseExpired   = errors.New("license has expired")
	ErrLicenseLimitReached = errors.New("license user limit reached")
)

// Application represents a product/module (like Jira Software, Jira Service Desk)
type Application struct {
	ID          uuid.UUID `db:"id"`
	Key         string    `db:"key"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
}

// Default applications
var (
	AppCore       = "core"       // Base functionality
	AppSoftware   = "software"   // Jira Software (Scrum, Kanban)
	AppServiceDesk = "servicedesk" // Jira Service Desk
)

// License represents the license for the system
type License struct {
	ID             uuid.UUID  `db:"id"`
	LicenseKey     string     `db:"license_key"`
	LicenseType    LicenseType `db:"license_type"`
	MaxUsers       int        `db:"max_users"`
	ExpiresAt      *time.Time `db:"expires_at"`
	LicensedTo     string     `db:"licensed_to"`
	SupportExpires *time.Time `db:"support_expires"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

type LicenseType string

const (
	LicenseTypeCommunity   LicenseType = "community"   // Free, limited users
	LicenseTypeStandard    LicenseType = "standard"    // Paid, per user
	LicenseTypeEnterprise  LicenseType = "enterprise"  // Unlimited users
	LicenseTypeEvaluation  LicenseType = "evaluation"  // Trial
)

func (l *License) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

func (l *License) IsUnlimited() bool {
	return l.MaxUsers <= 0 || l.LicenseType == LicenseTypeEnterprise
}

// ApplicationAccess grants a user access to an application
type ApplicationAccess struct {
	ID            uuid.UUID  `db:"id"`
	UserID        uuid.UUID  `db:"user_id"`
	ApplicationID uuid.UUID  `db:"application_id"`
	GrantedBy     *uuid.UUID `db:"granted_by"`
	CreatedAt     time.Time  `db:"created_at"`
}

// ApplicationGroupAccess grants all members of a group access to an application
type ApplicationGroupAccess struct {
	ID            uuid.UUID `db:"id"`
	GroupID       uuid.UUID `db:"group_id"`
	ApplicationID uuid.UUID `db:"application_id"`
	GrantedBy     *uuid.UUID `db:"granted_by"`
	CreatedAt     time.Time `db:"created_at"`
}

// DefaultAccess - if true, new users automatically get access to this app
type ApplicationDefault struct {
	ApplicationID uuid.UUID `db:"application_id"`
	IsDefault     bool      `db:"is_default"`
}
