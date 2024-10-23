CREATE TABLE IF NOT EXISTS "manifest"
(
    "srt_file"    TEXT,
    "srt_mod_time" TIMESTAMP
);

CREATE UNIQUE INDEX unique_srt ON manifest ("srt_file", "srt_mod_time");

CREATE TABLE IF NOT EXISTS "dialog"
(
    "id"              TEXT PRIMARY KEY,
    "media_id"        TEXT      NOT NULL,
    "pos"             INTEGER   NOT NULL,
    "start_timestamp" TIMESTAMP NULL,
    "end_timestamp"   INTEGER   NULL,
    "content"         TEXT      NOT NULL,
    "media_file_name" TEXT      NOT NULL
);

CREATE INDEX dialog_pos ON dialog ("pos");
CREATE INDEX ts ON dialog ("start_timestamp");
