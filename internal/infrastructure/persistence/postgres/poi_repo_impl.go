package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	
	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// poiRepo 实现 repository.POIRepository 接口
type poiRepo struct {
	db *sql.DB
}

// NewPOIRepository 创建 POI 仓储实例
func NewPOIRepository(db *sql.DB) repository.POIRepository {
	return &poiRepo{db: db}
}

// ========== 基础 CRUD 实现 ==========

// FindByID 根据 ID 查询 POI
func (r *poiRepo) FindByID(ctx context.Context, id string) (*entity.POI, error) {
	query := `
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	var poi entity.POI
	var openingHoursJSON, bestTimeJSON, tagsJSON, featuresJSON, similarPOIsJSON []byte
	var imagesJSON, videosJSON, tipsJSON, warningsJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&poi.ID,
		&poi.Name,
		&poi.NameEn,
		&poi.NameLocal,
		&poi.Category,
		&poi.SubCategory,
		&poi.Location.Lat,
		&poi.Location.Lng,
		&poi.Address,
		&poi.City,
		&poi.District,
		&poi.GeoHash,
		&openingHoursJSON,
		&poi.AvgStayTime,
		&bestTimeJSON,
		&poi.Duration,
		&poi.PriceLevel,
		&poi.TicketPrice,
		&poi.AvgPrice,
		&poi.Rating,
		&poi.RatingCount,
		&poi.Popularity,
		&poi.Rank,
		&tagsJSON,
		&featuresJSON,
		&similarPOIsJSON,
		&imagesJSON,
		&poi.Thumbnail,
		&videosJSON,
		&poi.BookingURL,
		&poi.IsBookable,
		&poi.Inventory,
		&poi.Description,
		&tipsJSON,
		&warningsJSON,
		&poi.Source,
		&poi.Confidence,
		&poi.CreatedAt,
		&poi.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrPOINotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find poi by id: %w", err)
	}
	
	// 解析 JSON 字段
	if err := json.Unmarshal(openingHoursJSON, &poi.OpeningHours); err != nil {
		return nil, fmt.Errorf("failed to parse opening_hours: %w", err)
	}
	if err := json.Unmarshal(bestTimeJSON, &poi.BestTime); err != nil {
		return nil, fmt.Errorf("failed to parse best_time: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &poi.Tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}
	if err := json.Unmarshal(featuresJSON, &poi.Features); err != nil {
		return nil, fmt.Errorf("failed to parse features: %w", err)
	}
	if err := json.Unmarshal(similarPOIsJSON, &poi.SimilarPOIs); err != nil {
		return nil, fmt.Errorf("failed to parse similar_pois: %w", err)
	}
	if err := json.Unmarshal(imagesJSON, &poi.Images); err != nil {
		return nil, fmt.Errorf("failed to parse images: %w", err)
	}
	if err := json.Unmarshal(videosJSON, &poi.Videos); err != nil {
		return nil, fmt.Errorf("failed to parse videos: %w", err)
	}
	if err := json.Unmarshal(tipsJSON, &poi.Tips); err != nil {
		return nil, fmt.Errorf("failed to parse tips: %w", err)
	}
	if err := json.Unmarshal(warningsJSON, &poi.Warnings); err != nil {
		return nil, fmt.Errorf("failed to parse warnings: %w", err)
	}
	
	return &poi, nil
}

// FindByIDs 批量查询 POI
func (r *poiRepo) FindByIDs(ctx context.Context, ids []string) ([]*entity.POI, error) {
	if len(ids) == 0 {
		return []*entity.POI{}, nil
	}
	
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find pois by ids: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// Save 保存 POI（新增或更新）
func (r *poiRepo) Save(ctx context.Context, poi *entity.POI) error {
	// 检查是否存在
	existing, err := r.FindByID(ctx, poi.ID)
	if err != nil && err != entity.ErrPOINotFound {
		return err
	}
	
	if existing != nil {
		return r.Update(ctx, poi)
	}
	
	return r.insert(ctx, poi)
}

// insert 插入新 POI
func (r *poiRepo) insert(ctx context.Context, poi *entity.POI) error {
	openingHoursJSON, _ := json.Marshal(poi.OpeningHours)
	bestTimeJSON, _ := json.Marshal(poi.BestTime)
	tagsJSON, _ := json.Marshal(poi.Tags)
	featuresJSON, _ := json.Marshal(poi.Features)
	similarPOIsJSON, _ := json.Marshal(poi.SimilarPOIs)
	imagesJSON, _ := json.Marshal(poi.Images)
	videosJSON, _ := json.Marshal(poi.Videos)
	tipsJSON, _ := json.Marshal(poi.Tips)
	warningsJSON, _ := json.Marshal(poi.Warnings)
	
	query := `
		INSERT INTO pois (
			id, name, name_en, name_local, category, sub_category,
			lat, lng, address, city, district, geohash,
			opening_hours, avg_stay_time, best_time, duration,
			price_level, ticket_price, avg_price,
			rating, rating_count, popularity, rank,
			tags, features, similar_pois,
			images, thumbnail, videos,
			booking_url, is_bookable, inventory,
			description, tips, warnings,
			source, confidence, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		poi.ID,
		poi.Name,
		poi.NameEn,
		poi.NameLocal,
		poi.Category,
		poi.SubCategory,
		poi.Location.Lat,
		poi.Location.Lng,
		poi.Address,
		poi.City,
		poi.District,
		poi.GeoHash,
		openingHoursJSON,
		poi.AvgStayTime,
		bestTimeJSON,
		poi.Duration,
		poi.PriceLevel,
		poi.TicketPrice,
		poi.AvgPrice,
		poi.Rating,
		poi.RatingCount,
		poi.Popularity,
		poi.Rank,
		tagsJSON,
		featuresJSON,
		similarPOIsJSON,
		imagesJSON,
		poi.Thumbnail,
		videosJSON,
		poi.BookingURL,
		poi.IsBookable,
		poi.Inventory,
		poi.Description,
		tipsJSON,
		warningsJSON,
		poi.Source,
		poi.Confidence,
		poi.CreatedAt,
		poi.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert poi: %w", err)
	}
	
	return nil
}

