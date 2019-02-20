-- +migrate Up
CREATE TABLE "public"."registry" (
    id serial NOT NULL UNIQUE,
    parent_id integer REFERENCES registry(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    key character varying NOT NULL,
    value character varying,
    secure boolean DEFAULT false,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
INSERT INTO "public"."registry" (key, created_at, updated_at)
VALUES ('{ROOT}', now(), now())
;
CREATE TABLE "public"."registry_log" (
    id serial NOT NULL UNIQUE,
    reg_id integer REFERENCES registry(id) ON DELETE CASCADE ON UPDATE RESTRICT,
    user_id integer REFERENCES users(id) ON DELETE CASCADE ON UPDATE RESTRICT,
    timestamp timestamp without time zone NOT NULL,
    change_type integer NOT NULL,
    old_value character varying,
    new_value character varying,
    PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."registry_log";
DROP TABLE "public"."registry";

