package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// ItineraryRepositoryImpl 是 ItineraryRepository 的 PostgreSQL 实现
type ItineraryRepositoryImpl struct {
	db *sql.DB
}

// NewItineraryRepositoryImpl 创建行程仓储实例
func NewItineraryRepositoryImpl(db *sql.DB) repository.ItineraryRepository {
	return &ItineraryRepositoryImpl{db: db}
}

// ========== 基础 CRUD 实现 ==========

// FindByID 根据 ID 查询行程
func (r *ItineraryRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE id = $1 AND deleted_at IS NULL
	`

	var itinerary entity.Itinerary
	var budgetJSON, statisticsJSON, constraintsJSON []byte
	var publishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&itinerary.ID,
		&itinerary.UserID,
		&itinerary.Name,
		&itinerary.Description,
		&itinerary.Status,
		&itinerary.StartDate,
		&itinerary.EndDate,
		&itinerary.DayCount,
		&budgetJSON,
		&statisticsJSON,
		&constraintsJSON,
		&itinerary.CreatedAt,
		&itinerary.UpdatedAt,
		&publishedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.ErrItineraryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find itinerary by id: %w", err)
	}

	// 解析 JSON 字段
	if err := json.Unmarshal(budgetJSON, &itinerary.Budget); err != nil {
		return nil, fmt.Errorf("failed to parse budget: %w", err)
	}
	if err := json.Unmarshal(statisticsJSON, &itinerary.Statistics); err != nil {
		return nil, fmt.Errorf("failed to parse statistics: %w", err)
	}
	if err := json.Unmarshal(constraintsJSON, &itinerary.Constraints); err != nil {
		return nil, fmt.Errorf("failed to parse constraints: %w", err)
	}

	if publishedAt.Valid {
		itinerary.PublishedAt = &publishedAt.Time
	}

	// 查询每天的行程详情
	days, err := r.fetchDays(ctx, itinerary.ID)
	if err != nil {
		return nil, err
	}
	itinerary.Days = days

	return &itinerary, nil
}

// FindByUserID 查询用户的所有行程
func (r *ItineraryRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by user id: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByName 根据行程名称模糊搜索
func (r *ItineraryRepositoryImpl) FindByName(ctx context.Context, name string) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE name ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by name: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// Save 保存行程（新增）
func (r *ItineraryRepositoryImpl) Save(ctx context.Context, itinerary *entity.Itinerary) error {
	// 检查是否已存在
	existing, err := r.FindByID(ctx, itinerary.ID)
	if err != nil && err != entity.ErrItineraryNotFound {
		return err
	}

	if existing != nil {
		// 存在则更新
		return r.Update(ctx, itinerary)
	}

	// 不存在则插入
	return r.insert(ctx, itinerary)
}

// insert 插入新行程
func (r *ItineraryRepositoryImpl) insert(ctx context.Context, itinerary *entity.Itinerary) error {
	budgetJSON, err := json.Marshal(itinerary.Budget)
	if err != nil {
		return fmt.Errorf("failed to marshal budget: %w", err)
	}

	statisticsJSON, err := json.Marshal(itinerary.Statistics)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	constraintsJSON, err := json.Marshal(itinerary.Constraints)
	if err != nil {
		return fmt.Errorf("failed to marshal constraints: %w", err)
	}

	var publishedAt interface{}
	if itinerary.PublishedAt != nil {
		publishedAt = *itinerary.PublishedAt
	} else {
		publishedAt = nil
	}

	query := `
		INSERT INTO itineraries (
			id, user_id, name, description, status, start_date, end_date, day_count,
			budget, statistics, constraints, created_at, updated_at, published_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.db.ExecContext(ctx, query,
		itinerary.ID,
		itinerary.UserID,
		itinerary.Name,
		itinerary.Description,
		itinerary.Status,
		itinerary.StartDate,
		itinerary.EndDate,
		itinerary.DayCount,
		budgetJSON,
		statisticsJSON,
		constraintsJSON,
		itinerary.CreatedAt,
		itinerary.UpdatedAt,
		publishedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert itinerary: %w", err)
	}

	// 保存每天的行程详情
	if err := r.saveDays(ctx, itinerary.ID, itinerary.Days); err != nil {
		return err
	}

	return nil
}

// Update 更新行程
func (r *ItineraryRepositoryImpl) Update(ctx context.Context, itinerary *entity.Itinerary) error {
	budgetJSON, err := json.Marshal(itinerary.Budget)
	if err != nil {
		return fmt.Errorf("failed to marshal budget: %w", err)
	}

	statisticsJSON, err := json.Marshal(itinerary.Statistics)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	constraintsJSON, err := json.Marshal(itinerary.Constraints)
	if err != nil {
		return fmt.Errorf("failed to marshal constraints: %w", err)
	}

	var publishedAt interface{}
	if itinerary.PublishedAt != nil {
		publishedAt = *itinerary.PublishedAt
	} else {
		publishedAt = nil
	}

	query := `
		UPDATE itineraries SET
			user_id = $2,
			name = $3,
			description = $4,
			status = $5,
			start_date = $6,
			end_date = $7,
			day_count = $8,
			budget = $9,
			statistics = $10,
			constraints = $11,
			updated_at = $12,
			published_at = $13
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		itinerary.ID,
		itinerary.UserID,
		itinerary.Name,
		itinerary.Description,
		itinerary.Status,
		itinerary.StartDate,
		itinerary.EndDate,
		itinerary.DayCount,
		budgetJSON,
		statisticsJSON,
		constraintsJSON,
		time.Now(),
		publishedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update itinerary: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrItineraryNotFound
	}

	// 先删除旧的 daily 记录，再插入新的
	if err := r.deleteDays(ctx, itinerary.ID); err != nil {
		return err
	}
	if err := r.saveDays(ctx, itinerary.ID, itinerary.Days); err != nil {
		return err
	}

	return nil
}

// Delete 软删除行程
func (r *ItineraryRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `UPDATE itineraries SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete itinerary: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrItineraryNotFound
	}

	return nil
}

