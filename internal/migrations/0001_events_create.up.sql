CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK (
        status IN (
            'actual',
            'expired',
            'cancelled'
        )
    ),
    event_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    bookwindow INT NOT NULL DEFAULT 1800 CHECK (bookwindow > 0), -- 30 минут
    total_seats INT NOT NULL CHECK (
        total_seats > 0
        AND total_seats >= avail_seats
    ),
    avail_seats INT NOT NULL CHECK (
        avail_seats >= 0
        AND avail_seats <= total_seats
    )
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    name TEXT,
    surname TEXT,
    tel TEXT,
    email TEXT UNIQUE NOT NULL,
    pass_hash TEXT NOT NULL,
);

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL,
    user_id INT NOT NULL,
    status TEXT NOT NULL CHECK (
        status IN (
            'created',
            'confirmed',
            'cancelled'
        )
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirm_deadline TIMESTAMPTZ NOT NULL,
    --confirmed_at TIMESTAMPTZ,
    CONSTRAINT fk_bookings_events FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_bookings_users FOREIGN KEY (user_id) REFERENCES users (uid) ON UPDATE CASCADE ON DELETE CASCADE
);
-- Индексы
CREATE INDEX idx_bookings_event ON bookings (event_id);

CREATE INDEX idx_bookings_status_created_at ON bookings (status, created_at);

CREATE UNIQUE INDEX idx_users_email ON users (email);