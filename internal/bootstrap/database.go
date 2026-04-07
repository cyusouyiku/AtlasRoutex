package bootstrap

import (
	"context"
	"sort"
	"strings"
	"time"

	appplanner "atlas-routex/internal/application/planner"
	apprecommender "atlas-routex/internal/application/recommender"
	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// DatabaseSet 聚合当前应用使用的仓储与用例。
type DatabaseSet struct {
	Users            repository.UserRepository
	POIs             repository.POIRepository
	Itineraries      repository.ItineraryRepository
	PlanUsecase      *appplanner.PlanUsecase
	AdjustUsecase    *appplanner.AdjustUsecase
	RecommendUsecase *apprecommender.RecommendUsecase
}

func InitDatabases() (*DatabaseSet, error) {
	users := newMemoryUserRepository(seedUsers())
	pois := newMemoryPOIRepository(seedPOIs())
	itineraries := newMemoryItineraryRepository()
	solver := &defaultTripSolver{}

	planUC := appplanner.NewPlanUsecase(users, pois, itineraries, solver)
	adjustUC := appplanner.NewAdjustUsecase(pois, itineraries, solver)
	recommendUC := apprecommender.NewRecommendUsecase(pois)

	return &DatabaseSet{
		Users:            users,
		POIs:             pois,
		Itineraries:      itineraries,
		PlanUsecase:      planUC,
		AdjustUsecase:    adjustUC,
		RecommendUsecase: recommendUC,
	}, nil
}

// ---- default solver ----

type defaultTripSolver struct{}

func (s *defaultTripSolver) Name() string { return "default-greedy" }

func (s *defaultTripSolver) Solve(_ context.Context, in *appplanner.SolverInput) (*appplanner.SolverOutput, error) {
	if in == nil {
		return &appplanner.SolverOutput{Days: [][]string{}}, nil
	}
	dayCount := in.DayCount
	if dayCount <= 0 {
		dayCount = 1
	}
	candidates := append([]*entity.POI(nil), in.Candidates...)
	sort.SliceStable(candidates, func(i, j int) bool {
		scoreI := candidates[i].Rating*10 + candidates[i].Popularity
		scoreJ := candidates[j].Rating*10 + candidates[j].Popularity
		if scoreI == scoreJ {
			return candidates[i].Name < candidates[j].Name
		}
		return scoreI > scoreJ
	})

	days := make([][]string, dayCount)
	remaining := in.TotalBudget
	for _, poi := range candidates {
		if poi == nil {
			continue
		}
		cost := poi.TicketPrice
		if poi.Category == entity.CategoryRestaurant && poi.AvgPrice > 0 {
			cost = poi.AvgPrice
		} else if cost <= 0 {
			cost = poi.AvgPrice
		}
		if in.TotalBudget > 0 && cost > 0 && remaining-cost < -0.0001 {
			continue
		}
		target := 0
		for i := 1; i < len(days); i++ {
			if len(days[i]) < len(days[target]) {
				target = i
			}
		}
		days[target] = append(days[target], poi.ID)
		if in.TotalBudget > 0 && cost > 0 {
			remaining -= cost
		}
	}

	return &appplanner.SolverOutput{Days: days}, nil
}

// ---- memory user repository ----

type memoryUserRepository struct {
	items map[string]*entity.User
}

func newMemoryUserRepository(users []*entity.User) *memoryUserRepository {
	m := &memoryUserRepository{items: make(map[string]*entity.User, len(users))}
	for _, u := range users {
		if u != nil {
			copy := *u
			m.items[u.ID] = &copy
		}
	}
	return m
}

func (m *memoryUserRepository) FindByID(_ context.Context, id string) (*entity.User, error) {
	if u, ok := m.items[id]; ok {
		copy := *u
		return &copy, nil
	}
	return nil, nil
}

func (m *memoryUserRepository) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	for _, u := range m.items {
		if strings.EqualFold(u.Email, email) {
			copy := *u
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *memoryUserRepository) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	out := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		if u, ok := m.items[id]; ok {
			copy := *u
			out = append(out, &copy)
		}
	}
	return out, nil
}

