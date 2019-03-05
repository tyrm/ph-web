-- +migrate Up
CREATE TABLE "public"."tg_photo_sizes" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  width integer NOT NULL,
  height integer NOT NULL,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_animations" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  thumbnail integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  file_name character varying,
  mime_types character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_audios" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  duration integer NOT NULL,
  performer character varying,
  title character varying,
  mime_type character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_chat_animations" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  width integer NOT NULL,
  height integer NOT NULL,
  duration integer NOT NULL,
  thumbnail_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  file_name character varying,
  mime_type character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_contacts" (
  id serial NOT NULL UNIQUE,
  phone_number character varying NOT NULL,
  first_name character varying NOT NULL,
  last_name character varying,
  user_id integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_documents" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  thumbnail_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  file_name character varying,
  mime_type character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_games" (
  id serial NOT NULL UNIQUE,
  title character varying NOT NULL UNIQUE,
  file_name character varying,
  mime_types character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_game_thumbnails" (
  id serial NOT NULL UNIQUE,
  tgg_id integer NOT NULL REFERENCES tg_games(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  tgps_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  created_at timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_locations" (
  id serial NOT NULL UNIQUE,
  longitude double precision,
  latitude double precision,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_stickers" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  width integer NOT NULL,
  height integer NOT NULL,
  thumbnail_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  emoji character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  set_name character varying,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_venues" (
  id serial NOT NULL UNIQUE,
  location_id integer NOT NULL REFERENCES tg_locations(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  title character varying NOT NULL,
  address character varying NOT NULL,
  foursquare_id character varying,
  foursquare_type character varying,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_videos" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  width integer NOT NULL,
  height integer NOT NULL,
  duration integer NOT NULL,
  thumbnail_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  mime_type character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_video_notes" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  length integer NOT NULL,
  duration integer NOT NULL,
  thumbnail_id integer REFERENCES tg_photo_sizes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

CREATE TABLE "public"."tg_voices" (
  id serial NOT NULL UNIQUE,
  file_id character varying NOT NULL UNIQUE,
  duration integer NOT NULL,
  mime_type character varying,
  file_size integer,
  file_location character varying,
  file_suffix character varying,
  file_retrieved_at timestamp without time zone,
  created_at timestamp without time zone NOT NULL,
  last_seen timestamp without time zone NOT NULL,
  PRIMARY KEY ("id")
)
;

-- +migrate Down
DROP TABLE "public"."tg_voices";
DROP TABLE "public"."tg_video_notes";
DROP TABLE "public"."tg_videos";
DROP TABLE "public"."tg_stickers";
DROP TABLE "public"."tg_game_thumbnails";
DROP TABLE "public"."tg_games";
DROP TABLE "public"."tg_documents";
DROP TABLE "public"."tg_contacts";
DROP TABLE "public"."tg_chat_animations";
DROP TABLE "public"."tg_audios";
DROP TABLE "public"."tg_photo_sizes";
