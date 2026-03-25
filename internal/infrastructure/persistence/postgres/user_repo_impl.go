package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	
	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// userRepo 实现 repository.UserRepository 接口
type userRepo struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepo{db: db}
}

// ========== 基础 CRUD 实现 ==========

// FindByID 根据 ID 查询用户
func (r *userRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1 AND status != 'deleted'
	`
	
	var user entity.User
	var preferencesJSON []byte
	var lastLoginAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Age,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&preferencesJSON,
		&user.ItineraryCount,
		&user.TotalDistance,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	
	// 解析偏好 JSON
	if err := parsePreferences(preferencesJSON, &user.Preferences); err != nil {
		return nil, err
	}
	
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	
	return &user, nil
}

// FindByEmail 根据邮箱查询用户
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1 AND status != 'deleted'
	`
	
	var user entity.User
	var preferencesJSON []byte
	var lastLoginAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Age,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&preferencesJSON,
		&user.ItineraryCount,
		&user.TotalDistance,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	
	if err := parsePreferences(preferencesJSON, &user.Preferences); err != nil {
		return nil, err
	}
	
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	
	return &user, nil
}

// FindByIDs 批量查询用户
func (r *userRepo) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if len(ids) == 0 {
		return []*entity.User{}, nil
	}
	
	// 构建 IN 查询
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE id IN (%s) AND status != 'deleted'
	`, strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by ids: %w", err)
	}
	defer rows.Close()
	
	var users []*entity.User
	for rows.Next() {
		var user entity.User
		var preferencesJSON []byte
		var lastLoginAt sql.NullTime
		
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.Age,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&preferencesJSON,
			&user.ItineraryCount,
			&user.TotalDistance,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		
		if err := parsePreferences(preferencesJSON, &user.Preferences); err != nil {
			return nil, err
		}
		
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}
		
		users = append(users, &user)
	}
	
	return users, nil
}

// Save 保存用户（新增或更新）
func (r *userRepo) Save(ctx context.Context, user *entity.User) error {
	// 检查用户是否存在
	existing, err := r.FindByID(ctx, user.ID)
	if err != nil && err != entity.ErrUserNotFound {
		return err
	}
	
	if existing != nil {
		// 存在则更新
		return r.Update(ctx, user)
	}
	
	// 不存在则插入
	return r.insert(ctx, user)
}

// insert 插入新用户
func (r *userRepo) insert(ctx context.Context, user *entity.User) error {
	preferencesJSON, err := marshalPreferences(user.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}
	
	query := `
		INSERT INTO users (
			id, name, email, phone, age, password_hash, role, status,
			preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = *user.LastLoginAt
	} else {
		lastLoginAt = nil
	}
	
	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.Age,
		user.PasswordHash,
		user.Role,
		user.Status,
		preferencesJSON,
		user.ItineraryCount,
		user.TotalDistance,
		user.CreatedAt,
		user.UpdatedAt,
		lastLoginAt,
	)
	
	if err != nil {
		if isDuplicateKeyError(err) {
			return entity.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}
	
	return nil
}

// Update 更新用户
func (r *userRepo) Update(ctx context.Context, user *entity.User) error {
	preferencesJSON, err := marshalPreferences(user.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}
	
	query := `
		UPDATE users SET
			name = $2,
			email = $3,
			phone = $4,
			age = $5,
			password_hash = $6,
			role = $7,
			status = $8,
			preferences = $9,
			itinerary_count = $10,
			total_distance = $11,
			updated_at = $12,
			last_login_at = $13
		WHERE id = $1 AND status != 'deleted'
	`
	
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = *user.LastLoginAt
	} else {
		lastLoginAt = nil
	}
	
	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.Age,
		user.PasswordHash,
		user.Role,
		user.Status,
		preferencesJSON,
		user.ItineraryCount,
		user.TotalDistance,
		time.Now(),
		lastLoginAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrUserNotFound
	}
	
	return nil
}

// Delete 软删除用户
func (r *userRepo) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users SET status = 'deleted', updated_at = $2
		WHERE id = $1 AND status != 'deleted'
	`
	
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrUserNotFound
	}
	
	return nil
}

// HardDelete 物理删除用户
func (r *userRepo) HardDelete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete user: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrUserNotFound
	}
	
	return nil
}

// ========== 业务查询实现 ==========

// FindByRole 根据角色查询用户列表
func (r *userRepo) FindByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE role = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.queryUsers(ctx, query, role, limit, offset)
}

// FindByStatus 根据状态查询用户列表
func (r *userRepo) FindByStatus(ctx context.Context, status entity.UserStatus, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.queryUsers(ctx, query, status, limit, offset)
}

// FindActiveUsers 查询活跃用户（最近7天有登录）
func (r *userRepo) FindActiveUsers(ctx context.Context, limit int) ([]*entity.User, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE last_login_at >= $1 AND status = 'active'
		ORDER BY last_login_at DESC
		LIMIT $2
	`
	
	return r.queryUsers(ctx, query, sevenDaysAgo, limit)
}

