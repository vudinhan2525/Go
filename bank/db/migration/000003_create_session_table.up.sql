CREATE TABLE "sessions" (
    "id" uuid PRIMARY KEY,
    "user_id" bigserial NOT NULL,
    "refresh_token" VARCHAR NOT NULL,
    "user_agent" VARCHAR NOT NULL,
    "client_ip" VARCHAR NOT NULL,
    "is_blocked" BOOLEAN NOT NULL DEFAULT false,
    "expired_at" timestamptz NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");
