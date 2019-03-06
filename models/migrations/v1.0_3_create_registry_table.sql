-- +migrate Up
CREATE TABLE "public"."registry" (
    id serial NOT NULL UNIQUE,
    parent_id integer REFERENCES registry(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    key character varying NOT NULL,
    value bytea,
    secure boolean DEFAULT false,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
INSERT INTO "public"."registry" (key, created_at, updated_at)
VALUES ('{ROOT}', now(), now())
;

-- +migrate Down
DROP TABLE "public"."registry";

