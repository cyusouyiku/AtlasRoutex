BEGIN;

INSERT INTO schema_migrations (version, description) VALUES
    ('001', '001_init_schema.sql'),
    ('002', '002_add_indexes.sql'),
    ('003', '003_seed_data.sql')
ON CONFLICT (version) DO NOTHING;

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
INSERT INTO users (
    id, name, email, phone, age, password_hash, role, status,
    preferences, itinerary_count, total_distance, created_at, updated_at
) VALUES (
    '00000000-0000-4000-a000-000000000001',
    'Demo User',
    'demo@example.com',
    '',
    0,
    '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'user',
    'active',
    '{"pace":"normal","currency":"CNY","languages":["zh","en"],"preferred_categories":["attraction"],"preferred_tags":[],"avoid_tags":[],"default_budget":500,"dietary_restrictions":[]}'::jsonb,
    2,
    0,
    NOW(),
    NOW()
),
(
    '00000000-0000-4000-a000-000000000002',
    '上海测试用户',
    'shanghai@example.com',
    '13800138000',
    28,
    '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'user',
    'active',
    '{"pace":"relaxed","currency":"CNY","languages":["zh"],"preferred_categories":["museum","shopping"],"preferred_tags":["夜景"],"avoid_tags":[],"default_budget":800,"dietary_restrictions":[]}'::jsonb,
    1,
    0,
    NOW(),
    NOW()
),
(
    '00000000-0000-4000-a000-000000000003',
    '日本线体验用户',
    'japan@example.com',
    '',
    35,
    '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'user',
    'active',
    '{"pace":"intensive","currency":"JPY","languages":["zh","ja"],"preferred_categories":["temple","nature"],"preferred_tags":[],"avoid_tags":[],"default_budget":120000,"dietary_restrictions":[]}'::jsonb,
    2,
    0,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- pois (explicit country_code: CN / JP)
-- ---------------------------------------------------------------------------
INSERT INTO pois (
    id, name, name_en, name_local, category, sub_category, country_code,
    lat, lng, address, city, district, geohash,
    opening_hours, avg_stay_time, best_time, duration,
    price_level, ticket_price, avg_price,
    rating, rating_count, popularity, rank,
    tags, features, similar_pois,
    images, thumbnail, videos,
    booking_url, is_bookable, inventory,
    description, tips, warnings,
    source, confidence, created_at, updated_at, deleted_at
) VALUES
-- 北京
(
    '00000000-0000-4000-a000-000000000011',
    '故宫博物院',
    'Forbidden City',
    '故宫博物院',
    'attraction',
    'palace',
    'CN',
    39.916345, 116.397155,
    '北京市东城区景山前街4号',
    '北京',
    '东城区',
    'wx4g0',
    '{}'::jsonb, 180, '["春季","秋季"]'::jsonb, 180,
    'medium', 60, 0,
    4.8, 12000, 95.0, 1,
    '[{"id":"tag-history","name":"历史","type":"feature"},{"id":"tag-unesco","name":"世界遗产","type":"feature"}]'::jsonb,
    '["室内","步行多"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '明清两朝皇宫，世界文化遗产。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-a000-000000000012',
    '天坛公园',
    'Temple of Heaven',
    '天坛公园',
    'attraction',
    'park',
    'CN',
    39.881913, 116.406825,
    '北京市东城区天坛东里甲1号',
    '北京',
    '东城区',
    'wx4g0',
    '{}'::jsonb, 120, '[]'::jsonb, 120,
    'cheap', 15, 0,
    4.7, 8000, 88.0, 2,
    '[{"id":"tag-park","name":"公园","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '明清皇帝祭天场所。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-a000-000000000013',
    '全聚德烤鸭店(前门店)',
    'Quanjude Roast Duck',
    '全聚德烤鸭店',
    'restaurant',
    'chinese',
    'CN',
    39.901085, 116.397470,
    '北京市东城区前门大街30号',
    '北京',
    '东城区',
    'wx4g0',
    '{}'::jsonb, 90, '[]'::jsonb, 90,
    'expensive', 0, 200,
    4.3, 3500, 72.0, 10,
    '[{"id":"tag-pekingduck","name":"烤鸭","type":"cuisine"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '老字号烤鸭。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-a000-000000000014',
    '北京饭店诺金',
    'Hotel NuO Beijing',
    '北京饭店诺金',
    'hotel',
    'luxury',
    'CN',
    39.909289, 116.410518,
    '北京市东城区东长安街33号',
    '北京',
    '东城区',
    'wx4g0',
    '{}'::jsonb, 0, '[]'::jsonb, 0,
    'luxury', 0, 1200,
    4.5, 900, 60.0, 3,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 40,
    '王府井附近高端酒店示例。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-a000-000000000015',
    '颐和园',
    'Summer Palace',
    '颐和园',
    'attraction',
    'palace',
    'CN',
    39.999982, 116.275460,
    '北京市海淀区新建宫门路19号',
    '北京',
    '海淀区',
    'wx4d2u',
    '{}'::jsonb, 150, '["秋季"]'::jsonb, 150,
    'medium', 30, 0,
    4.75, 9000, 90.0, 4,
    '[{"id":"tag-royal-garden","name":"皇家园林","type":"feature"}]'::jsonb,
    '["湖景","颐和园路绕行"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '中国古典园林代表作之一。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-a000-000000000016',
    '慕田峪长城',
    'Mutianyu Great Wall',
    '慕田峪长城',
    'attraction',
    'nature',
    'CN',
    40.431908, 116.570374,
    '北京市怀柔区渤海镇慕田峪村',
    '北京',
    '怀柔区',
    'wxey9',
    '{}'::jsonb, 240, '["春","秋"]'::jsonb, 240,
    'medium', 45, 0,
    4.65, 5000, 85.0, 5,
    '[{"id":"tag-greatwall","name":"长城","type":"feature"}]'::jsonb,
    '["登山","缆车可选"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 2000,
    '相对缓和、植被好的长城段落示例数据。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
-- 上海
(
    '00000000-0000-4000-b000-000000000021',
    '外滩',
    'The Bund',
    '外滩',
    'attraction',
    'waterfront',
    'CN',
    31.239668, 121.488225,
    '上海市黄浦区中山东一路',
    '上海',
    '黄浦区',
    'wtw3ed',
    '{}'::jsonb, 90, '["夜景"]'::jsonb, 120,
    'free', 0, 0,
    4.85, 15000, 98.0, 1,
    '[{"id":"tag-bund","name":"外滩","type":"feature"},{"id":"tag-night","name":"夜景","type":"atmosphere"}]'::jsonb,
    '["步行","拍照"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '浦江经典天际线。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000022',
    '豫园',
    'Yu Garden',
    '豫园',
    'attraction',
    'garden',
    'CN',
    31.227131, 121.492065,
    '上海市黄浦区福佑路168号',
    '上海',
    '黄浦区',
    'wtw3er',
    '{}'::jsonb, 90, '[]'::jsonb, 90,
    'cheap', 40, 0,
    4.55, 6000, 80.0, 2,
    '[{"id":"tag-garden","name":"园林","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '江南古典园林，城隍庙商圈旁。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000023',
    '上海博物馆',
    'Shanghai Museum',
    '上海博物馆',
    'museum',
    'history',
    'CN',
    31.228667, 121.475297,
    '上海市黄浦区人民大道201号',
    '上海',
    '黄浦区',
    'wtw3su',
    '{}'::jsonb, 120, '[]'::jsonb, 150,
    'free', 0, 0,
    4.78, 7000, 87.0, 3,
    '[{"id":"tag-museum","name":"博物馆","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '中国古代艺术收藏重镇（示例）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000024',
    '南翔馒头店（豫园店）',
    'Nanxiang Steamed Bun Restaurant',
    '南翔馒头店',
    'restaurant',
    'chinese',
    'CN',
    31.226300, 121.491800,
    '上海市黄浦区豫园路85号',
    '上海',
    '黄浦区',
    'wtw3er',
    '{}'::jsonb, 60, '[]'::jsonb, 60,
    'medium', 0, 85,
    4.2, 4200, 70.0, 8,
    '[{"id":"tag-xiaolong","name":"小笼包","type":"cuisine"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '豫园商圈小笼包名店（示例数据）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000025',
    '上海和平饭店',
    'Fairmont Peace Hotel',
    '和平饭店',
    'hotel',
    'heritage',
    'CN',
    31.240833, 121.488611,
    '上海市黄浦区南京东路20号',
    '上海',
    '黄浦区',
    'wtw3ed',
    '{}'::jsonb, 0, '[]'::jsonb, 0,
    'luxury', 0, 1600,
    4.72, 2100, 75.0, 4,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 60,
    '外滩历史地标酒店。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000026',
    '田子坊',
    'Tianzifang",
    '田子坊',
    'shopping',
    'lanes',
    'CN',
    31.210658, 121.470339,
    '上海市黄浦区泰康路210弄',
    '上海',
    '黄浦区',
    'wtw3st',
    '{}'::jsonb, 120, '[]'::jsonb, 90,
    'cheap', 0, 120,
    4.1, 8000, 65.0, 6,
    '[{"id":"tag-creative","name":"文创","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '石库门里弄改造的购物漫步区。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000027',
    '上海迪士尼乐园',
    'Shanghai Disneyland',
    '上海迪士尼乐园',
    'entertainment',
    'theme_park',
    'CN',
    31.143870, 121.655942,
    '上海市浦东新区川沙新镇黄赵路310号',
    '上海',
    '浦东新区',
    'wtteyz',
    '{}'::jsonb, 480, '["节假日"]'::jsonb, 480,
    'luxury', 399, 300,
    4.6, 20000, 92.0, 5,
    '[{"id":"tag-family","name":"亲子","type":"audience"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 10000,
    '大型主题乐园（示例门票价）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
-- 东京
(
    '00000000-0000-4000-b000-000000000031',
    '浅草寺',
    'Senso-ji Temple',
    '浅草寺',
    'temple',
    'buddhist',
    'JP',
    35.714765, 139.796655,
    '東京都台東区浅草2-3-1',
    '东京',
    '台东区',
    'xn774c',
    '{}'::jsonb, 90, '["春季"]'::jsonb, 90,
    'free', 0, 0,
    4.65, 18000, 93.0, 1,
    '[{"id":"tag-asakusa","name":"浅草","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '东京最古老的寺院之一，雷门・仲见世通。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000032',
    '東京スカイツリー',
    'Tokyo Skytree',
    '東京スカイツリー',
    'attraction',
    'tower',
    'JP',
    35.710064, 139.810700,
    '東京都墨田区押上1-1-2',
    '东京',
    '墨田区',
    'xn774c',
    '{}'::jsonb, 120, '["夜景"]'::jsonb, 120,
    'expensive', 2100, 0,
    4.5, 12000, 91.0, 2,
    '[{"id":"tag-skytree","name":"晴空塔","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 5000,
    '押上地标展望台（示例日元门票）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000033',
    '明治神宮',
    'Meiji Jingu',
    '明治神宮',
    'temple',
    'shinto',
    'JP',
    35.676398, 139.699326,
    '東京都渋谷区代々木神園町1-1',
    '东京',
    '涩谷区',
    'xn76hk',
    '{}'::jsonb, 75, '[]'::jsonb, 75,
    'free', 0, 0,
    4.7, 14000, 88.0, 3,
    '[{"id":"tag-meiji","name":"明治神宫","type":"feature"}]'::jsonb,
    '["森林步道"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '都市内大片林地的神宫。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000034',
    'すし大辻（築地场外意象店）',
    'Tsukiji Outer Market Sushi (sample)',
    '築地すし',
    'restaurant',
    'japanese',
    'JP',
    35.665361, 139.770509,
    '東京都中央区築地（场外市场一带）',
    '东京',
    '中央区',
    'xn774b',
    '{}'::jsonb, 75, '["早间"]'::jsonb, 75,
    'expensive', 0, 8000,
    4.45, 3000, 82.0, 7,
    '[{"id":"tag-sushi","name":"寿司","type":"cuisine"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '筑地场外海鲜与寿司（示例点位）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000035',
    '一蘭 新宿中央東口店',
    'Ichiran Shinjuku',
    '一蘭',
    'restaurant',
    'ramen',
    'JP',
    35.693840, 139.701427,
    '東京都新宿区神室町（示意地址）',
    '东京',
    '新宿区',
    'xn76j8',
    '{}'::jsonb, 45, '[]'::jsonb, 45,
    'cheap', 0, 1200,
    4.35, 9000, 78.0, 9,
    '[{"id":"tag-ramen","name":"拉面","type":"cuisine"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '博多风豚骨拉面连锁示例。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000036',
    '京王プラザホテル',
    'Keio Plaza Hotel Tokyo',
    '京王广场酒店',
    'hotel',
    'business',
    'JP',
    35.693004, 139.693762,
    '東京都新宿区西新宿2-2-1',
    '东京',
    '新宿区',
    'xn76j8',
    '{}'::jsonb, 0, '[]'::jsonb, 0,
    'expensive', 0, 22000,
    4.4, 5000, 70.0, 5,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 120,
    '新宿西站口大型酒店示例。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000037',
    '皇居東御苑',
    'East Gardens of the Imperial Palace',
    '皇居東御苑',
    'park',
    'garden',
    'JP',
    35.685175, 139.755393,
    '東京都千代田区千代田1-1',
    '东京',
    '千代田区',
    'xn76me',
    '{}'::jsonb, 90, '["樱花季"]'::jsonb, 90,
    'free', 0, 0,
    4.55, 4000, 76.0, 4,
    '[{"id":"tag-palace","name":"皇居","type":"feature"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '皇居开放区域散步绿地。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
-- 京都
(
    '00000000-0000-4000-b000-000000000041',
    '清水寺',
    'Kiyomizu-dera',
    '清水寺',
    'temple',
    'buddhist',
    'JP',
    34.994856, 135.785046,
    '京都府京都市東山区清水1-294',
    '京都',
    '东山区',
    'xn7ms8',
    '{}'::jsonb, 120, '["红叶"]'::jsonb, 120,
    'medium', 400, 0,
    4.8, 11000, 94.0, 1,
    '[{"id":"tag-kiyomizu","name":"清水寺","type":"feature"}]'::jsonb,
    '["山路","清水舞台"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '世界遗产，眺望京都市景名所。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000042',
    '伏見稲荷大社',
    'Fushimi Inari Taisha',
    '伏見稲荷大社',
    'temple',
    'shinto',
    'JP',
    34.967102, 135.772701,
    '京都府京都市伏見区深草藪之内町68',
    '京都',
    '伏见区',
    'xn7ms2',
    '{}'::jsonb, 150, '["清晨"]'::jsonb, 150,
    'free', 0, 0,
    4.85, 16000, 96.0, 2,
    '[{"id":"tag-torii","name":"千本鸟居","type":"feature"}]'::jsonb,
    '["登山步道"]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '千本鸟居参道盛名。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000043',
    '祇園花見小路',
    'Gion Hanamikoji',
    '祇園',
    'entertainment',
    'district',
    'JP',
    35.003962, 135.775897,
    '京都市东山区祇园町北侧',
    '京都',
    '东山区',
    'xn7ms8',
    '{}'::jsonb, 90, '["晚间"]'::jsonb, 90,
    'free', 0, 5000,
    4.3, 7000, 80.0, 6,
    '[{"id":"tag-geisha","name":"花见小路","type":"atmosphere"}]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', false, 0,
    '传统街区散步与茶馆意象。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
),
(
    '00000000-0000-4000-b000-000000000044',
    '京都柊家旅馆（意象）',
    'Hiiragiya Ryokan (sample)',
    '柊家',
    'hotel',
    'ryokan',
    'JP',
    35.005983, 135.763089,
    '京都市中京区麸屋町姉小路上ル中麸屋町477',
    '京都',
    '中京区',
    'xn7ms6',
    '{}'::jsonb, 0, '[]'::jsonb, 0,
    'luxury', 0, 55000,
    4.9, 800, 55.0, 3,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '', '[]'::jsonb,
    '', true, 15,
    '京町家旅馆示例（价格单位为日元意象）。',
    '[]'::jsonb,
    '[]'::jsonb,
    'seed', 1.0,
    NOW(), NOW(), NULL
)
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- poi_search_synonyms
-- ---------------------------------------------------------------------------
INSERT INTO poi_search_synonyms (id, poi_id, lang, alias) VALUES
('00000000-0000-4000-d000-000000000001', '00000000-0000-4000-b000-000000000031', 'en', 'Asakusa Temple'),
('00000000-0000-4000-d000-000000000002', '00000000-0000-4000-b000-000000000032', 'en', 'Tokyo Skytree'),
('00000000-0000-4000-d000-000000000003', '00000000-0000-4000-b000-000000000041', 'ja', 'きよみずでら'),
('00000000-0000-4000-d000-000000000004', '00000000-0000-4000-b000-000000000021', 'en', 'Waitan')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- itineraries
-- ---------------------------------------------------------------------------
INSERT INTO itineraries (
    id, user_id, name, description, status,
    start_date, end_date, day_count,
    budget, statistics, "constraints",
    favorite_count, created_at, updated_at, published_at, deleted_at
) VALUES (
    '00000000-0000-4000-c000-000000000021',
    '00000000-0000-4000-a000-000000000001',
    '北京两日文化线（示例）',
    '故宫、天坛、颐和园与城市美食（示例）。',
    'planned',
    DATE '2026-04-10',
    DATE '2026-04-11',
    2,
    '{"total_budget":8000,"total_cost":400,"currency":"CNY","categories":[],"remaining":7600}'::jsonb,
    '{"total_distance":20,"total_walking_time":120,"total_attraction_time":480,"average_score":4.6,"place_count":4,"rest_day_count":0}'::jsonb,
    '[]'::jsonb,
    1,
    NOW(),
    NOW(),
    NULL,
    NULL
),
(
    '00000000-0000-4000-c000-000000000022',
    '00000000-0000-4000-a000-000000000002',
    '上海周末：外滩 · 豫园 · 文博',
    '浦江夜景与老城厢（示例）。',
    'planned',
    DATE '2026-05-01',
    DATE '2026-05-02',
    2,
    '{"total_budget":6000,"total_cost":520,"currency":"CNY","categories":[],"remaining":5480}'::jsonb,
    '{"total_distance":8,"total_walking_time":60,"total_attraction_time":300,"average_score":4.5,"place_count":3,"rest_day_count":0}'::jsonb,
    '[]'::jsonb,
    1,
    NOW(),
    NOW(),
    NULL,
    NULL
),
(
    '00000000-0000-4000-c000-000000000023',
    '00000000-0000-4000-a000-000000000003',
    '东京三日：下町、天空线、皇城绿肺',
    '浅草・晴空塔・明治神宫・筑地/新宿（示例）。',
    'confirmed',
    DATE '2026-06-01',
    DATE '2026-06-03',
    3,
    '{"total_budget":150000,"total_cost":12000,"currency":"JPY","categories":[],"remaining":138000}'::jsonb,
    '{"total_distance":15,"total_walking_time":150,"total_attraction_time":600,"average_score":4.55,"place_count":5,"rest_day_count":0}'::jsonb,
    '[]'::jsonb,
    1,
    NOW(),
    NOW(),
    NOW(),
    NULL
),
(
    '00000000-0000-4000-c000-000000000024',
    '00000000-0000-4000-a000-000000000003',
    '京都两日：寺庙与街区',
    '清水寺、伏见稻荷、祇园散步（示例）。',
    'draft',
    DATE '2026-06-10',
    DATE '2026-06-11',
    2,
    '{"total_budget":80000,"total_cost":0,"currency":"JPY","categories":[],"remaining":80000}'::jsonb,
    '{"total_distance":10,"total_walking_time":90,"total_attraction_time":360,"average_score":4.7,"place_count":3,"rest_day_count":0}'::jsonb,
    '[]'::jsonb,
    0,
    NOW(),
    NOW(),
    NULL,
    NULL
)
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 北京 itinerary days / attractions / meal / hotel ( ids renumbered for 4.x itinerary )
-- ---------------------------------------------------------------------------
INSERT INTO itinerary_days (
    id, itinerary_id, day_number, date, notes,
    walking_distance, walking_time, place_count, daily_cost, attraction_time,
    elevation_gain, is_rest
) VALUES
('00000000-0000-4000-e000-000000000031', '00000000-0000-4000-c000-000000000021', 1, DATE '2026-04-10',
 '紫禁城与天坛', 6.0, 50, 2, 200.0, 300, 0, false),
('00000000-0000-4000-e000-000000000032', '00000000-0000-4000-c000-000000000021', 2, DATE '2026-04-11',
 '皇家园林与长城（示例）', 8.0, 40, 2, 280.0, 390, 120, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_attractions (
    id, itinerary_id, day_number, poi_id,
    start_time, end_time, stay_duration, "order", cost,
    transportation, notes
) VALUES
('00000000-0000-4000-f000-000000000041', '00000000-0000-4000-c000-000000000021', 1,
 '00000000-0000-4000-a000-000000000011',
 '2026-04-10T09:00:00+08:00'::timestamptz, '2026-04-10T12:00:00+08:00'::timestamptz, 180, 1, 60.0, NULL, ''),
('00000000-0000-4000-f000-000000000042', '00000000-0000-4000-c000-000000000021', 1,
 '00000000-0000-4000-a000-000000000012',
 '2026-04-10T14:00:00+08:00'::timestamptz, '2026-04-10T16:30:00+08:00'::timestamptz, 150, 2, 15.0,
 '{"type":"subway","distance":12,"duration":35,"cost":5}'::jsonb, ''),
('00000000-0000-4000-f000-000000000043', '00000000-0000-4000-c000-000000000021', 2,
 '00000000-0000-4000-a000-000000000015',
 '2026-04-11T09:30:00+08:00'::timestamptz, '2026-04-11T12:30:00+08:00'::timestamptz, 180, 1, 30.0, NULL, ''),
('00000000-0000-4000-f000-000000000044', '00000000-0000-4000-c000-000000000021', 2,
 '00000000-0000-4000-a000-000000000016',
 '2026-04-11T14:00:00+08:00'::timestamptz, '2026-04-11T17:00:00+08:00'::timestamptz, 180, 2, 45.0,
 '{"type":"bus","distance":65,"duration":90,"cost":30}'::jsonb, '慕田峪当日往返示意')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_meals (id, itinerary_id, day_number, meal_type, restaurant_id, meal_time, cost, notes) VALUES
('00000000-0000-4000-f000-000000000051', '00000000-0000-4000-c000-000000000021', 1, 'lunch',
 '00000000-0000-4000-a000-000000000013', '2026-04-10T12:15:00+08:00'::timestamptz, 200.0, '烤鸭')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_hotels (id, itinerary_id, day_number, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes) VALUES
('00000000-0000-4000-f000-000000000061', '00000000-0000-4000-c000-000000000021', 1,
 '北京饭店诺金', '北京市东城区东长安街33号',
 '2026-04-10T21:00:00+08:00'::timestamptz, '2026-04-11T10:00:00+08:00'::timestamptz, 880.0, '豪华大床', '')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 上海 itinerary
-- ---------------------------------------------------------------------------
INSERT INTO itinerary_days (
    id, itinerary_id, day_number, date, notes,
    walking_distance, walking_time, place_count, daily_cost, attraction_time,
    elevation_gain, is_rest
) VALUES
('00000000-0000-4000-e000-000000000071', '00000000-0000-4000-c000-000000000022', 1, DATE '2026-05-01', '外滩夜景与万国建筑', 4.0, 35, 2, 350.0, 200, 0, false),
('00000000-0000-4000-e000-000000000072', '00000000-0000-4000-c000-000000000022', 2, DATE '2026-05-02', '豫园与博物馆', 5.5, 40, 2, 180.0, 240, 0, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_attractions (
    id, itinerary_id, day_number, poi_id,
    start_time, end_time, stay_duration, "order", cost, transportation, notes
) VALUES
('00000000-0000-4000-f000-000000000081', '00000000-0000-4000-c000-000000000022', 1,
 '00000000-0000-4000-b000-000000000026',
 '2026-05-01T15:00:00+08:00'::timestamptz, '2026-05-01T17:00:00+08:00'::timestamptz, 120, 1, 0,
 '{"type":"walk","distance":2.5,"duration":35,"cost":0}'::jsonb, '田子坊'),
('00000000-0000-4000-f000-000000000082', '00000000-0000-4000-c000-000000000022', 1,
 '00000000-0000-4000-b000-000000000021',
 '2026-05-01T18:00:00+08:00'::timestamptz, '2026-05-01T20:30:00+08:00'::timestamptz, 150, 2, 0, NULL, '外滩夜景'),
('00000000-0000-4000-f000-000000000083', '00000000-0000-4000-c000-000000000022', 2,
 '00000000-0000-4000-b000-000000000022',
 '2026-05-02T09:30:00+08:00'::timestamptz, '2026-05-02T11:30:00+08:00'::timestamptz, 120, 1, 40.0, NULL, ''),
('00000000-0000-4000-f000-000000000084', '00000000-0000-4000-c000-000000000022', 2,
 '00000000-0000-4000-b000-000000000023',
 '2026-05-02T13:00:00+08:00'::timestamptz, '2026-05-02T15:30:00+08:00'::timestamptz, 150, 2, 0,
 '{"type":"subway","distance":8,"duration":25,"cost":3}'::jsonb, '')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_meals (id, itinerary_id, day_number, meal_type, restaurant_id, meal_time, cost, notes) VALUES
('00000000-0000-4000-f000-000000000085', '00000000-0000-4000-c000-000000000022', 2, 'lunch',
 '00000000-0000-4000-b000-000000000024', '2026-05-02T12:00:00+08:00'::timestamptz, 120.0, '小笼包')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_hotels (id, itinerary_id, day_number, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes) VALUES
('00000000-0000-4000-f000-000000000086', '00000000-0000-4000-c000-000000000022', 1,
 '上海和平饭店', '上海市黄浦区南京东路20号',
 '2026-05-01T22:00:00+08:00'::timestamptz, '2026-05-02T12:00:00+08:00'::timestamptz, 1600.0, '费尔蒙套房', '')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 东京 itinerary (3 days)
-- ---------------------------------------------------------------------------
INSERT INTO itinerary_days (
    id, itinerary_id, day_number, date, notes,
    walking_distance, walking_time, place_count, daily_cost, attraction_time,
    elevation_gain, is_rest
) VALUES
('00000000-0000-4000-e000-000000000091', '00000000-0000-4000-c000-000000000023', 1, DATE '2026-06-01', '下町浅草', 5.0, 45, 1, 3500.0, 90, 0, false),
('00000000-0000-4000-e000-000000000092', '00000000-0000-4000-c000-000000000023', 2, DATE '2026-06-02', '晴空塔与皇居', 7.0, 50, 2, 4500.0, 210, 0, false),
('00000000-0000-4000-e000-000000000093', '00000000-0000-4000-c000-000000000023', 3, DATE '2026-06-03', '神宫外苑与新宿', 6.0, 40, 2, 5000.0, 150, 0, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_attractions (
    id, itinerary_id, day_number, poi_id,
    start_time, end_time, stay_duration, "order", cost, transportation, notes
) VALUES
('00000000-0000-4000-f000-000000000101', '00000000-0000-4000-c000-000000000023', 1,
 '00000000-0000-4000-b000-000000000031',
 '2026-06-01T08:30:00+09:00'::timestamptz, '2026-06-01T11:30:00+09:00'::timestamptz, 180, 1, 0, NULL, ''),
('00000000-0000-4000-f000-000000000102', '00000000-0000-4000-c000-000000000023', 2,
 '00000000-0000-4000-b000-000000000032',
 '2026-06-02T10:00:00+09:00'::timestamptz, '2026-06-02T12:00:00+09:00'::timestamptz, 120, 1, 2100.0, NULL, '展望台'),
('00000000-0000-4000-f000-000000000103', '00000000-0000-4000-c000-000000000023', 2,
 '00000000-0000-4000-b000-000000000037',
 '2026-06-02T14:00:00+09:00'::timestamptz, '2026-06-02T16:00:00+09:00'::timestamptz, 120, 2, 0,
 '{"type":"walk","distance":4,"duration":45,"cost":0}'::jsonb, ''),
('00000000-0000-4000-f000-000000000104', '00000000-0000-4000-c000-000000000023', 3,
 '00000000-0000-4000-b000-000000000033',
 '2026-06-03T09:00:00+09:00'::timestamptz, '2026-06-03T10:30:00+09:00'::timestamptz, 90, 1, 0, NULL, ''),
('00000000-0000-4000-f000-000000000105', '00000000-0000-4000-c000-000000000023', 3,
 '00000000-0000-4000-b000-000000000034',
 '2026-06-03T12:00:00+09:00'::timestamptz, '2026-06-03T13:30:00+09:00'::timestamptz, 90, 2, 8000.0, NULL, '寿司午餐示意')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_meals (id, itinerary_id, day_number, meal_type, restaurant_id, meal_time, cost, notes) VALUES
('00000000-0000-4000-f000-000000000106', '00000000-0000-4000-c000-000000000023', 3, 'dinner',
 '00000000-0000-4000-b000-000000000035', '2026-06-03T19:00:00+09:00'::timestamptz, 2500.0, '拉面')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_hotels (id, itinerary_id, day_number, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes) VALUES
('00000000-0000-4000-f000-000000000107', '00000000-0000-4000-c000-000000000023', 1,
 '京王プラザホテル', '東京都新宿区西新宿2-2-1',
 '2026-06-01T15:00:00+09:00'::timestamptz, '2026-06-04T11:00:00+09:00'::timestamptz, 65000.0, '标准房', '连住示意')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 京都 itinerary
-- ---------------------------------------------------------------------------
INSERT INTO itinerary_days (
    id, itinerary_id, day_number, date, notes,
    walking_distance, walking_time, place_count, daily_cost, attraction_time,
    elevation_gain, is_rest
) VALUES
('00000000-0000-4000-e000-000000000111', '00000000-0000-4000-c000-000000000024', 1, DATE '2026-06-10', '东山清水寺', 4.0, 55, 1, 8000.0, 120, 80, false),
('00000000-0000-4000-e000-000000000112', '00000000-0000-4000-c000-000000000024', 2, DATE '2026-06-11', '稻荷与祇园', 6.0, 60, 2, 12000.0, 240, 50, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_attractions (
    id, itinerary_id, day_number, poi_id,
    start_time, end_time, stay_duration, "order", cost, transportation, notes
) VALUES
('00000000-0000-4000-f000-000000000121', '00000000-0000-4000-c000-000000000024', 1,
 '00000000-0000-4000-b000-000000000041',
 '2026-06-10T08:00:00+09:00'::timestamptz, '2026-06-10T11:00:00+09:00'::timestamptz, 180, 1, 400.0, NULL, ''),
('00000000-0000-4000-f000-000000000122', '00000000-0000-4000-c000-000000000024', 2,
 '00000000-0000-4000-b000-000000000042',
 '2026-06-11T07:30:00+09:00'::timestamptz, '2026-06-11T10:30:00+09:00'::timestamptz, 180, 1, 0, NULL, '清晨鸟居'),
('00000000-0000-4000-f000-000000000123', '00000000-0000-4000-c000-000000000024', 2,
 '00000000-0000-4000-b000-000000000043',
 '2026-06-11T16:00:00+09:00'::timestamptz, '2026-06-11T18:00:00+09:00'::timestamptz, 120, 2, 0,
 '{"type":"train","distance":6,"duration":30,"cost":220}'::jsonb, '')
ON CONFLICT (id) DO NOTHING;

INSERT INTO day_hotels (id, itinerary_id, day_number, hotel_name, address, check_in_time, check_out_time, cost, room_type, notes) VALUES
('00000000-0000-4000-f000-000000000124', '00000000-0000-4000-c000-000000000024', 1,
 '京都柊家旅馆（意象）', '京都市中京区',
 '2026-06-10T15:00:00+09:00'::timestamptz, '2026-06-11T11:00:00+09:00'::timestamptz, 55000.0, '和室', '')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- user_saved_pois & itinerary favorites & feedback
-- ---------------------------------------------------------------------------
INSERT INTO user_saved_pois (user_id, poi_id, note) VALUES
('00000000-0000-4000-a000-000000000001', '00000000-0000-4000-b000-000000000021', '想去外滩拍夜景'),
('00000000-0000-4000-a000-000000000002', '00000000-0000-4000-a000-000000000011', '北京故宫清单'),
('00000000-0000-4000-a000-000000000003', '00000000-0000-4000-b000-000000000041', '清水寺红叶季')
ON CONFLICT (user_id, poi_id) DO NOTHING;

INSERT INTO user_itinerary_favorites (user_id, itinerary_id) VALUES
('00000000-0000-4000-a000-000000000001', '00000000-0000-4000-c000-000000000022'),
('00000000-0000-4000-a000-000000000002', '00000000-0000-4000-c000-000000000021'),
('00000000-0000-4000-a000-000000000003', '00000000-0000-4000-c000-000000000023')
ON CONFLICT (user_id, itinerary_id) DO NOTHING;

INSERT INTO feedback_events (
    id, idempotency_key, user_id, itinerary_id, poi_id, kind, rating, comment, occurred_at, repeat_count, extra
) VALUES
(
    '00000000-0000-4000-g000-000000000001',
    'seed-poi-rating-shanghai-bund-001',
    '00000000-0000-4000-a000-000000000002',
    NULL,
    '00000000-0000-4000-b000-000000000021',
    'poi_rating',
    5.0,
    '夜景很棒（种子数据）',
    NOW() - INTERVAL '2 days',
    1,
    '{"source":"seed"}'::jsonb
),
(
    '00000000-0000-4000-g000-000000000002',
    'seed-itin-rating-tokyo-023-001',
    '00000000-0000-4000-a000-000000000003',
    '00000000-0000-4000-c000-000000000023',
    NULL,
    'itinerary_rating',
    4.5,
    '行程节奏合理，示例反馈。',
    NOW() - INTERVAL '1 day',
    1,
    '{}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

COMMIT;
