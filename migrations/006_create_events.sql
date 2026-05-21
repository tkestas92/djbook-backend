CREATE TABLE IF NOT EXISTS events (
    id             CHAR(36) PRIMARY KEY,
    profile_id     CHAR(36) NOT NULL,
    title          VARCHAR(255) NOT NULL,
    venue          VARCHAR(255) NOT NULL,
    date           DATE NOT NULL,
    start_time     TIME NOT NULL,
    end_time       TIME,
    notes          TEXT,
    amount_eur     DECIMAL(10,2),
    event_status   ENUM('PENDING','CONFIRMED','COMPLETED','CANCELLED') DEFAULT 'PENDING',
    payment_status ENUM('PAID','UNPAID') DEFAULT 'UNPAID',
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_events_profile FOREIGN KEY (profile_id) REFERENCES dj_profiles(id) ON DELETE CASCADE
);