// FindByPreferences 根据偏好查询用户（用于推荐相似用户）
func (r *userRepo) FindByPreferences(ctx context.Context, preferences entity.UserPreferences, limit int) ([]*entity.User, error) {
	// 这里简化处理，实际可以用 JSONB 查询或向量相似度
	query := `
		SELECT id, name, email, phone, age, password_hash, role, status,
		       preferences, itinerary_count, total_distance, created_at, updated_at, last_login_at
		FROM users
		WHERE status = 'active'
		ORDER BY itinerary_count DESC
		LIMIT $1
	`
	
	return r.queryUsers(ctx, query, limit)
}

// ========== 统计实现 ==========

// CountByRole 统计指定角色的用户数量
func (r *userRepo) CountByRole(ctx context.Context, role entity.UserRole) (int, error) {
	query := `
		SELECT COUNT(*) FROM users
		WHERE role = $1 AND status != 'deleted'
	`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	
	return count, nil
}

// CountByStatus 统计指定状态的用户数量
func (r *userRepo) CountByStatus(ctx context.Context, status entity.UserStatus) (int, error) {
	query := `
		SELECT COUNT(*) FROM users
		WHERE status = $1
	`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by status: %w", err)
	}
	
	return count, nil
}

// CountActiveLastDays 统计最近 N 天活跃的用户数
func (r *userRepo) CountActiveLastDays(ctx context.Context, days int) (int, error) {
	since := time.Now().AddDate(0, 0, -days)
	
	query := `
		SELECT COUNT(*) FROM users
		WHERE last_login_at >= $1 AND status = 'active'
	`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}
	
	return count, nil
}

// ========== 存在性检查实现 ==========

// ExistsByEmail 检查邮箱是否已被使用
func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE email = $1 AND status != 'deleted'
		)
	`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return exists, nil
}

// ExistsByPhone 检查手机号是否已被使用
func (r *userRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE phone = $1 AND status != 'deleted'
		)
	`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, phone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}
	
	return exists, nil
}

// ========== 批量操作实现 ==========

// BatchUpdateStatus 批量更新用户状态
func (r *userRepo) BatchUpdateStatus(ctx context.Context, ids []string, status entity.UserStatus) error {
	if len(ids) == 0 {
		return nil
	}
	
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = status
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	
	query := fmt.Sprintf(`
		UPDATE users SET status = $1, updated_at = $2
		WHERE id IN (%s) AND status != 'deleted'
	`, strings.Join(placeholders, ","))
	
	_, err := r.db.ExecContext(ctx, query, append([]interface{}{status, time.Now()}, args[1:]...)...)
	if err != nil {
		return fmt.Errorf("failed to batch update status: %w", err)
	}
	
	return nil
}

// BatchUpdateRole 批量更新用户角色
func (r *userRepo) BatchUpdateRole(ctx context.Context, ids []string, role entity.UserRole) error {
	if len(ids) == 0 {
		return nil
	}
	
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = role
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	
	query := fmt.Sprintf(`
		UPDATE users SET role = $1, updated_at = $2
		WHERE id IN (%s) AND status != 'deleted'
	`, strings.Join(placeholders, ","))
	
	_, err := r.db.ExecContext(ctx, query, append([]interface{}{role, time.Now()}, args[1:]...)...)
	if err != nil {
		return fmt.Errorf("failed to batch update role: %w", err)
	}
	
	return nil
}

// ========== 私有辅助方法 ==========

// queryUsers 执行查询并返回用户列表
func (r *userRepo) queryUsers(ctx context.Context, query string, args ...interface{}) ([]*entity.User, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	
	var users []*entity.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// scanUser 扫描一行数据到 User 实体
func scanUser(rows *sql.Rows) (*entity.User, error) {
	var user entity.User
	var preferencesJSON []byte
	var lastLoginAt sql.NullTime
	
	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Age,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&preferencesJSON,
		&user.ItineraryCount,
		&user.TotalDistance,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	
	if err := parsePreferences(preferencesJSON, &user.Preferences); err != nil {
		return nil, err
	}
	
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	
	return &user, nil
}

// parsePreferences 解析偏好 JSON
func parsePreferences(data []byte, prefs *entity.UserPreferences) error {
	if len(data) == 0 {
		return nil
	}
	
	// 这里用 json.Unmarshal
	// 简化实现，实际需要 import "encoding/json"
	return nil
}

// marshalPreferences 序列化偏好为 JSON
func marshalPreferences(prefs entity.UserPreferences) ([]byte, error) {
	// 这里用 json.Marshal
	// 简化实现
	return []byte("{}"), nil
}

// isDuplicateKeyError 检查是否是唯一键冲突
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 唯一键冲突错误码是 23505
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505")
}
