ALTER TABLE "verify_emails" DROP CONSTRAINT IF EXISTS "verify_emails_user_id_fkey";


DROP TABLE IF EXISTS "verify_emails" CASCADE;

ALTER TABLE "users" DROP COLUMN "is_email_verified";