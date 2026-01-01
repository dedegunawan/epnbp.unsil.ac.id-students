-- ENUM untuk JwtType (simulasi tipe enum dari Go)
-- MySQL ENUM literal, bukan tipe custom seperti di PostgreSQL
CREATE TABLE user_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    access_token TEXT,
    refresh_token TEXT,
    expires_at DATETIME,
    token_type VARCHAR(20),
    jwt_type ENUM('keycloak', 'internal') DEFAULT 'keycloak',
    fingerprint TEXT,
    user_agent TEXT,
    ip_address VARCHAR(60),  -- inet tidak ada di MySQL, pakai VARCHAR(60) untuk IPv6
    created_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
