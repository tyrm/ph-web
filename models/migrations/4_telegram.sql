-- +migrate Up
CREATE TABLE "public"."tg_users" (
    id serial NOT NULL UNIQUE,
    api_id integer NOT NULL UNIQUE,
    is_bot bool NOT NULL,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
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

CREATE TABLE "public"."tg_chats" (
    id serial NOT NULL UNIQUE,
    api_id integer NOT NULL UNIQUE,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;
CREATE TABLE "public"."tg_chats_history" (
    id serial NOT NULL UNIQUE,
    tgc_id integer NOT NULL REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type integer NOT NULL,
    title character varying,
    username character varying,
    first_name character varying,
    last_name character varying,
    all_members_are_administrators bool,
    created_at timestamp without time zone NOT NULL,
    last_seen timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_messages" (
    id serial NOT NULL UNIQUE,
    message_id integer NOT NULL,
    "from" integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    date timestamp without time zone NOT NULL,
    chat integer NOT NULL REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from_chat integer REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from_message_id integer,
    forward_signature character varying,
    forward_date timestamp without time zone,
    reply_to_message integer REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    edit_date timestamp without time zone,
    media_group_id character varying,
    author_signature character varying,
    text character varying,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
);
CREATE TABLE "public"."tg_message_entities" (
    id serial NOT NULL UNIQUE,
    tgm_id integer NOT NULL REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type integer NOT NULL,
    "offset" integer,
    length integer,
    url character varying,
    "user" integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."tg_message_entities";
DROP TABLE "public"."tg_messages";
DROP TABLE "public"."tg_chats_history";
DROP TABLE "public"."tg_chats";
DROP TABLE "public"."tg_users_history";
DROP TABLE "public"."tg_users";

