-- +migrate Up
CREATE TABLE "public"."users" (
    id character(8) NOT NULL UNIQUE,
    username character varying NOT NULL,
    password character varying NOT NULL,
    login_count integer NOT NULL DEFAULT 0,
    last_login timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."users";

