CREATE TABLE IF NOT EXISTS releases (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  profile_id CHAR(36) NOT NULL,
  title VARCHAR(255) NOT NULL,
  artist VARCHAR(255) NOT NULL,
  artwork_url VARCHAR(500),
  song_link_url VARCHAR(500) NOT NULL,
  platforms_json TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (profile_id) REFERENCES dj_profiles(id)
);
