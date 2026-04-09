CREATE TABLE IF NOT EXISTS subscriptions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    last_seen_tag VARCHAR(255) NOT NULL DEFAULT '',
    confirm_token VARCHAR(255),
    unsubscribe_token VARCHAR(255) NOT NULL,
    UNIQUE KEY unique_email_repo (email, repo)
);
