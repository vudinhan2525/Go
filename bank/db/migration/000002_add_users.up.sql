CREATE TYPE user_role AS ENUM ('admin', 'user', 'guest');

CREATE TABLE "users" (
    "user_id" bigserial PRIMARY KEY,
    "hashed_password" varchar NOT NULL,
    "full_name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "role" user_role NOT NULL DEFAULT 'user',
    "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
    "created_at" timestamptz DEFAULT (now()) NOT NULL
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("user_id");

CREATE UNIQUE INDEX ON "accounts" ("owner","currency");