// Update 更新 POI
func (r *poiRepo) Update(ctx context.Context, poi *entity.POI) error {
	openingHoursJSON, _ := json.Marshal(poi.OpeningHours)
	bestTimeJSON, _ := json.Marshal(poi.BestTime)
	tagsJSON, _ := json.Marshal(poi.Tags)
	featuresJSON, _ := json.Marshal(poi.Features)
	similarPOIsJSON, _ := json.Marshal(poi.SimilarPOIs)
	imagesJSON, _ := json.Marshal(poi.Images)
	videosJSON, _ := json.Marshal(poi.Videos)
	tipsJSON, _ := json.Marshal(poi.Tips)
	warningsJSON, _ := json.Marshal(poi.Warnings)
	
	query := `
		UPDATE pois SET
			name = $2,
			name_en = $3,
			name_local = $4,
			category = $5,
			sub_category = $6,
			lat = $7,
			lng = $8,
			address = $9,
			city = $10,
			district = $11,
			geohash = $12,
			opening_hours = $13,
			avg_stay_time = $14,
			best_time = $15,
			duration = $16,
			price_level = $17,
			ticket_price = $18,
			avg_price = $19,
			rating = $20,
			rating_count = $21,
			popularity = $22,
			rank = $23,
			tags = $24,
			features = $25,
			similar_pois = $26,
			images = $27,
			thumbnail = $28,
			videos = $29,
			booking_url = $30,
			is_bookable = $31,
			inventory = $32,
			description = $33,
			tips = $34,
			warnings = $35,
			source = $36,
			confidence = $37,
			updated_at = $38
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query,
		poi.ID,
		poi.Name,
		poi.NameEn,
		poi.NameLocal,
		poi.Category,
		poi.SubCategory,
		poi.Location.Lat,
		poi.Location.Lng,
		poi.Address,
		poi.City,
		poi.District,
		poi.GeoHash,
		openingHoursJSON,
		poi.AvgStayTime,
		bestTimeJSON,
		poi.Duration,
		poi.PriceLevel,
		poi.TicketPrice,
		poi.AvgPrice,
		poi.Rating,
		poi.RatingCount,
		poi.Popularity,
		poi.Rank,
		tagsJSON,
		featuresJSON,
		similarPOIsJSON,
		imagesJSON,
		poi.Thumbnail,
		videosJSON,
		poi.BookingURL,
		poi.IsBookable,
		poi.Inventory,
		poi.Description,
		tipsJSON,
		warningsJSON,
		poi.Source,
		poi.Confidence,
		time.Now(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to update poi: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrPOINotFound
	}
	
	return nil
}

// Delete 软删除 POI
func (r *poiRepo) Delete(ctx context.Context, id string) error {
	query := `UPDATE pois SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete poi: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrPOINotFound
	}
	
	return nil
}

// ========== 分类查询实现 ==========

