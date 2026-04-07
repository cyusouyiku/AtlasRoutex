package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserName     = errors.New("invalid user name")
	ErrInvalidUserEmail    = errors.New("invalid user email")
	ErrInvalidUserPassword = errors.New("invalid user password")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidUserRole     = errors.New("invalid user role")
	ErrInvalidUserStatus   = errors.New("invalid user status")
	ErrInvalidUserAge      = errors.New("invalid user age")
	ErrInvalidUserPhone    = errors.New("invalid user phone")
)

func generateUserID() string {
	return uuid.New().String()
}

// UserRole 用户角色。
type UserRole string

const (
	UserRoleAdmin        UserRole = "admin"
	UserRoleUser         UserRole = "user"
	UserRoleVip          UserRole = "vip"
	UserRoleBlackDiamond UserRole = "black_diamond"
	// 兼容早期拼写。
	UserRoleBlackDiomond = UserRoleBlackDiamond
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleUser, UserRoleVip, UserRoleBlackDiamond:
		return true
	default:
		return false
	}
}

func (r UserRole) Level() int {
	switch r {
	case UserRoleAdmin:
		return 4
	case UserRoleBlackDiamond:
		return 3
	case UserRoleVip:
		return 2
	case UserRoleUser:
		return 1
	default:
		return 0
	}
}

// UserStatus 用户状态。
type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
	UserStatusDeleted   UserStatus = "deleted"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusInactive, UserStatusSuspended, UserStatusBanned, UserStatusDeleted:
		return true
	default:
		return false
	}
}

// User 用户实体。
type User struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone,omitempty"`
	Age            int             `json:"age,omitempty"`
	PasswordHash   string          `json:"-"`
	Role           UserRole        `json:"role"`
	Status         UserStatus      `json:"status"`
	Preferences    UserPreferences `json:"preferences"`
	ItineraryCount int             `json:"itinerary_count"`
	TotalDistance  float64         `json:"total_distance"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	LastLoginAt    *time.Time      `json:"last_login_at,omitempty"`
}

// UserPreferences 表示用户的旅行偏好。
type UserPreferences struct {
	PreferredCategories []string `json:"preferred_categories"`
	PreferredTags       []string `json:"preferred_tags"`
	AvoidTags           []string `json:"avoid_tags"`
	Pace                string   `json:"pace"`
	DefaultBudget       int      `json:"default_budget"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	Languages           []string `json:"languages"`
	Currency            string   `json:"currency"`
}

// NewUser 创建一个默认激活的普通用户。
func NewUser(name, email, password string) (*User, error) {
	now := time.Now()
	user := &User{
		ID:     generateUserID(),
		Name:   strings.TrimSpace(name),
		Email:  strings.TrimSpace(email),
		Role:   UserRoleUser,
		Status: UserStatusActive,
		Preferences: UserPreferences{
			Pace:      "normal",
			Currency:  "CNY",
			Languages: []string{"zh"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

// Validate 验证用户信息完整性。
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidUserID
	}
	if n := len(strings.TrimSpace(u.Name)); n < 2 || n > 50 {
		return ErrInvalidUserName
	}
	if !isValidEmail(u.Email) {
		return ErrInvalidUserEmail
	}
	if !u.Role.IsValid() {
		return ErrInvalidUserRole
	}
	if !u.Status.IsValid() {
		return ErrInvalidUserStatus
	}
	if u.Age < 0 || u.Age > 150 {
		return ErrInvalidUserAge
	}
	if u.Phone != "" && !isValidPhone(u.Phone) {
		return ErrInvalidUserPhone
	}
	return nil
}

// SetPassword 设置密码哈希。
func (u *User) SetPassword(plainPassword string) error {
	if len(plainPassword) < 6 {
		return ErrInvalidUserPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// VerifyPassword 校验明文密码。
func (u *User) VerifyPassword(plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPassword)) == nil
}

// Activate 激活用户。
func (u *User) Activate() error {
	if u.Status == UserStatusDeleted {
		return errors.New("cannot activate deleted user")
	}
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
	return nil
}

// Suspend 暂停用户。
func (u *User) Suspend() error {
	if u.Status == UserStatusDeleted {
		return errors.New("cannot suspend deleted user")
	}
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
	return nil
}

// Delete 标记用户为已删除。
func (u *User) Delete() {
	u.Status = UserStatusDeleted
	u.UpdatedAt = time.Now()
}

func (u *User) PromoteToVip() {
	if u.Role.Level() < UserRoleVip.Level() {
		u.Role = UserRoleVip
		u.UpdatedAt = time.Now()
	}
}

func (u *User) PromoteToBlackDiamond() {
	if u.Role.Level() < UserRoleBlackDiamond.Level() {
		u.Role = UserRoleBlackDiamond
		u.UpdatedAt = time.Now()
	}
}

// UpdatePreferences 更新用户偏好。
func (u *User) UpdatePreferences(prefs UserPreferences) {
	u.Preferences = prefs
	u.UpdatedAt = time.Now()
}

func (u *User) AddPreferredCategory(category string) {
	for _, c := range u.Preferences.PreferredCategories {
		if c == category {
			return
		}
	}
	u.Preferences.PreferredCategories = append(u.Preferences.PreferredCategories, category)
	u.UpdatedAt = time.Now()
}

func (u *User) AddPreferredTag(tag string) {
	for _, t := range u.Preferences.PreferredTags {
		if t == tag {
			return
		}
	}
	u.Preferences.PreferredTags = append(u.Preferences.PreferredTags, tag)
	u.UpdatedAt = time.Now()
}

func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 {
		return false
	}
	atIndex := strings.Index(email, "@")
	dotIndex := strings.LastIndex(email, ".")
	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}

func isValidPhone(phone string) bool {
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, ch := range phone {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
