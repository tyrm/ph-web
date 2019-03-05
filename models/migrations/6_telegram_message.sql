-- +migrate Up
CREATE TABLE "public"."tg_messages" (
    id serial NOT NULL UNIQUE,
    message_id integer NOT NULL,
    "from_id" integer NOT NULL REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    date timestamp without time zone NOT NULL,
    chat_id integer NOT NULL REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from_id integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from_chat_id integer REFERENCES tg_chats(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    forwarded_from_message_id integer,
    forward_signature character varying,
    forward_date timestamp without time zone,
    reply_to_message integer REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    edit_date timestamp without time zone,
    text character varying,
    audio_id integer REFERENCES tg_audios(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    document_id integer REFERENCES tg_documents(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    animation_id integer REFERENCES tg_chat_animations(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    game_id integer REFERENCES tg_games(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    sticker_id integer REFERENCES tg_stickers(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    video_id integer REFERENCES tg_videos(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    video_note_id integer REFERENCES tg_video_notes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    voice_id integer REFERENCES tg_voices(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    caption character varying,
    contact_id integer REFERENCES tg_contacts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    location_id integer REFERENCES tg_locations(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    venue_id integer REFERENCES tg_venues(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    left_chat_member_id integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    new_chat_title character varying,
    delete_chat_photo boolean,
    group_chat_created boolean,
    supergroup_chat_created boolean,
    channel_chat_created boolean,
    migrate_to_chat_id bigint,
    migrate_from_chat_id bigint,
    pinned_message_id integer REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_message_entities" (
    id serial NOT NULL UNIQUE,
    tgm_id integer NOT NULL REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type character varying NOT NULL,
    "offset" integer,
    length integer,
    url character varying,
    "user" integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_message_new_chat_members" (
    id serial NOT NULL UNIQUE,
    tgm_id integer NOT NULL REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    tgu_id integer REFERENCES tg_users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_message_new_chat_photos" (
    id serial NOT NULL UNIQUE,
    tgm_id integer NOT NULL REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    tgps_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_message_photos" (
    id serial NOT NULL UNIQUE,
    tgm_id integer NOT NULL REFERENCES tg_messages(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    tgps_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at timestamp without time zone NOT NULL,
    PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."tg_message_entities";
DROP TABLE "public"."tg_messages";
