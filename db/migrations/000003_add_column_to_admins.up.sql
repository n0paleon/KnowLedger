ALTER TABLE admins ADD COLUMN api_key VARCHAR(100) DEFAULT '';

CREATE INDEX idx_admin_api_key ON admins(api_key);