// HardDelete 物理删除行程
func (r *ItineraryRepositoryImpl) HardDelete(ctx context.Context, id string) error {
	// 先删除关联的 days
	if err := r.deleteDays(ctx, id); err != nil {
		return err
	}

	query := `DELETE FROM itineraries WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete itinerary: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrItineraryNotFound
	}

	return nil
}

// ========== 业务查询实现 ==========

// FindByStatus 根据状态查询行程
func (r *ItineraryRepositoryImpl) FindByStatus(ctx context.Context, status entity.ItineraryStatus) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by status: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByUserIDAndStatus 查询用户的特定状态行程
func (r *ItineraryRepositoryImpl) FindByUserIDAndStatus(ctx context.Context, userID string, status entity.ItineraryStatus) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE user_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by user and status: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByDateRange 查询指定日期范围内的行程
func (r *ItineraryRepositoryImpl) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE start_date >= $1 AND end_date <= $2 AND deleted_at IS NULL
		ORDER BY start_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by date range: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByUserIDAndDateRange 查询用户在指定日期范围内的行程
func (r *ItineraryRepositoryImpl) FindByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE user_id = $1 AND start_date >= $2 AND end_date <= $3 AND deleted_at IS NULL
		ORDER BY start_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by user and date range: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByBudgetRange 查询指定预算范围内的行程
func (r *ItineraryRepositoryImpl) FindByBudgetRange(ctx context.Context, minBudget, maxBudget float64) ([]*entity.Itinerary, error) {
	// 使用 JSONB 查询预算字段
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE (budget->>'total_cost')::float BETWEEN $1 AND $2
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, minBudget, maxBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by budget range: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByCity 查询指定城市的行程
func (r *ItineraryRepositoryImpl) FindByCity(ctx context.Context, city string, limit int) ([]*entity.Itinerary, error) {
	// 需要先从 days 表中查询景点信息来获取城市
	// 这里简化处理，假设行程表有 city 字段
	query := `
		SELECT DISTINCT i.id, i.user_id, i.name, i.description, i.status, i.start_date, i.end_date, i.day_count,
		       i.budget, i.statistics, i.constraints, i.created_at, i.updated_at, i.published_at
		FROM itineraries i
		JOIN itinerary_days d ON i.id = d.itinerary_id
		JOIN day_attractions a ON d.id = a.day_id
		JOIN pois p ON a.poi_id = p.id
		WHERE p.city = $1 AND i.deleted_at IS NULL
		ORDER BY i.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, city, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by city: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByAttraction 查询包含指定景点的行程
func (r *ItineraryRepositoryImpl) FindByAttraction(ctx context.Context, attractionID string) ([]*entity.Itinerary, error) {
	query := `
		SELECT i.id, i.user_id, i.name, i.description, i.status, i.start_date, i.end_date, i.day_count,
		       i.budget, i.statistics, i.constraints, i.created_at, i.updated_at, i.published_at
		FROM itineraries i
		JOIN itinerary_days d ON i.id = d.itinerary_id
		JOIN day_attractions a ON d.id = a.day_id
		WHERE a.poi_id = $1 AND i.deleted_at IS NULL
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, attractionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by attraction: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindRecentItineraries 查询最近创建的行程
func (r *ItineraryRepositoryImpl) FindRecentItineraries(ctx context.Context, limit int) ([]*entity.Itinerary, error) {
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent itineraries: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindPopularItineraries 查询热门行程
func (r *ItineraryRepositoryImpl) FindPopularItineraries(ctx context.Context, limit int) ([]*entity.Itinerary, error) {
	// 按收藏数排序（需要在 itinerary 表增加 favorite_count 字段）
	query := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE deleted_at IS NULL AND status = 'published'
		ORDER BY favorite_count DESC, created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find popular itineraries: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// FindByConstraints 根据约束条件查询行程（高级查询）
func (r *ItineraryRepositoryImpl) FindByConstraints(ctx context.Context, query *repository.ItineraryQuery) ([]*entity.Itinerary, error) {
	sql, args := r.buildQuery(query)

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find itineraries by constraints: %w", err)
	}
	defer rows.Close()

	return r.scanItineraries(rows)
}

// ========== 统计实现 ==========

// CountByUserID 统计用户的行程数量
func (r *ItineraryRepositoryImpl) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM itineraries WHERE user_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count itineraries by user: %w", err)
	}

	return count, nil
}

// CountByCity 统计指定城市的行程数量
func (r *ItineraryRepositoryImpl) CountByCity(ctx context.Context, city string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT i.id)
		FROM itineraries i
		JOIN itinerary_days d ON i.id = d.itinerary_id
		JOIN day_attractions a ON d.id = a.day_id
		JOIN pois p ON a.poi_id = p.id
		WHERE p.city = $1 AND i.deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, city).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count itineraries by city: %w", err)
	}

	return count, nil
}

// CountByStatus 统计指定状态的行程数量
func (r *ItineraryRepositoryImpl) CountByStatus(ctx context.Context, status entity.ItineraryStatus) (int, error) {
	query := `SELECT COUNT(*) FROM itineraries WHERE status = $1 AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count itineraries by status: %w", err)
	}

	return count, nil
}

