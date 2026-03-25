package entity

import (
	"errors"
	"time"
	"github.com/google/uuid"
)

var (
	ErrInvalidUserName  = "invalid user name"
	ErrInvalidUserEmail = "invalid user email"
	ErrInvalidUserPassword = "invalid user password"
	ErrUserNotFound = "user not found"
	ErrUserAlreadyExists = "user already exists"
	ErrInvalidUserID = "invalid user ID"
	ErrInvalidUserRole = "invalid user role"
	ErrInvalidUserStatus = "invalid user status"
)

func generateUserID() string {
	return uuid.New().String()//依靠谷歌那个库来生成unique的ID
}

// UserRole 用户角色
type UserRole string

const (
	UserRoleAdmin UserRole = "admin" // 管理员
	UserRoleUser  UserRole = "user"  // 普通用户
	UserRoleVip UserRole = "vip"   // VIP用户
	UserRoleBlackDiomond UserRole = "black_diamond" // 黑钻用户
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleUser, UserRoleVip, UserRoleBlackDiomond:
		return true
	default:
		return false
	}
}

func (r UserRole) Level() int {
	switch r {
	case UserRoleAdmin:
		return 4
	case UserRoleBlackDiomond:
		return 3		
	case UserRoleVip:
		return 2
	case UserRoleUser:
		return 1
	default:
		return 0
	}
}

// UserStatus 用户状态
type UserStatus string	

const (
	UserStatusPending  UserStatus = "pending"  // 待审核
	UserStatusActive   UserStatus = "active"   // 活跃
	UserStatusInactive UserStatus = "inactive" // 非活跃
	UserStatusBanned   UserStatus = "banned"   // 被封禁
	UserStatusDeleted  UserStatus = "deleted"  // 已删除
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusInactive, UserStatusBanned, UserStatusDeleted:	
		return true
	default:
		return false
	}
}

// User 用户实体
type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`       // 用户名
	Email    string     `json:"email"`      // 邮箱
	Password string     `json:"-"`          // 密码（不导出）
	Role     UserRole   `json:"role"`       // 角色
	Status   UserStatus `json:"status"`     // 状态
	CreatedAt time.Time  `json:"created_at"` // 创建时间
	UpdatedAt time.Time  `json:"updated_at"` // 更新时间
}

type UserPreferences struct {
	// 旅行偏好
	PreferredCategories []string `json:"preferred_categories"` // 喜欢的景点类型
	PreferredTags       []string `json:"preferred_tags"`       // 喜欢的标签
	AvoidTags           []string `json:"avoid_tags"`           // 避开的标签
	
	// 节奏偏好
	Pace string `json:"pace"` // relaxed, normal, intensive
	
	// 预算偏好
	DefaultBudget int `json:"default_budget"` // 默认每日预算
	
	// 饮食偏好
	DietaryRestrictions []string `json:"dietary_restrictions"` // 饮食限制（素食、清真等）
	
	// 其他
	Languages []string `json:"languages"` // 语言偏好
	Currency  string   `json:"currency"`  // 货币偏好
}

//构造函数
func NewUser(name, email, password string) (*User, error) {
	// 生成 ID
	id := uuid.New().String()
	
	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      UserRoleUser,      // 默认普通用户
		Status:    UserStatusActive,  // 默认激活状态
		Preferences: UserPreferences{
			Pace:        "normal",      // 默认正常节奏
			Currency:    "CNY",         // 默认人民币
			Languages:   []string{"zh"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	
	// 验证用户信息
	if err := user.Validate(); err != nil {
		return nil, err
	}
	
	return user, nil
}

//Validate 验证用户信息，这里就是验证个完整性
func (u *User) Validate() error {
	// 验证 ID
	if u.ID == "" {
		return ErrInvalidUserID
	}
	
	// 验证用户名（非空，长度 2-50）
	if len(u.Name) < 2 || len(u.Name) > 50 {
		return ErrInvalidUserName
	}
	
	// 验证邮箱（非空，格式基本校验）
	if u.Email == "" {
		return ErrInvalidUserEmail
	}
	// 简单的邮箱格式校验
	if !isValidEmail(u.Email) {
		return ErrInvalidUserEmail
	}
	
	// 验证角色
	if !u.Role.IsValid() {
		return ErrInvalidUserRole
	}
	
	// 验证状态
	if !u.Status.IsValid() {
		return ErrInvalidUserStatus
	}
	
	// 验证年龄（可选，0 表示未设置）
	if u.Age < 0 || u.Age > 150 {
		return ErrInvalidUserAge
	}
	
	// 验证手机号（可选）
	if u.Phone != "" && !isValidPhone(u.Phone) {
		return ErrInvalidUserPhone
	}
	
	return nil
}

//密码管理
func (u *User) SetPassword(plainPassword string) error {
	if len(plainPassword) < 6 {
		return ErrInvalidUserPassword
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.PasswordHash = string(hash)
	return nil
}

//验证密码
func (u *User) VerifyPassword(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPassword))
	return err == nil
}

//激活用户
func (u *User) Activate() {
	if u.Status == UserStatusDeleted {
		return errors.New("cannot activate deleted user")
	}
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()	
	return nil
}

//封禁用户
func (u *User) Suspend() error {
	if u.Status == UserStatusDeleted {
		return errors.New("cannot suspend deleted user")
	}
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
	return nil
}

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


// UpdatePreferences 更新用户偏好
func (u *User) UpdatePreferences(prefs UserPreferences) {
	u.Preferences = prefs
	u.UpdatedAt = time.Now()
}

// AddPreferredCategory 添加偏好的景点类型
func (u *User) AddPreferredCategory(category string) {
	// 去重
	for _, c := range u.Preferences.PreferredCategories {
		if c == category {
			return
		}
	}
	u.Preferences.PreferredCategories = append(u.Preferences.PreferredCategories, category)
	u.UpdatedAt = time.Now()
}

// AddPreferredTag 添加偏好的标签
func (u *User) AddPreferredTag(tag string) {
	for _, t := range u.Preferences.PreferredTags {
		if t == tag {
			return
		}
	}
	u.Preferences.PreferredTags = append(u.Preferences.PreferredTags, tag)
	u.UpdatedAt = time.Now()
}


// RecordLogin 记录登录时间
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}



// isValidEmail 简单的邮箱格式校验
func isValidEmail(email string) bool {
	// 这里用简单的校验，生产环境可以用正则
	if len(email) < 3 {
		return false
	}
	atIndex := -1
	dotIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
		}
		if c == '.' && atIndex != -1 {
			dotIndex = i
		}
	}
	return atIndex > 0 && dotIndex > atIndex+1
}

// isValidPhone 简单的手机号校验（中国）
func isValidPhone(phone string) bool {
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	return true
}
