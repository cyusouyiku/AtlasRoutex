CREATE INDEX IF NOT EXISTS idx2_pois_name_trgm
    ON pois USING gin (name gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx2_pois_name_en_trgm
    ON pois USING gin (name_en gin_trgm_ops)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- POI: composite filters (city + category), tag-heavy JSONB
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx2_pois_country_city_category_live
    ON pois (country_code, city, category)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx2_pois_city_category_live
    ON pois (city, category)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx2_pois_city_rating_live
    ON pois (city, rating DESC NULLS LAST, popularity DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx2_pois_tags_gin
    ON pois USING gin (tags)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Itineraries: list by user + recency; budget range (FindByBudgetRange)
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx2_itineraries_user_created_live
    ON itineraries (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx2_itineraries_budget_total_cost
    ON itineraries (((budget->>'total_cost')::double precision))
    WHERE deleted_at IS NULL AND budget IS NOT NULL AND budget ? 'total_cost';

-- Optional: planner filters on total budget ceiling
CREATE INDEX IF NOT EXISTS idx2_itineraries_budget_total_budget
    ON itineraries (((budget->>'total_budget')::double precision))
    WHERE deleted_at IS NULL AND budget IS NOT NULL AND budget ? 'total_budget';

-- ---------------------------------------------------------------------------
-- Users: phone lookups (ExistsByPhone)
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx2_users_phone_active
    ON users (phone)
    WHERE status <> 'deleted' AND phone IS NOT NULL AND btrim(phone) <> '';

-- ---------------------------------------------------------------------------
-- user_saved_pois / favorites / feedback / synonyms
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx2_user_saved_pois_user_created
    ON user_saved_pois (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx2_user_itin_fav_user_created
    ON user_itinerary_favorites (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx2_poi_synonyms_alias_trgm
    ON poi_search_synonyms USING gin (alias gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx2_feedback_occurred
    ON feedback_events (occurred_at DESC);