// CountByDateRange 统计指定日期范围内的行程数量
func (r *ItineraryRepositoryImpl) CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM itineraries
		WHERE start_date >= $1 AND end_date <= $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count itineraries by date range: %w", err)
	}

	return count, nil
}

// ========== 存在性检查 ==========

// ExistsByUserIDAndName 检查用户是否已有同名行程
func (r *ItineraryRepositoryImpl) ExistsByUserIDAndName(ctx context.Context, userID, name string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM itineraries
			WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}

// ========== 批量操作 ==========

// BatchUpdateStatus 批量更新行程状态
func (r *ItineraryRepositoryImpl) BatchUpdateStatus(ctx context.Context, ids []string, status entity.ItineraryStatus) error {
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
		UPDATE itineraries SET status = $1, updated_at = $2
		WHERE id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, append([]interface{}{status, time.Now()}, args[1:]...)...)
	if err != nil {
		return fmt.Errorf("failed to batch update status: %w", err)
	}

	return nil
}

// BatchDelete 批量软删除行程
func (r *ItineraryRepositoryImpl) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE itineraries SET deleted_at = $%d
		WHERE id IN (%s) AND deleted_at IS NULL
	`, len(ids)+1, strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, append(args, time.Now())...)
	if err != nil {
		return fmt.Errorf("failed to batch delete itineraries: %w", err)
	}

	return nil
}

// ========== 私有辅助方法 ==========

// fetchDays 获取行程的每日详情
func (r *ItineraryRepositoryImpl) fetchDays(ctx context.Context, itineraryID string) ([]*entity.ItineraryDay, error) {
	// 查询 days
	daysQuery := `
		SELECT id, day_number, date, notes, walking_distance, walking_time, place_count, daily_cost, attraction_time, is_rest
		FROM itinerary_days
		WHERE itinerary_id = $1
		ORDER BY day_number ASC
	`

	rows, err := r.db.QueryContext(ctx, daysQuery, itineraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch days: %w", err)
	}
	defer rows.Close()

	var days []*entity.ItineraryDay
	for rows.Next() {
		day := &entity.ItineraryDay{
			Attractions: make([]*entity.DayAttraction, 0),
			Meals:       make([]*entity.DayMeal, 0),
			Statistics:  &entity.DayStatistics{},
		}

		err := rows.Scan(
			&day.DayNumber,
			&day.Date,
			&day.Notes,
			&day.Statistics.WalkingDistance,
			&day.Statistics.WalkingTime,
			&day.Statistics.PlaceCount,
			&day.Statistics.DailyCost,
			&day.Statistics.AttractionTime,
			&day.Statistics.IsRest,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan day: %w", err)
		}

		// 查询该天的景点
		attractions, err := r.fetchAttractions(ctx, day.DayNumber, itineraryID)
		if err != nil {
			return nil, err
		}
		day.Attractions = attractions

		// 查询该天的餐饮
		meals, err := r.fetchMeals(ctx, day.DayNumber, itineraryID)
		if err != nil {
			return nil, err
		}
		day.Meals = meals

		// 查询该天的酒店
		hotel, err := r.fetchHotel(ctx, day.DayNumber, itineraryID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		day.Hotel = hotel

		days = append(days, day)
	}

	return days, nil
}

// fetchAttractions 查询某天的景点
func (r *ItineraryRepositoryImpl) fetchAttractions(ctx context.Context, dayNumber int, itineraryID string) ([]*entity.DayAttraction, error) {
	query := `
		SELECT id, poi_id, start_time, end_time, stay_duration, "order", cost, transportation, notes
		FROM day_attractions
		WHERE itinerary_id = $1 AND day_number = $2
		ORDER BY "order" ASC
	`

	rows, err := r.db.QueryContext(ctx, query, itineraryID, dayNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attractions: %w", err)
	}
	defer rows.Close()

	var attractions []*entity.DayAttraction
	for rows.Next() {
		attr := &entity.DayAttraction{}
		var poiID string
		var transportationJSON []byte

		err := rows.Scan(
			&attr.ID,
			&poiID,
			&attr.StartTime,
			&attr.EndTime,
			&attr.StayDuration,
			&attr.Order,
			&attr.Cost,
			&transportationJSON,
			&attr.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attraction: %w", err)
		}

		// 加载 POI 信息
		poi, err := r.loadPOI(ctx, poiID)
		if err != nil {
			return nil, err
		}
		attr.POI = poi

		// 解析交通信息
		if len(transportationJSON) > 0 {
			var trans entity.Transportation
			if err := json.Unmarshal(transportationJSON, &trans); err == nil {
				attr.Transportation = &trans
			}
		}

		attractions = append(attractions, attr)
	}

	return attractions, nil
}

// fetchMeals 查询某天的餐饮
func (r *ItineraryRepositoryImpl) fetchMeals(ctx context.Context, dayNumber int, itineraryID string) ([]*entity.DayMeal, error) {
	query := `
		SELECT id, meal_type, restaurant_id, meal_time, cost, notes
		FROM day_meals
		WHERE itinerary_id = $1 AND day_number = $2
		ORDER BY meal_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, itineraryID, dayNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch meals: %w", err)
	}
	defer rows.Close()

	var meals []*entity.DayMeal
	for rows.Next() {
		meal := &entity.DayMeal{}
		var restaurantID string

		err := rows.Scan(
			&meal.ID,
			&meal.MealType,
			&restaurantID,
			&meal.Time,
			&meal.Cost,
			&meal.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal: %w", err)
		}

		// 加载餐厅信息
		if restaurantID != "" {
			restaurant, err := r.loadPOI(ctx, restaurantID)
			if err == nil {
				meal.Restaurant = restaurant
			}
		}

		meals = append(meals, meal)
	}

	return meals, nil
}