// FindByCategory 根据分类查询 POI 列表
func (r *poiRepo) FindByCategory(ctx context.Context, category entity.Category, city string, limit, offset int) ([]*entity.POI, error) {
	query := `
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE category = $1 AND city = $2 AND deleted_at IS NULL
		ORDER BY rating DESC, popularity DESC
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.db.QueryContext(ctx, query, category, city, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find pois by category: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// FindByCity 根据城市查询 POI
func (r *poiRepo) FindByCity(ctx context.Context, city string, limit, offset int) ([]*entity.POI, error) {
	query := `
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE city = $1 AND deleted_at IS NULL
		ORDER BY rating DESC, popularity DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, city, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find pois by city: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// FindByTags 根据标签查询 POI
func (r *poiRepo) FindByTags(ctx context.Context, tags []string, city string, limit int) ([]*entity.POI, error) {
	if len(tags) == 0 {
		return r.FindByCity(ctx, city, limit, 0)
	}
	
	// 使用 JSONB 查询，检查 tags 数组中是否包含任意指定标签
	placeholders := make([]string, len(tags))
	args := make([]interface{}, len(tags)+2)
	args[0] = city
	args[len(args)-1] = limit
	for i, tag := range tags {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = tag
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE city = $1 
		  AND deleted_at IS NULL
		  AND EXISTS (
		      SELECT 1 FROM jsonb_array_elements(tags) AS t
		      WHERE t->>'id' IN (%s) OR t->>'name' IN (%s)
		  )
		ORDER BY rating DESC
		LIMIT $%d
	`, strings.Join(placeholders, ","), strings.Join(placeholders, ","), len(args))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find pois by tags: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// FindNearby 查询附近的 POI
func (r *poiRepo) FindNearby(ctx context.Context, lat, lng float64, radius int, category *entity.Category, limit int) ([]*entity.POI, error) {
	// 使用 PostGIS 的地球距离计算
	// 这里简化为经纬度范围查询（实际应使用 PostGIS 的 ST_DWithin）
	// 1度 ≈ 111公里
	degreeRadius := float64(radius) / 111.0
	
	var categoryFilter string
	var args []interface{}
	
	if category != nil {
		categoryFilter = "AND category = $3"
		args = []interface{}{lat - degreeRadius, lat + degreeRadius, lng - degreeRadius, lng + degreeRadius, *category, limit}
	} else {
		categoryFilter = ""
		args = []interface{}{lat - degreeRadius, lat + degreeRadius, lng - degreeRadius, lng + degreeRadius, limit}
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at,
		       (6371 * acos(cos(radians($1)) * cos(radians(lat)) * cos(radians(lng) - radians($3)) + sin(radians($1)) * sin(radians(lat)))) AS distance
		FROM pois
		WHERE lat BETWEEN $1 AND $2
		  AND lng BETWEEN $3 AND $4
		  AND deleted_at IS NULL
		  %s
		ORDER BY distance
		LIMIT $%d
	`, categoryFilter, len(args))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby pois: %w", err)
	}
	defer rows.Close()
	
	// 扫描时忽略 distance 字段
	return r.scanPOIs(rows)
}

// ========== 高级查询 ==========

// FindByQuery 复杂条件查询
func (r *poiRepo) FindByQuery(ctx context.Context, query *repository.POIQuery) ([]*entity.POI, error) {
	sql, args := r.buildQuery(query)
	
	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find pois by query: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// ========== 热门推荐 ==========

// GetPopularPOIs 获取热门 POI
func (r *poiRepo) GetPopularPOIs(ctx context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error) {
	var categoryFilter string
	var args []interface{}
	
	if category != nil {
		categoryFilter = "AND category = $2"
		args = []interface{}{city, *category, limit}
	} else {
		categoryFilter = ""
		args = []interface{}{city, limit}
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE city = $1
		  AND deleted_at IS NULL
		  %s
		ORDER BY popularity DESC, rating DESC
		LIMIT $%d
	`, categoryFilter, len(args))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular pois: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// GetTopRatedPOIs 获取高分 POI
func (r *poiRepo) GetTopRatedPOIs(ctx context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error) {
	var categoryFilter string
	var args []interface{}
	
	if category != nil {
		categoryFilter = "AND category = $2"
		args = []interface{}{city, *category, limit}
	} else {
		categoryFilter = ""
		args = []interface{}{city, limit}
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE city = $1
		  AND deleted_at IS NULL
		  AND rating >= 4.0
		  %s
		ORDER BY rating DESC, rating_count DESC
		LIMIT $%d
	`, categoryFilter, len(args))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top rated pois: %w", err)
	}
	defer rows.Close()
	
	return r.scanPOIs(rows)
}

