-- +migrate Up
CREATE TABLE "public"."users" (
    id serial NOT NULL,
    username character varying NOT NULL UNIQUE,
    password character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."users";

