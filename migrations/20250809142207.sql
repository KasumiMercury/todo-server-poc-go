-- Create "tasks" table
CREATE TABLE "tasks" (
  "id" character varying(36) NOT NULL,
  "title" character varying(255) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_tasks_user_id" to table: "tasks"
CREATE INDEX "idx_tasks_user_id" ON "tasks" ("user_id");