// fetchHotel 查询某天的酒店
func (r *ItineraryRepositoryImpl) fetchHotel(ctx context.Context, dayNumber int, itineraryID string) (*entity.DayHotel, error) {
	query := `
		SELECT id, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes
		FROM day_hotels
		WHERE itinerary_id = $1 AND day_number = $2
	`

	hotel := &entity.DayHotel{}
	err := r.db.QueryRowContext(ctx, query, itineraryID, dayNumber).Scan(
		&hotel.ID,
		&hotel.HotelName,
		&hotel.Address,
		&hotel.CheckInTime,
		&hotel.CheckOutTime,
		&hotel.Cost,
		&hotel.RoomType,
		&hotel.Notes,
	)

	if err != nil {
		return nil, err
	}

	return hotel, nil
}

// loadPOI 加载 POI 信息
func (r *ItineraryRepositoryImpl) loadPOI(ctx context.Context, poiID string) (*entity.POI, error) {
	// 这里调用 POI 仓储的方法，避免循环依赖
	// 实际项目中应该使用依赖注入
	query := `
		SELECT id, name, name_en, category, lat, lng, address, city, rating, price_level
		FROM pois
		WHERE id = $1
	`

	var poi entity.POI
	err := r.db.QueryRowContext(ctx, query, poiID).Scan(
		&poi.ID,
		&poi.Name,
		&poi.NameEn,
		&poi.Category,
		&poi.Location.Lat,
		&poi.Location.Lng,
		&poi.Address,
		&poi.City,
		&poi.Rating,
		&poi.PriceLevel,
	)

	if err != nil {
		return nil, err
	}

	return &poi, nil
}