func (m *memoryUserRepository) Save(_ context.Context, user *entity.User) error {
	copy := *user
	m.items[user.ID] = &copy
	return nil
}
func (m *memoryUserRepository) Update(ctx context.Context, user *entity.User) error {
	return m.Save(ctx, user)
}
func (m *memoryUserRepository) Delete(_ context.Context, id string) error {
	delete(m.items, id)
	return nil
}
func (m *memoryUserRepository) FindByRole(_ context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error) {
	return m.filterUsers(func(u *entity.User) bool { return u.Role == role }, limit, offset), nil
}
func (m *memoryUserRepository) FindByStatus(_ context.Context, status entity.UserStatus, limit, offset int) ([]*entity.User, error) {
	return m.filterUsers(func(u *entity.User) bool { return u.Status == status }, limit, offset), nil
}
func (m *memoryUserRepository) FindActiveUsers(_ context.Context, limit int) ([]*entity.User, error) {
	return m.filterUsers(func(u *entity.User) bool { return u.Status == entity.UserStatusActive }, limit, 0), nil
}
func (m *memoryUserRepository) FindByPreferences(_ context.Context, preferences entity.UserPreferences, limit int) ([]*entity.User, error) {
	return m.filterUsers(func(u *entity.User) bool {
		if len(preferences.PreferredCategories) == 0 {
			return true
		}
		for _, want := range preferences.PreferredCategories {
			for _, got := range u.Preferences.PreferredCategories {
				if strings.EqualFold(want, got) {
					return true
				}
			}
		}
		return false
	}, limit, 0), nil
}
func (m *memoryUserRepository) CountByRole(_ context.Context, role entity.UserRole) (int, error) {
	return len(m.filterUsers(func(u *entity.User) bool { return u.Role == role }, 0, 0)), nil
}
func (m *memoryUserRepository) CountByStatus(_ context.Context, status entity.UserStatus) (int, error) {
	return len(m.filterUsers(func(u *entity.User) bool { return u.Status == status }, 0, 0)), nil
}
func (m *memoryUserRepository) CountActiveLastDays(_ context.Context, days int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	count := 0
	for _, u := range m.items {
		if u.LastLoginAt != nil && u.LastLoginAt.After(cutoff) {
			count++
		}
	}
	return count, nil
}
func (m *memoryUserRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	u, _ := m.FindByEmail(context.Background(), email)
	return u != nil, nil
}
func (m *memoryUserRepository) ExistsByPhone(_ context.Context, phone string) (bool, error) {
	for _, u := range m.items {
		if u.Phone == phone {
			return true, nil
		}
	}
	return false, nil
}
func (m *memoryUserRepository) BatchUpdateStatus(_ context.Context, ids []string, status entity.UserStatus) error {
	for _, id := range ids {
		if u, ok := m.items[id]; ok {
			u.Status = status
			u.UpdatedAt = time.Now()
		}
	}
	return nil
}
func (m *memoryUserRepository) BatchUpdateRole(_ context.Context, ids []string, role entity.UserRole) error {
	for _, id := range ids {
		if u, ok := m.items[id]; ok {
			u.Role = role
			u.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (m *memoryUserRepository) filterUsers(fn func(*entity.User) bool, limit, offset int) []*entity.User {
	list := make([]*entity.User, 0)
	for _, u := range m.items {
		if fn(u) {
			copy := *u
			list = append(list, &copy)
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	if offset > len(list) {
		return []*entity.User{}
	}
	list = list[offset:]
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}

// ---- memory poi repository ----

type memoryPOIRepository struct {
	items map[string]*entity.POI
}

func newMemoryPOIRepository(pois []*entity.POI) *memoryPOIRepository {
	m := &memoryPOIRepository{items: make(map[string]*entity.POI, len(pois))}
	for _, p := range pois {
		if p != nil {
			m.items[p.ID] = p.Clone()
		}
	}
	return m
}

func (m *memoryPOIRepository) FindByID(_ context.Context, id string) (*entity.POI, error) {
	if p, ok := m.items[id]; ok {
		return p.Clone(), nil
	}
	return nil, nil
}
func (m *memoryPOIRepository) FindByIDs(_ context.Context, ids []string) ([]*entity.POI, error) {
	out := make([]*entity.POI, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.items[id]; ok {
			out = append(out, p.Clone())
		}
	}
	return out, nil
}
func (m *memoryPOIRepository) Save(_ context.Context, poi *entity.POI) error {
	m.items[poi.ID] = poi.Clone()
	return nil
}
func (m *memoryPOIRepository) Update(ctx context.Context, poi *entity.POI) error {
	return m.Save(ctx, poi)
}
func (m *memoryPOIRepository) Delete(_ context.Context, id string) error {
	delete(m.items, id)
	return nil
}
func (m *memoryPOIRepository) FindByCategory(_ context.Context, category entity.Category, city string, limit, offset int) ([]*entity.POI, error) {
	return m.filterPOIs(func(p *entity.POI) bool {
		return p.Category == category && (city == "" || strings.EqualFold(p.City, city))
	}, limit, offset), nil
}
func (m *memoryPOIRepository) FindByCity(_ context.Context, city string, limit, offset int) ([]*entity.POI, error) {
	return m.filterPOIs(func(p *entity.POI) bool { return strings.EqualFold(p.City, city) }, limit, offset), nil
}
func (m *memoryPOIRepository) FindByTags(_ context.Context, tags []string, city string, limit int) ([]*entity.POI, error) {
	return m.filterPOIs(func(p *entity.POI) bool { return (city == "" || strings.EqualFold(p.City, city)) && p.MatchTags(tags) }, limit, 0), nil
}
func (m *memoryPOIRepository) FindNearby(_ context.Context, lat, lng float64, radius int, category *entity.Category, limit int) ([]*entity.POI, error) {
	origin := entity.Location{Lat: lat, Lng: lng}
	return m.filterPOIs(func(p *entity.POI) bool {
		if category != nil && *category != "" && p.Category != *category {
			return false
		}
		return origin.DistanceTo(p.Location) <= float64(radius)
	}, limit, 0), nil
}
func (m *memoryPOIRepository) FindByQuery(_ context.Context, q *repository.POIQuery) ([]*entity.POI, error) {
	if q == nil {
		return m.filterPOIs(func(*entity.POI) bool { return true }, 20, 0), nil
	}
	return m.filterPOIs(func(p *entity.POI) bool {
		if q.City != nil && *q.City != "" && !strings.EqualFold(p.City, *q.City) {
			return false
		}
		if q.Category != nil && *q.Category != "" && p.Category != *q.Category {
			return false
		}
		if q.Keyword != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(q.Keyword)) && !strings.Contains(strings.ToLower(p.Address), strings.ToLower(q.Keyword)) {
			return false
		}
		if q.MinRating != nil && p.Rating < *q.MinRating {
			return false
		}
		if q.MaxPrice != nil {
			price := p.TicketPrice
			if price <= 0 {
				price = p.AvgPrice
			}
			if price > *q.MaxPrice {
				return false
			}
		}
		if q.PriceLevel != nil && *q.PriceLevel != "" && p.PriceLevel != *q.PriceLevel {
			return false
		}
		if len(q.Tags) > 0 && !p.MatchTags(q.Tags) {
			return false
		}
		if len(q.TagsAll) > 0 && !p.MatchAllTags(q.TagsAll) {
			return false
		}
		if q.Nearby != nil {
			origin := entity.Location{Lat: q.Nearby.Lat, Lng: q.Nearby.Lng}
			if origin.DistanceTo(p.Location) > float64(q.Nearby.Radius) {
				return false
			}
		}
		return true
	}, q.Limit, q.Offset), nil
}
func (m *memoryPOIRepository) GetPopularPOIs(_ context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error) {
	return m.filterPOIs(func(p *entity.POI) bool {
		return (city == "" || strings.EqualFold(p.City, city)) && (category == nil || *category == "" || p.Category == *category)
	}, limit, 0), nil
}
func (m *memoryPOIRepository) GetTopRatedPOIs(ctx context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error) {
	return m.GetPopularPOIs(ctx, city, category, limit)
}
func (m *memoryPOIRepository) GetDistanceMatrix(_ context.Context, poiIDs []string) (*repository.DistanceMatrix, error) {
	matrix := &repository.DistanceMatrix{Matrix: map[string]map[string]float64{}, Duration: map[string]map[string]int{}}
	for _, fromID := range poiIDs {
		from, ok := m.items[fromID]
		if !ok {
			continue
		}
		matrix.Matrix[fromID] = map[string]float64{}
		matrix.Duration[fromID] = map[string]int{}
		for _, toID := range poiIDs {
			to, ok := m.items[toID]
			if !ok {
				continue
			}
			d := from.Location.DistanceTo(to.Location)
			matrix.Matrix[fromID][toID] = d
			matrix.Duration[fromID][toID] = int(d / 4 * 60)
		}
	}
	return matrix, nil
}
func (m *memoryPOIRepository) CountByCity(_ context.Context, city string) (int, error) {
	return len(m.filterPOIs(func(p *entity.POI) bool { return strings.EqualFold(p.City, city) }, 0, 0)), nil
}
func (m *memoryPOIRepository) CountByCategory(_ context.Context, category entity.Category, city string) (int, error) {
	return len(m.filterPOIs(func(p *entity.POI) bool {
		return p.Category == category && (city == "" || strings.EqualFold(p.City, city))
	}, 0, 0)), nil
}
func (m *memoryPOIRepository) ExistsByID(_ context.Context, id string) (bool, error) {
	_, ok := m.items[id]
	return ok, nil
}

func (m *memoryPOIRepository) filterPOIs(fn func(*entity.POI) bool, limit, offset int) []*entity.POI {
	list := make([]*entity.POI, 0)
	for _, p := range m.items {
		if fn(p) {
			list = append(list, p.Clone())
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		si := list[i].Rating*10 + list[i].Popularity
		sj := list[j].Rating*10 + list[j].Popularity
		if si == sj {
			return list[i].Name < list[j].Name
		}
		return si > sj
	})
	if offset > len(list) {
		return []*entity.POI{}
	}
	list = list[offset:]
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}

// ---- memory itinerary repository ----

type memoryItineraryRepository struct {
	items map[string]*entity.Itinerary
}

func newMemoryItineraryRepository() *memoryItineraryRepository {
	return &memoryItineraryRepository{items: map[string]*entity.Itinerary{}}
}

func (m *memoryItineraryRepository) FindByID(_ context.Context, id string) (*entity.Itinerary, error) {
	if it, ok := m.items[id]; ok {
		return it, nil
	}
	return nil, nil
}
func (m *memoryItineraryRepository) FindByUserID(_ context.Context, userID string) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool { return it.UserID == userID }, 0), nil
}
func (m *memoryItineraryRepository) FindByName(_ context.Context, name string) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		return strings.Contains(strings.ToLower(it.Name), strings.ToLower(name))
	}, 0), nil
}
func (m *memoryItineraryRepository) Save(_ context.Context, itinerary *entity.Itinerary) error {
	m.items[itinerary.ID] = itinerary
	return nil
}
func (m *memoryItineraryRepository) Update(ctx context.Context, itinerary *entity.Itinerary) error {
	return m.Save(ctx, itinerary)
}
func (m *memoryItineraryRepository) Delete(_ context.Context, id string) error {
	delete(m.items, id)
	return nil
}
func (m *memoryItineraryRepository) HardDelete(ctx context.Context, id string) error {
	return m.Delete(ctx, id)
}
func (m *memoryItineraryRepository) FindByStatus(_ context.Context, status entity.ItineraryStatus) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool { return it.Status == status }, 0), nil
}
func (m *memoryItineraryRepository) FindByUserIDAndStatus(_ context.Context, userID string, status entity.ItineraryStatus) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool { return it.UserID == userID && it.Status == status }, 0), nil
}
func (m *memoryItineraryRepository) FindByDateRange(_ context.Context, startDate, endDate time.Time) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool { return !it.StartDate.Before(startDate) && !it.EndDate.After(endDate) }, 0), nil
}
func (m *memoryItineraryRepository) FindByUserIDAndDateRange(_ context.Context, userID string, startDate, endDate time.Time) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		return it.UserID == userID && !it.StartDate.Before(startDate) && !it.EndDate.After(endDate)
	}, 0), nil
}
func (m *memoryItineraryRepository) FindByBudgetRange(_ context.Context, minBudget, maxBudget float64) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		return it.Budget != nil && it.Budget.TotalCost >= minBudget && it.Budget.TotalCost <= maxBudget
	}, 0), nil
}
func (m *memoryItineraryRepository) FindByCity(_ context.Context, city string, limit int) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		for _, d := range it.Days {
			for _, a := range d.Attractions {
				if a != nil && a.POI != nil && strings.EqualFold(a.POI.City, city) {
					return true
				}
			}
		}
		return false
	}, limit), nil
}
func (m *memoryItineraryRepository) FindByAttraction(_ context.Context, attractionID string) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		for _, d := range it.Days {
			for _, a := range d.Attractions {
				if a != nil && a.POI != nil && a.POI.ID == attractionID {
					return true
				}
			}
		}
		return false
	}, 0), nil
}
func (m *memoryItineraryRepository) FindRecentItineraries(_ context.Context, limit int) ([]*entity.Itinerary, error) {
	return m.filterItineraries(func(*entity.Itinerary) bool { return true }, limit), nil
}
func (m *memoryItineraryRepository) FindPopularItineraries(ctx context.Context, limit int) ([]*entity.Itinerary, error) {
	return m.FindRecentItineraries(ctx, limit)
}
func (m *memoryItineraryRepository) FindByConstraints(_ context.Context, q *repository.ItineraryQuery) ([]*entity.Itinerary, error) {
	if q == nil {
		return m.filterItineraries(func(*entity.Itinerary) bool { return true }, 0), nil
	}
	return m.filterItineraries(func(it *entity.Itinerary) bool {
		if q.UserID != nil && *q.UserID != "" && it.UserID != *q.UserID {
			return false
		}
		if q.Status != nil && *q.Status != "" && it.Status != *q.Status {
			return false
		}
		if q.Keyword != "" && !strings.Contains(strings.ToLower(it.Name), strings.ToLower(q.Keyword)) {
			return false
		}
		return true
	}, q.Limit), nil
}
func (m *memoryItineraryRepository) CountByUserID(_ context.Context, userID string) (int, error) {
	return len(m.filterItineraries(func(it *entity.Itinerary) bool { return it.UserID == userID }, 0)), nil
}
func (m *memoryItineraryRepository) CountByCity(ctx context.Context, city string) (int, error) {
	list, _ := m.FindByCity(ctx, city, 0)
	return len(list), nil
}
func (m *memoryItineraryRepository) CountByStatus(_ context.Context, status entity.ItineraryStatus) (int, error) {
	return len(m.filterItineraries(func(it *entity.Itinerary) bool { return it.Status == status }, 0)), nil
}
func (m *memoryItineraryRepository) CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error) {
	list, _ := m.FindByDateRange(ctx, startDate, endDate)
	return len(list), nil
}
func (m *memoryItineraryRepository) ExistsByUserIDAndName(_ context.Context, userID, name string) (bool, error) {
	for _, it := range m.items {
		if it.UserID == userID && strings.EqualFold(it.Name, name) {
			return true, nil
		}
	}
	return false, nil
}
func (m *memoryItineraryRepository) BatchUpdateStatus(_ context.Context, ids []string, status entity.ItineraryStatus) error {
	for _, id := range ids {
		if it, ok := m.items[id]; ok {
			it.Status = status
			it.UpdatedAt = time.Now()
		}
	}
	return nil
}
func (m *memoryItineraryRepository) BatchDelete(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.items, id)
	}
	return nil
}

