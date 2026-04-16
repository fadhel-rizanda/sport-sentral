package entity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email          string    `gorm:"uniqueIndex;not null"`
	Username       string    `gorm:"uniqueIndex;not null"`
	FullName       string    `gorm:"not null;default:''"`
	HashedPassword string    `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Roles []*Role `gorm:"many2many:user_roles;"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func NewUser(email, username, fullName, password string) (*User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		Email:          email,
		Username:       username,
		FullName:       fullName,
		HashedPassword: string(hashed),
	}, nil
}

func (u *User) CheckPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(plain)) == nil
}

func (u *User) Update(fullName, username string) {
	if fullName != "" {
		u.FullName = fullName
	}
	if username != "" {
		u.Username = username
	}
}

func (u *User) RoleIDs() []string {
	ids := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		ids[i] = role.ID.String()
	}
	return ids
}
