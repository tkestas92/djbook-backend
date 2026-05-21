CREATE TABLE IF NOT EXISTS profile_genres (
    id         CHAR(36) PRIMARY KEY,
    profile_id CHAR(36) NOT NULL,
    genre      VARCHAR(100) NOT NULL,
    CONSTRAINT fk_profile_genres_profile FOREIGN KEY (profile_id) REFERENCES dj_profiles(id) ON DELETE CASCADE
);
