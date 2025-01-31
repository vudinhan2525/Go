
DROP INDEX IF EXISTS "accounts_owner_currency_key";


ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";


DROP TABLE IF EXISTS "users";

DROP TYPE IF EXISTS user_role;