func (m *memoryItineraryRepository) filterItineraries(fn func(*entity.Itinerary) bool, limit int) []*entity.Itinerary {
	list := make([]*entity.Itinerary, 0)
	for _, it := range m.items {
		if fn(it) {
			list = append(list, it)
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}

// ---- demo seed data ----

func seedUsers() []*entity.User {
	u, _ := entity.NewUser("Demo User", "demo@atlas.local", "secret123")
	u.ID = "demo-user-1"
	u.Preferences.PreferredCategories = []string{"attraction", "restaurant"}
	u.Preferences.PreferredTags = []string{"culture", "food"}
	u.RecordLogin()
	return []*entity.User{u}
}

func seedPOIs() []*entity.POI {
	now := time.Now()
	mk := func(id, name, city string, category entity.Category, lat, lng, rating, popularity, ticket, avgPrice float64, tags ...string) *entity.POI {
		poi, _ := entity.NewPOI(name, category, lat, lng, city)
		poi.ID = id
		poi.Address = city + " center"
		poi.Rating = rating
		poi.Popularity = popularity
		poi.TicketPrice = ticket
		poi.AvgPrice = avgPrice
		poi.Duration = 90
		poi.CreatedAt = now
		poi.UpdatedAt = now
		for _, tag := range tags {
			poi.AddTag(entity.Tag{ID: tag, Name: tag, Type: "theme"})
		}
		return poi
	}
	return []*entity.POI{
		mk("tokyo-sensoji", "Senso-ji", "Tokyo", entity.CategoryTemple, 35.7148, 139.7967, 4.8, 95, 0, 0, "culture", "temple"),
		mk("tokyo-skytree", "Tokyo Skytree", "Tokyo", entity.CategoryAttraction, 35.7101, 139.8107, 4.7, 92, 180, 0, "view", "landmark"),
		mk("tokyo-tsukiji", "Tsukiji Market", "Tokyo", entity.CategoryRestaurant, 35.6655, 139.7708, 4.6, 90, 0, 120, "food", "market"),
		mk("tokyo-ueno", "Ueno Park", "Tokyo", entity.CategoryPark, 35.7156, 139.7730, 4.5, 88, 0, 0, "nature", "culture"),
		mk("shanghai-bund", "The Bund", "Shanghai", entity.CategoryAttraction, 31.2400, 121.4900, 4.7, 93, 0, 0, "view", "landmark"),
	}
}
