CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    age             INTEGER NOT NULL DEFAULT 0,
    password_hash   VARCHAR(255) NOT NULL, -- bcrypt
    role            VARCHAR(32) NOT NULL DEFAULT 'user',
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    preferences     JSONB NOT NULL DEFAULT '{}',
    itinerary_count INTEGER NOT NULL DEFAULT 0,
    total_distance  DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active
    ON users (email)
    WHERE status <> 'deleted';

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role) WHERE status <> 'deleted';
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users (last_login_at DESC)
    WHERE status = 'active' AND last_login_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- pois (soft delete via deleted_at)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pois (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(200) NOT NULL,
    name_en         VARCHAR(200) NOT NULL DEFAULT '',
    name_local      VARCHAR(200) NOT NULL DEFAULT '',
    category        VARCHAR(50) NOT NULL,
    sub_category    VARCHAR(100) NOT NULL DEFAULT '',
    country_code    CHAR(2) NOT NULL DEFAULT 'CN',
    lat             DOUBLE PRECISION NOT NULL,
    lng             DOUBLE PRECISION NOT NULL,
    address         TEXT NOT NULL DEFAULT '',
    city            VARCHAR(100) NOT NULL,
    district        VARCHAR(100) NOT NULL DEFAULT '',
    geohash         VARCHAR(20),
    opening_hours   JSONB NOT NULL DEFAULT '{}',
    avg_stay_time   INTEGER NOT NULL DEFAULT 60,
    best_time       JSONB NOT NULL DEFAULT '[]',
    duration        INTEGER NOT NULL DEFAULT 60,
    price_level     VARCHAR(20),
    ticket_price    DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_price       DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating          DOUBLE PRECISION NOT NULL DEFAULT 0
        CHECK (rating >= 0 AND rating <= 5),
    rating_count    INTEGER NOT NULL DEFAULT 0,
    popularity      DOUBLE PRECISION NOT NULL DEFAULT 0
        CHECK (popularity >= 0 AND popularity <= 100),
    rank            INTEGER NOT NULL DEFAULT 0,
    tags            JSONB NOT NULL DEFAULT '[]',
    features        JSONB NOT NULL DEFAULT '[]',
    similar_pois    JSONB NOT NULL DEFAULT '[]',
    images          JSONB NOT NULL DEFAULT '[]',
    thumbnail       VARCHAR(500) NOT NULL DEFAULT '',
    videos          JSONB NOT NULL DEFAULT '[]',
    booking_url     VARCHAR(500) NOT NULL DEFAULT '',
    is_bookable     BOOLEAN NOT NULL DEFAULT FALSE,
    inventory       INTEGER NOT NULL DEFAULT 0,
    description     TEXT NOT NULL DEFAULT '',
    tips            JSONB NOT NULL DEFAULT '[]',
    warnings        JSONB NOT NULL DEFAULT '[]',
    source          VARCHAR(100) NOT NULL DEFAULT '',
    confidence      DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pois_city_live ON pois (city) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_country_city_live ON pois (country_code, city) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_category_live ON pois (category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_rating_live ON pois (rating DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_popularity_live ON pois (popularity DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_lat_lng_live ON pois (lat, lng) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pois_geohash_live ON pois (geohash) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- itineraries (aggregate root + JSON blobs; soft delete)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS itineraries (
    id             VARCHAR(36) PRIMARY KEY,
    user_id        VARCHAR(36) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name           VARCHAR(200) NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    status         VARCHAR(32) NOT NULL DEFAULT 'draft',
    start_date     DATE NOT NULL,
    end_date       DATE NOT NULL,
    day_count      INTEGER NOT NULL,
    budget         JSONB,
    statistics     JSONB,
    "constraints"  JSONB,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_itineraries_user_live ON itineraries (user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_itineraries_status_live ON itineraries (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_itineraries_start_date ON itineraries (start_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_itineraries_favorite_live ON itineraries (favorite_count DESC)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- itinerary_days (per-day stats columns — matches saveDays / fetchDays)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS itinerary_days (
    id                VARCHAR(36) PRIMARY KEY,
    itinerary_id      VARCHAR(36) NOT NULL REFERENCES itineraries (id) ON DELETE CASCADE,
    day_number        INTEGER NOT NULL CHECK (day_number >= 1),
    date              DATE NOT NULL,
    notes             TEXT NOT NULL DEFAULT '',
    walking_distance  DOUBLE PRECISION NOT NULL DEFAULT 0,
    walking_time      INTEGER NOT NULL DEFAULT 0,
    place_count       INTEGER NOT NULL DEFAULT 0,
    daily_cost        DOUBLE PRECISION NOT NULL DEFAULT 0,
    attraction_time   INTEGER NOT NULL DEFAULT 0,
    elevation_gain    DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_rest           BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (itinerary_id, day_number)
);

CREATE INDEX IF NOT EXISTS idx_itinerary_days_itinerary ON itinerary_days (itinerary_id);

-- ---------------------------------------------------------------------------
-- day_attractions (itinerary_id + day_number; no day_id — see repo INSERT)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS day_attractions (
    id              VARCHAR(36) PRIMARY KEY,
    itinerary_id    VARCHAR(36) NOT NULL REFERENCES itineraries (id) ON DELETE CASCADE,
    day_number      INTEGER NOT NULL,
    poi_id          VARCHAR(36) NOT NULL REFERENCES pois (id),
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    stay_duration   INTEGER,
    "order"         INTEGER NOT NULL,
    cost            DOUBLE PRECISION NOT NULL DEFAULT 0,
    transportation  JSONB,
    notes           TEXT NOT NULL DEFAULT '',
    CHECK (day_number >= 1)
);

CREATE INDEX IF NOT EXISTS idx_day_attractions_itin_day ON day_attractions (itinerary_id, day_number);
CREATE INDEX IF NOT EXISTS idx_day_attractions_order ON day_attractions (itinerary_id, day_number, "order");
CREATE INDEX IF NOT EXISTS idx_day_attractions_poi ON day_attractions (poi_id);
CREATE INDEX IF NOT EXISTS idx_day_attractions_poi_itin ON day_attractions (poi_id, itinerary_id);

-- ---------------------------------------------------------------------------
-- day_meals
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS day_meals (
    id              VARCHAR(36) PRIMARY KEY,
    itinerary_id    VARCHAR(36) NOT NULL REFERENCES itineraries (id) ON DELETE CASCADE,
    day_number      INTEGER NOT NULL,
    meal_type       VARCHAR(20) NOT NULL,
    restaurant_id   VARCHAR(36) REFERENCES pois (id),
    meal_time       TIMESTAMPTZ,
    cost            DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes           TEXT NOT NULL DEFAULT '',
    CHECK (day_number >= 1)
);

CREATE INDEX IF NOT EXISTS idx_day_meals_itin_day ON day_meals (itinerary_id, day_number);
CREATE INDEX IF NOT EXISTS idx_day_meals_restaurant ON day_meals (restaurant_id);

-- ---------------------------------------------------------------------------
-- day_hotels (at most one row per day: enforced in app; optional DB guard below)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS day_hotels (
    id              VARCHAR(36) PRIMARY KEY,
    itinerary_id    VARCHAR(36) NOT NULL REFERENCES itineraries (id) ON DELETE CASCADE,
    day_number      INTEGER NOT NULL,
    hotel_name      VARCHAR(200) NOT NULL,
    address         TEXT NOT NULL DEFAULT '',
    check_in_time   TIMESTAMPTZ,
    check_out_time  TIMESTAMPTZ,
    cost            DOUBLE PRECISION NOT NULL DEFAULT 0,
    room_type       VARCHAR(100) NOT NULL DEFAULT '',
    notes           TEXT NOT NULL DEFAULT '',
    CHECK (day_number >= 1),
    UNIQUE (itinerary_id, day_number)
);

CREATE INDEX IF NOT EXISTS idx_day_hotels_itin_day ON day_hotels (itinerary_id, day_number);

-- ---------------------------------------------------------------------------
-- schema_migrations — CLI / runner bookkeeping (version strings, e.g. 001, 002)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     VARCHAR(64) PRIMARY KEY,
    description TEXT NOT NULL DEFAULT '',
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- user_saved_pois — 用户收藏 POI（推荐 / 画像扩展）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_saved_pois (
    user_id    VARCHAR(36) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    poi_id     VARCHAR(36) NOT NULL REFERENCES pois (id) ON DELETE CASCADE,
    note       VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, poi_id)
);

CREATE INDEX IF NOT EXISTS idx_user_saved_pois_poi ON user_saved_pois (poi_id);

-- ---------------------------------------------------------------------------
-- user_itinerary_favorites — 用户收藏行程（favorite_count 由应用同步）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_itinerary_favorites (
    user_id      VARCHAR(36) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    itinerary_id VARCHAR(36) NOT NULL REFERENCES itineraries (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, itinerary_id)
);

CREATE INDEX IF NOT EXISTS idx_user_itin_fav_itin ON user_itinerary_favorites (itinerary_id);

-- ---------------------------------------------------------------------------
-- feedback_events — 对齐 application/feedback 持久化
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS feedback_events (
    id               VARCHAR(36) PRIMARY KEY,
    idempotency_key  VARCHAR(128) NOT NULL,
    user_id          VARCHAR(36) REFERENCES users (id) ON DELETE SET NULL,
    itinerary_id     VARCHAR(36),
    poi_id           VARCHAR(36),
    kind             VARCHAR(40) NOT NULL,
    rating           DOUBLE PRECISION,
    comment          TEXT NOT NULL DEFAULT '',
    occurred_at      TIMESTAMPTZ NOT NULL,
    repeat_count     INTEGER NOT NULL DEFAULT 1,
    extra            JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_feedback_user ON feedback_events (user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_poi ON feedback_events (poi_id);
CREATE INDEX IF NOT EXISTS idx_feedback_itin ON feedback_events (itinerary_id);
CREATE INDEX IF NOT EXISTS idx_feedback_kind_time ON feedback_events (kind, occurred_at DESC);

-- ---------------------------------------------------------------------------
-- poi_search_synonyms — 多语言别名 / 检索扩展
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS poi_search_synonyms (
    id         VARCHAR(36) PRIMARY KEY,
    poi_id     VARCHAR(36) NOT NULL REFERENCES pois (id) ON DELETE CASCADE,
    lang       VARCHAR(10) NOT NULL DEFAULT '',
    alias      VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (poi_id, lang, alias)
);

CREATE INDEX IF NOT EXISTS idx_poi_synonyms_alias ON poi_search_synonyms (alias);
CREATE INDEX IF NOT EXISTS idx_poi_synonyms_poi ON poi_search_synonyms (poi_id);

