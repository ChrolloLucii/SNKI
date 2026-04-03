-- Seed-данные для локального тестирования и демонстрации
-- Используются фиксированные UUIDs для предсказуемости в dev окружении

-- ============================================================================
-- ДЕМО-ПОЛЬЗОВАТЕЛИ (3 штуки)
-- ============================================================================
INSERT INTO users (id, token, created_at) VALUES
    -- Основной демо-пользователь (из docker-compose DEMO_USER_TOKEN)
    ('11111111-1111-1111-1111-111111111111', 'changeme_user_token', NOW()),
    -- Алиса (забронировала футбол, не оплатила)
    ('22222222-2222-2222-2222-222222222222', 'demo_user_token_alice', NOW()),
    -- Боб (забронировал и оплатил баскетбол)
    ('33333333-3333-3333-3333-333333333333', 'demo_user_token_bob', NOW())
ON CONFLICT (token) DO NOTHING;

-- ============================================================================
-- ДЕМО-СЛОТЫ (5 штук разных видов спорта и районов)
-- ============================================================================
INSERT INTO slots (
    id, sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price, rules_text, status
) VALUES
    -- Футбол в Центральном (открыт, есть места)
    (
        '44444444-4444-4444-4444-444444444444',
        'football', 'Центральный', 'Стадион Динамо', 'ул. Ленина 1',
        NOW() + INTERVAL '3 days',
        NOW() + INTERVAL '3 days' - INTERVAL '2 hours',
        90,
        10, 6,
        500, 700,
        'Casual game, no hard tackles. Bring your own water.',
        'OPEN'
    ),

    -- Баскетбол в Северном (открыт, есть места)
    (
        '55555555-5555-5555-5555-555555555555',
        'basketball', 'Северный', 'Спорткомплекс Север', 'пр. Мира 45',
        NOW() + INTERVAL '5 days',
        NOW() + INTERVAL '5 days' - INTERVAL '3 hours',
        60,
        8, 4,
        600, 900,
        '3x3 format. All skill levels welcome.',
        'OPEN'
    ),

    -- Волейбол в Южном (открыт, много мест)
    (
        '66666666-6666-6666-6666-666666666666',
        'volleyball', 'Южный', 'Зал Олимпик', 'ул. Спортивная 12',
        NOW() + INTERVAL '7 days',
        NOW() + INTERVAL '7 days' - INTERVAL '4 hours',
        120,
        12, 8,
        400, 600,
        'Beach volleyball rules. Indoor court.',
        'OPEN'
    ),

    -- Футбол в Восточном (открыт, почти заполнен для тестирования овербукинга)
    (
        '77777777-7777-7777-7777-777777777777',
        'football', 'Восточный', 'Арена Восток', 'ул. Восточная 99',
        NOW() + INTERVAL '2 days',
        NOW() + INTERVAL '2 days' - INTERVAL '1 hour',
        90,
        6, 4,
        450, 650,
        'Competitive level. Please be on time.',
        'OPEN'
    ),

    -- Баскетбол в Центральном (скоро дедлайн для тестирования urgency)
    (
        '88888888-8888-8888-8888-888888888888',
        'basketball', 'Центральный', 'ТЦ Спорт', 'ул. Центральная 5',
        NOW() + INTERVAL '1 day',
        NOW() + INTERVAL '18 hours',
        90,
        10, 6,
        500, 800,
        '5x5 full court. Intermediate level preferred.',
        'OPEN'
    )
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ДЕМО-УЧАСТИЯ (2 штуки - один RESERVED, один PAID)
-- ============================================================================
INSERT INTO participants (slot_id, user_id, status, reserved_at, paid_at) VALUES
    -- Алиса забронировала футбол, но еще не оплатила (тестирование payment flow)
    (
        '44444444-4444-4444-4444-444444444444',
        '22222222-2222-2222-2222-222222222222',
        'RESERVED',
        NOW() - INTERVAL '1 hour',
        NULL
    ),

    -- Боб забронировал и оплатил баскетбол (полный цикл)
    (
        '55555555-5555-5555-5555-555555555555',
        '33333333-3333-3333-3333-333333333333',
        'PAID',
        NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '1 hour'
    )
ON CONFLICT (slot_id, user_id) DO NOTHING;

-- ============================================================================
-- ДЕМО-ПЛАТЕЖИ (1 штука для PAID участия)
-- ============================================================================
INSERT INTO payments (
    id,
    participant_id,
    idempotency_key,
    amount,
    currency,
    provider,
    status,
    provider_payment_id,
    created_at
)
SELECT
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    p.id,
    'demo-idempotency-key-bob-basketball',
    600,
    'RUB',
    'FAKE',
    'PAID',
    'fake_payment_12345',
    NOW() - INTERVAL '1 hour'
FROM participants p
WHERE p.slot_id = '55555555-5555-5555-5555-555555555555'
  AND p.user_id = '33333333-3333-3333-3333-333333333333'
ON CONFLICT (idempotency_key) DO NOTHING;

-- ============================================================================
-- ПРОВЕРКА SEED-ДАННЫХ (для логов)
-- ============================================================================
DO $$
BEGIN
    RAISE NOTICE 'Seed data loaded:';
    RAISE NOTICE '  Users: % rows', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '  Slots: % rows', (SELECT COUNT(*) FROM slots);
    RAISE NOTICE '  Participants: % rows', (SELECT COUNT(*) FROM participants);
    RAISE NOTICE '  Payments: % rows', (SELECT COUNT(*) FROM payments);
END $$;
