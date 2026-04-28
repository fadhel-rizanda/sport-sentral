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
	StatusID       uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	VerifiedAt     *time.Time

	Status    Status     `gorm:"foreignKey:StatusID"`
	UserRoles []UserRole `gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = id
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
	ids := make([]string, len(u.UserRoles))
	for i, role := range u.UserRoles {
		ids[i] = role.RoleID.String()
	}
	return ids
}

func (u *User) Verify() {
	now := time.Now()
	u.VerifiedAt = &now
}

func (u *User) UpdatePassword(newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashed)
	return nil
}
