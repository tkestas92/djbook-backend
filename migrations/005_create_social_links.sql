CREATE TABLE IF NOT EXISTS social_links (
    id         CHAR(36) PRIMARY KEY,
    profile_id CHAR(36) NOT NULL,
    platform   VARCHAR(100) NOT NULL,
    url        VARCHAR(500) NOT NULL,
    CONSTRAINT fk_social_links_profile FOREIGN KEY (profile_id) REFERENCES dj_profiles(id) ON DELETE CASCADE
);
