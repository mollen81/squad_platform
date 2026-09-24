CREATE TABLE IF NOT EXISTS events (
    event_id                UUID PRIMARY KEY,
    name                    TEXT NOT NULL,
    user_create_id          UUID NOT NULL,
    enemy_side_leader_id    UUID NOT NULL,
    time_start              TIMESTAMP NOT NULL,
    time_finish             TIMESTAMP,
    create_time             TIMESTAMP NOT NULL DEFAULT now(),
    user_count              INT NOT NULL DEFAULT 0,
    target_game_count       INT NOT NULL CHECK (target_game_count BETWEEN 1 AND 3),
    game_count              INT NOT NULL DEFAULT 0,
    status                  TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'in_progress', 'finished', 'declined', 'canceled')),
    winner_side             TEXT NOT NULL DEFAULT '' CHECK (winner_side IN ('', 'ally', 'enemy', 'draw')),

    -- сигнал об аренде сервера уходит один раз за ивент; флаг хранится в БД,
    -- а не в памяти, чтобы после рестарта сервиса он не ушёл повторно
    rent_server_sent        BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_events_time_start ON events(time_start);

CREATE TABLE IF NOT EXISTS users (
    user_event_id        UUID PRIMARY KEY,
    user_id              UUID NOT NULL,
    event_id             UUID NOT NULL REFERENCES events(event_id),
    clan_id              UUID NOT NULL,
    -- сторона, за которую игрок вошёл в ивент: попасть в команду
    -- противоположной стороны нельзя
    enemy                BOOLEAN NOT NULL,
    -- роли здесь нет намеренно: она имеет смысл только в рамках конкретной
    -- игры, поэтому живёт в team_members (см. SetRole)
    join_time            TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT uq_users_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_users_event_id ON users(event_id);

-- games как отдельная таблица больше не нужна: одна "игра" внутри ивента
-- теперь представлена парой строк в team (по одной на каждую сторону,
-- team.game_number объединяет их в рамках одной игры, team.side_leader_id
-- отличает сторону — совпадает либо с events.user_create_id, либо с
-- events.enemy_side_leader_id).

CREATE TABLE IF NOT EXISTS team (
    team_id               UUID PRIMARY KEY,
    event_id              UUID NOT NULL REFERENCES events(event_id),
    winner                BOOLEAN NOT NULL DEFAULT false,
    side_leader_id        UUID NOT NULL REFERENCES users(user_event_id),
    game_number           INT NOT NULL,
    members_count         INT NOT NULL DEFAULT 0,

    -- состояние участия команды в игре:
    --   pending     — игра не началась (только тогда можно менять состав и роли),
    --   in_progress — игра идёт,
    --   finished    — игра закончена (можно заносить статистику игроков)
    status                TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'finished')),

    -- пустой time_start => игра ещё не началась, пустой time_finish => ещё не окончена
    time_start              TIMESTAMP,
    time_finish             TIMESTAMP,

    -- суммарная командная статистика за эту игру (складывается автоматически
    -- из team_members при FinishTeamGame, см. SumTeamMemberStats)
    total_kills                BIGINT NOT NULL DEFAULT 0,
    total_deaths               BIGINT NOT NULL DEFAULT 0,
    total_points               BIGINT NOT NULL DEFAULT 0,
    total_revival              BIGINT NOT NULL DEFAULT 0,
    total_destroyed_vehicles   BIGINT NOT NULL DEFAULT 0,

    -- (event_id, game_number) встречается дважды — по одному team на сторону,
    -- поэтому уникальность считается по тройке вместе с side_leader_id
    CONSTRAINT uq_team_event_game_side UNIQUE (event_id, game_number, side_leader_id)
);

CREATE INDEX IF NOT EXISTS idx_team_event_id ON team(event_id);

CREATE TABLE IF NOT EXISTS team_members (
    team_id         UUID NOT NULL REFERENCES team(team_id),
    user_event_id   UUID NOT NULL REFERENCES users(user_event_id),
    role            TEXT NOT NULL CHECK (role IN ('player', 'squad_leader', 'side_leader')),
    kills           BIGINT NOT NULL DEFAULT 0,
    deaths          BIGINT NOT NULL DEFAULT 0,
    points          BIGINT NOT NULL DEFAULT 0,
    revival             BIGINT NOT NULL DEFAULT 0,
    destroyed_vehicles  BIGINT NOT NULL DEFAULT 0,

    -- Привязан к команде (team_members), а не к ивенту (users): другой
    -- микросервис перед стартом конкретной игры смотрит именно на этот флаг
    -- у участника команды и решает, звать его в игру или нет (true — зовёт,
    -- false — игнорирует). Это не блокирует что-либо в event-service —
    -- такие игроки спокойно существуют в team_members, просто с флагом false.
    six_clan_members    BOOLEAN NOT NULL DEFAULT false,

    PRIMARY KEY (team_id, user_event_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
