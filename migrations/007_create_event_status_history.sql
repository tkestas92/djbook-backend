CREATE TABLE IF NOT EXISTS event_status_history (
    id         CHAR(36) PRIMARY KEY,
    event_id   CHAR(36) NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_event_status_history_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);
