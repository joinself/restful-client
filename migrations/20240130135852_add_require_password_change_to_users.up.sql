ALTER TABLE account
ADD COLUMN requires_password_change INTEGER DEFAULT 0 NOT NULL;
