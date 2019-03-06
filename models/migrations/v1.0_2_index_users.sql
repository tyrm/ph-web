-- +migrate Up
CREATE INDEX IF NOT EXISTS users_lower_username_idx
    ON public.users USING btree (lower(username))
    TABLESPACE pg_default;

CREATE UNIQUE INDEX IF NOT EXISTS users_id_idx
    ON public.users USING btree (id)
    TABLESPACE pg_default;

-- +migrate Down
DROP INDEX users_id_idx;
DROP INDEX users_lower_username_idx;