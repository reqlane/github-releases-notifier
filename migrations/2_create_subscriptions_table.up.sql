CREATE TABLE IF NOT EXISTS subscriptions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    repo_id INT NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    confirm_token VARCHAR(255),
    unsubscribe_token VARCHAR(255) NOT NULL,
    UNIQUE KEY idx_email_repo (email, repo_id),
    INDEX idx_confirm_token (confirm_token),
    INDEX idx_unsubscribe_token (unsubscribe_token),
    FOREIGN KEY (repo_id) REFERENCES repos(id)
);
