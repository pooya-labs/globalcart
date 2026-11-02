ALTER TABLE carts
    ADD COLUMN owner_user_id UUID REFERENCES users(id) ON DELETE RESTRICT;

CREATE INDEX carts_owner_user_created_idx
    ON carts(owner_user_id, created_at DESC)
    WHERE owner_user_id IS NOT NULL;