// saveDays 保存每天的行程详情
func (r *ItineraryRepositoryImpl) saveDays(ctx context.Context, itineraryID string, days []*entity.ItineraryDay) error {
	for _, day := range days {
		// 插入 day
		dayQuery := `
			INSERT INTO itinerary_days (
				id, itinerary_id, day_number, date, notes, walking_distance, walking_time,
				place_count, daily_cost, attraction_time, is_rest
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		dayID := generateID()
		_, err := r.db.ExecContext(ctx, dayQuery,
			dayID,
			itineraryID,
			day.DayNumber,
			day.Date,
			day.Notes,
			day.Statistics.WalkingDistance,
			day.Statistics.WalkingTime,
			day.Statistics.PlaceCount,
			day.Statistics.DailyCost,
			day.Statistics.AttractionTime,
			day.Statistics.IsRest,
		)
		if err != nil {
			return fmt.Errorf("failed to save day: %w", err)
		}

		// 保存该天的景点
		for _, attr := range day.Attractions {
			if err := r.saveAttraction(ctx, itineraryID, day.DayNumber, attr); err != nil {
				return err
			}
		}

		// 保存该天的餐饮
		for _, meal := range day.Meals {
			if err := r.saveMeal(ctx, itineraryID, day.DayNumber, meal); err != nil {
				return err
			}
		}

		// 保存该天的酒店
		if day.Hotel != nil {
			if err := r.saveHotel(ctx, itineraryID, day.DayNumber, day.Hotel); err != nil {
				return err
			}
		}
	}

	return nil
}

// saveAttraction 保存景点
func (r *ItineraryRepositoryImpl) saveAttraction(ctx context.Context, itineraryID string, dayNumber int, attr *entity.DayAttraction) error {
	if attr == nil || attr.POI == nil {
		return nil
	}
	transportationJSON, _ := json.Marshal(attr.Transportation)

	query := `
		INSERT INTO day_attractions (
			id, itinerary_id, day_number, poi_id, start_time, end_time,
			stay_duration, "order", cost, transportation, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		generateID(),
		itineraryID,
		dayNumber,
		attr.POI.ID,
		attr.StartTime,
		attr.EndTime,
		attr.StayDuration,
		attr.Order,
		attr.Cost,
		transportationJSON,
		attr.Notes,
	)

	return err
}

// saveMeal 保存餐饮。
func (r *ItineraryRepositoryImpl) saveMeal(ctx context.Context, itineraryID string, dayNumber int, meal *entity.DayMeal) error {
	if meal == nil {
		return nil
	}
	restaurantID := ""
	if meal.Restaurant != nil {
		restaurantID = meal.Restaurant.ID
	}

	query := `
		INSERT INTO day_meals (
			id, itinerary_id, day_number, meal_type, restaurant_id, meal_time, cost, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		generateID(),
		itineraryID,
		dayNumber,
		meal.MealType,
		restaurantID,
		meal.Time,
		meal.Cost,
		meal.Notes,
	)
	return err
}

// saveHotel 保存酒店。
func (r *ItineraryRepositoryImpl) saveHotel(ctx context.Context, itineraryID string, dayNumber int, hotel *entity.DayHotel) error {
	if hotel == nil {
		return nil
	}
	query := `
		INSERT INTO day_hotels (
			id, itinerary_id, day_number, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		generateID(),
		itineraryID,
		dayNumber,
		hotel.HotelName,
		hotel.Address,
		hotel.CheckInTime,
		hotel.CheckOutTime,
		hotel.Cost,
		hotel.RoomType,
		hotel.Notes,
	)
	return err
}

// deleteDays 删除一个行程关联的日程明细表。
func (r *ItineraryRepositoryImpl) deleteDays(ctx context.Context, itineraryID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM day_attractions WHERE itinerary_id = $1`, itineraryID); err != nil {
		return fmt.Errorf("failed to delete attractions: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM day_meals WHERE itinerary_id = $1`, itineraryID); err != nil {
		return fmt.Errorf("failed to delete meals: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM day_hotels WHERE itinerary_id = $1`, itineraryID); err != nil {
		return fmt.Errorf("failed to delete hotels: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM itinerary_days WHERE itinerary_id = $1`, itineraryID); err != nil {
		return fmt.Errorf("failed to delete days: %w", err)
	}
	return nil
}

// scanItineraries 扫描行程列表并补全 days。
func (r *ItineraryRepositoryImpl) scanItineraries(rows *sql.Rows) ([]*entity.Itinerary, error) {
	var itineraries []*entity.Itinerary
	for rows.Next() {
		var it entity.Itinerary
		var budgetJSON, statisticsJSON, constraintsJSON []byte
		var publishedAt sql.NullTime

		err := rows.Scan(
			&it.ID,
			&it.UserID,
			&it.Name,
			&it.Description,
			&it.Status,
			&it.StartDate,
			&it.EndDate,
			&it.DayCount,
			&budgetJSON,
			&statisticsJSON,
			&constraintsJSON,
			&it.CreatedAt,
			&it.UpdatedAt,
			&publishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan itinerary: %w", err)
		}

		if len(budgetJSON) > 0 {
			_ = json.Unmarshal(budgetJSON, &it.Budget)
		}
		if len(statisticsJSON) > 0 {
			_ = json.Unmarshal(statisticsJSON, &it.Statistics)
		}
		if len(constraintsJSON) > 0 {
			_ = json.Unmarshal(constraintsJSON, &it.Constraints)
		}
		if publishedAt.Valid {
			it.PublishedAt = &publishedAt.Time
		}

		days, err := r.fetchDays(context.Background(), it.ID)
		if err == nil {
			it.Days = days
		}

		itineraries = append(itineraries, &it)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return itineraries, nil
}

// buildQuery 生成约束查询 SQL。
func (r *ItineraryRepositoryImpl) buildQuery(q *repository.ItineraryQuery) (string, []interface{}) {
	base := `
		SELECT id, user_id, name, description, status, start_date, end_date, day_count,
		       budget, statistics, constraints, created_at, updated_at, published_at
		FROM itineraries
		WHERE deleted_at IS NULL
	`
	if q == nil {
		return base + ` ORDER BY created_at DESC`, nil
	}

	var (
		conds []string
		args  []interface{}
	)
	add := func(cond string, val interface{}) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if q.UserID != nil && *q.UserID != "" {
		add("user_id = $%d", *q.UserID)
	}
	if q.Status != nil {
		add("status = $%d", *q.Status)
	}
	if q.Keyword != "" {
		args = append(args, "%"+q.Keyword+"%")
		idx := len(args)
		conds = append(conds, fmt.Sprintf("(name ILIKE $%d)", idx))
	}

	query := base
	if len(conds) > 0 {
		query += " AND " + strings.Join(conds, " AND ")
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit)
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return query, args
}

func generateID() string {
	return uuid.New().String()
}