// ========== 距离矩阵 ==========

// GetDistanceMatrix 获取 POI 之间的距离矩阵
func (r *poiRepo) GetDistanceMatrix(ctx context.Context, poiIDs []string) (*repository.DistanceMatrix, error) {
	if len(poiIDs) == 0 {
		return &repository.DistanceMatrix{
			Matrix:   make(map[string]map[string]float64),
			Duration: make(map[string]map[string]int),
		}, nil
	}
	
	// 先获取所有 POI 的坐标
	placeholders := make([]string, len(poiIDs))
	args := make([]interface{}, len(poiIDs))
	for i, id := range poiIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	
	coordQuery := fmt.Sprintf(`
		SELECT id, lat, lng FROM pois
		WHERE id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, coordQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinates: %w", err)
	}
	defer rows.Close()
	
	// 存储坐标
	coords := make(map[string]struct{ Lat, Lng float64 })
	for rows.Next() {
		var id string
		var lat, lng float64
		if err := rows.Scan(&id, &lat, &lng); err != nil {
			return nil, err
		}
		coords[id] = struct{ Lat, Lng float64 }{Lat: lat, Lng: lng}
	}
	
	// 计算距离矩阵
	matrix := &repository.DistanceMatrix{
		Matrix:   make(map[string]map[string]float64),
		Duration: make(map[string]map[string]int),
	}
	
	for _, fromID := range poiIDs {
		fromCoord, ok := coords[fromID]
		if !ok {
			continue
		}
		
		matrix.Matrix[fromID] = make(map[string]float64)
		matrix.Duration[fromID] = make(map[string]int)
		
		for _, toID := range poiIDs {
			toCoord, ok := coords[toID]
			if !ok {
				continue
			}
			
			if fromID == toID {
				matrix.Matrix[fromID][toID] = 0
				matrix.Duration[fromID][toID] = 0
				continue
			}
			
			// 计算距离（简化版 Haversine）
			distance := haversine(fromCoord.Lat, fromCoord.Lng, toCoord.Lat, toCoord.Lng)
			matrix.Matrix[fromID][toID] = distance
			
			// 估算交通时间（假设平均速度 30 km/h）
			matrix.Duration[fromID][toID] = int(distance / 30 * 60)
		}
	}
	
	return matrix, nil
}

// ========== 统计 ==========

// CountByCity 统计城市 POI 数量
func (r *poiRepo) CountByCity(ctx context.Context, city string) (int, error) {
	query := `SELECT COUNT(*) FROM pois WHERE city = $1 AND deleted_at IS NULL`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, city).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pois by city: %w", err)
	}
	
	return count, nil
}

// CountByCategory 统计分类 POI 数量
func (r *poiRepo) CountByCategory(ctx context.Context, category entity.Category, city string) (int, error) {
	query := `SELECT COUNT(*) FROM pois WHERE category = $1 AND city = $2 AND deleted_at IS NULL`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, category, city).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pois by category: %w", err)
	}
	
	return count, nil
}

// ========== 存在性检查 ==========

// ExistsByID 检查 POI 是否存在
func (r *poiRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pois WHERE id = $1 AND deleted_at IS NULL)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check poi existence: %w", err)
	}
	
	return exists, nil
}

// ========== 私有辅助方法 ==========

// scanPOIs 扫描多行 POI 数据
func (r *poiRepo) scanPOIs(rows *sql.Rows) ([]*entity.POI, error) {
	var pois []*entity.POI
	
	for rows.Next() {
		poi, err := r.scanPOI(rows)
		if err != nil {
			return nil, err
		}
		pois = append(pois, poi)
	}
	
	return pois, nil
}

// scanPOI 扫描单行 POI 数据
func (r *poiRepo) scanPOI(rows *sql.Rows) (*entity.POI, error) {
	var poi entity.POI
	var openingHoursJSON, bestTimeJSON, tagsJSON, featuresJSON, similarPOIsJSON []byte
	var imagesJSON, videosJSON, tipsJSON, warningsJSON []byte
	
	err := rows.Scan(
		&poi.ID,
		&poi.Name,
		&poi.NameEn,
		&poi.NameLocal,
		&poi.Category,
		&poi.SubCategory,
		&poi.Location.Lat,
		&poi.Location.Lng,
		&poi.Address,
		&poi.City,
		&poi.District,
		&poi.GeoHash,
		&openingHoursJSON,
		&poi.AvgStayTime,
		&bestTimeJSON,
		&poi.Duration,
		&poi.PriceLevel,
		&poi.TicketPrice,
		&poi.AvgPrice,
		&poi.Rating,
		&poi.RatingCount,
		&poi.Popularity,
		&poi.Rank,
		&tagsJSON,
		&featuresJSON,
		&similarPOIsJSON,
		&imagesJSON,
		&poi.Thumbnail,
		&videosJSON,
		&poi.BookingURL,
		&poi.IsBookable,
		&poi.Inventory,
		&poi.Description,
		&tipsJSON,
		&warningsJSON,
		&poi.Source,
		&poi.Confidence,
		&poi.CreatedAt,
		&poi.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// 解析 JSON 字段
	json.Unmarshal(openingHoursJSON, &poi.OpeningHours)
	json.Unmarshal(bestTimeJSON, &poi.BestTime)
	json.Unmarshal(tagsJSON, &poi.Tags)
	json.Unmarshal(featuresJSON, &poi.Features)
	json.Unmarshal(similarPOIsJSON, &poi.SimilarPOIs)
	json.Unmarshal(imagesJSON, &poi.Images)
	json.Unmarshal(videosJSON, &poi.Videos)
	json.Unmarshal(tipsJSON, &poi.Tips)
	json.Unmarshal(warningsJSON, &poi.Warnings)
	
	return &poi, nil
}

// buildQuery 构建动态查询
func (r *poiRepo) buildQuery(query *repository.POIQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	// 城市筛选
	if query.City != nil {
		conditions = append(conditions, fmt.Sprintf("city = $%d", argIndex))
		args = append(args, *query.City)
		argIndex++
	}
	
	// 分类筛选
	if query.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *query.Category)
		argIndex++
	}
	
	// 关键词搜索
	if query.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR name_en ILIKE $%d OR address ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+query.Keyword+"%")
		argIndex++
	}
	
	// 最低评分
	if query.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", argIndex))
		args = append(args, *query.MinRating)
		argIndex++
	}
	
	// 价格等级
	if query.PriceLevel != nil {
		conditions = append(conditions, fmt.Sprintf("price_level = $%d", argIndex))
		args = append(args, *query.PriceLevel)
		argIndex++
	}
	
	// 最大价格
	if query.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("ticket_price <= $%d", argIndex))
		args = append(args, *query.MaxPrice)
		argIndex++
	}
	
	// 基础条件
	baseSQL := `
		SELECT id, name, name_en, name_local, category, sub_category,
		       lat, lng, address, city, district, geohash,
		       opening_hours, avg_stay_time, best_time, duration,
		       price_level, ticket_price, avg_price,
		       rating, rating_count, popularity, rank,
		       tags, features, similar_pois,
		       images, thumbnail, videos,
		       booking_url, is_bookable, inventory,
		       description, tips, warnings,
		       source, confidence, created_at, updated_at
		FROM pois
		WHERE deleted_at IS NULL
	`
	
	if len(conditions) > 0 {
		baseSQL += " AND " + strings.Join(conditions, " AND ")
	}
	
	// 排序
	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}
	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	baseSQL += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	// 分页
	if query.Limit > 0 {
		baseSQL += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, query.Limit)
		argIndex++
	}
	if query.Offset > 0 {
		baseSQL += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, query.Offset)
	}
	
	return baseSQL, args
}

// haversine 计算两点之间的距离（公里）
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // 地球半径（公里）
	
	lat1Rad := lat1 * 3.1415926 / 180
	lat2Rad := lat2 * 3.1415926 / 180
	dLat := (lat2 - lat1) * 3.1415926 / 180
	dLng := (lng2 - lng1) * 3.1415926 / 180
	
	a := sin(dLat/2)*sin(dLat/2) +
		cos(lat1Rad)*cos(lat2Rad)*
			sin(dLng/2)*sin(dLng/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))
	
	return R * c
}

// 辅助数学函数（避免导入 math 包）
func sin(x float64) float64 {
	return x - x*x*x/6 + x*x*x*x*x/120
}

func cos(x float64) float64 {
	return 1 - x*x/2 + x*x*x*x/24
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func atan2(y, x float64) float64 {
	if x > 0 {
		return atan(y / x)
	}
	if x < 0 {
		if y >= 0 {
			return atan(y/x) + 3.1415926
		}
		return atan(y/x) - 3.1415926
	}
	if y > 0 {
		return 3.1415926 / 2
	}
	if y < 0 {
		return -3.1415926 / 2
	}
	return 0
}

func atan(x float64) float64 {
	return x - x*x*x/3 + x*x*x*x*x/5
}
