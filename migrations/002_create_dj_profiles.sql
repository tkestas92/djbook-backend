CREATE TABLE IF NOT EXISTS dj_profiles (
    id         CHAR(36) PRIMARY KEY,
    user_id    CHAR(36) NOT NULL,
    dj_name    VARCHAR(255) NOT NULL,
    bio        TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dj_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
