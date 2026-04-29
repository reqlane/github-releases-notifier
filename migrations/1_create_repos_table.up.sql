CREATE TABLE IF NOT EXISTS repos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    repo VARCHAR(255) NOT NULL UNIQUE,
    last_seen_tag VARCHAR(255),
    INDEX idx_repo (repo)
);
