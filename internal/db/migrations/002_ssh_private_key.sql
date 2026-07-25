ALTER TABLE instances ALTER COLUMN ssh_password_enc DROP NOT NULL;
ALTER TABLE instances ADD COLUMN IF NOT EXISTS ssh_private_key_enc BYTEA;
