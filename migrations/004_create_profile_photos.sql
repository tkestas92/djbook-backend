CREATE TABLE IF NOT EXISTS profile_photos (
    id         CHAR(36) PRIMARY KEY,
    profile_id CHAR(36) NOT NULL,
    url        VARCHAR(500) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_profile_photos_profile FOREIGN KEY (profile_id) REFERENCES dj_profiles(id) ON DELETE CASCADE
);
