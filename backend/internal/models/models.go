package models

import (
	"time"

	"github.com/google/uuid"
)

/*
User is a platform account. password_hash is NULL for OAuth-only users.
Confirmation gate: login is blocked while IsConfirmed is false.
*/
type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email               string     `gorm:"size:320;uniqueIndex;not null"`
	PasswordHash        *string    `gorm:"size:255"`
	IsConfirmed         bool       `gorm:"not null;default:false"`
	ConfirmationToken   *string    `gorm:"size:255;uniqueIndex"`
	ConfirmationExpires *time.Time `gorm:"column:confirmation_expires_at;type:timestamptz"`
	CreatedAt           time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt           time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (User) TableName() string { return "users" }

// Session is an opaque, persistent browser session.
type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"type:timestamptz;not null;index"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Session) TableName() string { return "sessions" }

/*
Service is one row of the static service catalog (weather, github, rss...).
AuthType drives which user_services columns get filled:
"none" | "api_key" | "credentials" | "oauth2".
*/
type Service struct {
	ID                uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name              string    `gorm:"size:64;uniqueIndex;not null"`
	DisplayName       string    `gorm:"size:128;not null"`
	Description       string    `gorm:"type:text;not null;default:''"`
	AuthType          string    `gorm:"size:16;not null;check:auth_type in ('none','api_key','credentials','oauth2')"`
	OAuthClientID     *string   `gorm:"size:255"`
	OAuthClientSecret *string   `gorm:"size:255"`
	OAuthAuthorizeURL *string   `gorm:"type:text"`
	OAuthTokenURL     *string   `gorm:"type:text"`
	OAuthScopes       *string   `gorm:"type:text"`
}

func (Service) TableName() string { return "services" }

// SncfStation is a station from the Navitia stop_areas catalog.
type SncfStation struct {
	ID   string `gorm:"type:text;primaryKey"`
	Name string `gorm:"type:text;not null"`
}

func (SncfStation) TableName() string { return "sncf_stations" }

/*
UserService links a user to a service they subscribed to.
*_enc columns hold ciphertext produced by the Go backend BEFORE insert
(application-side encryption — never plaintext, never decrypted by SQL).
*/
type UserService struct {
	ID                   uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID               uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:uq_user_services_user_service"`
	User                 User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ServiceID            uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:uq_user_services_user_service"`
	Service              Service    `gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
	ExternalUsername     *string    `gorm:"size:320"`
	ExternalPasswordEnc  *string    `gorm:"type:text"`
	OAuthAccessTokenEnc  *string    `gorm:"type:text"`
	OAuthRefreshTokenEnc *string    `gorm:"type:text"`
	OAuthTokenExpiresAt  *time.Time `gorm:"type:timestamptz"`
	SubscribedAt         time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (UserService) TableName() string { return "user_services" }

/*
Widget is one entry of the per-service widget catalog.
Declared params live in WidgetParam; valid types: "integer", "string"
(must match the /about.json contract).
*/
type Widget struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ServiceID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uq_widgets_service_name"`
	Service     Service   `gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
	Name        string    `gorm:"size:64;not null;uniqueIndex:uq_widgets_service_name"`
	DisplayName string    `gorm:"size:128;not null"`
	Description string    `gorm:"type:text;not null;default:''"`
}

func (Widget) TableName() string { return "widgets" }

// WidgetParam is one configurable parameter of a catalog widget.
type WidgetParam struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WidgetID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uq_widget_params_widget_name"`
	Widget       Widget    `gorm:"foreignKey:WidgetID;constraint:OnDelete:CASCADE"`
	Name         string    `gorm:"size:64;not null;uniqueIndex:uq_widget_params_widget_name"`
	Type         string    `gorm:"size:16;not null;check:type in ('integer','string')"`
	IsRequired   bool      `gorm:"not null;default:false"`
	DefaultValue *string   `gorm:"type:text"`
}

func (WidgetParam) TableName() string { return "widget_params" }

/*
WidgetInstance is a configured widget placed on a user's dashboard.
Config holds this instance's param values; CachedData the last fetch result.
*/
type WidgetInstance struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index"`
	User               User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	WidgetID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	Widget             Widget         `gorm:"foreignKey:WidgetID;constraint:OnDelete:CASCADE"`
	Config             map[string]any `gorm:"type:jsonb;serializer:json;not null;default:'{}'"`
	RefreshRateSeconds int            `gorm:"not null;check:refresh_rate_seconds > 0"`
	PositionX          int            `gorm:"not null;default:0"`
	PositionY          int            `gorm:"not null;default:0"`
	LastRefreshedAt    *time.Time     `gorm:"type:timestamptz;index"`
	CachedData         map[string]any `gorm:"type:jsonb;serializer:json"`
	CreatedAt          time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt          time.Time      `gorm:"type:timestamptz;not null;default:now()"`
}

func (WidgetInstance) TableName() string { return "widget_instances" }

// All returns every model for AutoMigrate, in dependency order.
func All() []any {
	return []any{
		&User{},
		&Session{},
		&Service{},
		&SncfStation{},
		&UserService{},
		&Widget{},
		&WidgetParam{},
		&WidgetInstance{},
	}
}
