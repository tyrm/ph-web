-- +migrate Up
CREATE TABLE "public"."tg_users" (
    id serial NOT NULL UNIQUE,
    api_id integer NOT NULL UNIQUE,
    is_bot bool NOT NULL,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
CREATE INDEX ON "public"."tg_users" USING btree ("api_id");
CREATE TABLE "public"."tg_users_history" (
    id serial NOT NULL UNIQUE,
    tgu_id integer NOT NULL REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    first_name character varying NOT NULL,
    last_name character varying,
    username character varying,
    language_code character varying,
    created_at timestamp without time zone NOT NULL,
    last_seen timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
CREATE INDEX ON "public"."tg_users_history" USING btree ("tgu_id");

CREATE TABLE "public"."tg_chats" (
    id serial NOT NULL UNIQUE,
    api_id bigint NOT NULL UNIQUE,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
CREATE INDEX ON "public"."tg_chats" USING btree ("api_id");
CREATE TABLE "public"."tg_chats_history" (
    id serial NOT NULL UNIQUE,
    tgc_id integer NOT NULL REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type character varying NOT NULL,
    title character varying,
    username character varying,
    first_name character varying,
    last_name character varying,
    all_members_are_admins bool NOT NULL,
    created_at timestamp without time zone NOT NULL,
    last_seen timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
CREATE INDEX ON "public"."tg_chats_history" USING btree ("tgc_id");

-- +migrate Down
DROP TABLE "public"."tg_chats_history";
DROP TABLE "public"."tg_chats";
DROP TABLE "public"."tg_users_history";
DROP TABLE "public"."tg_users";
