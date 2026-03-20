-- Create "guide_categories" table
CREATE TABLE "guide_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "slug" character varying(200) NOT NULL,
  "icon" character varying(100) NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "parent_category_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_categories_child_categories" FOREIGN KEY ("parent_category_id") REFERENCES "guide_categories" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);
-- Create index "idx_guide_categories_deleted_at" to table: "guide_categories"
CREATE INDEX "idx_guide_categories_deleted_at" ON "guide_categories" ("deleted_at");
-- Create index "idx_guide_categories_parent" to table: "guide_categories"
CREATE INDEX "idx_guide_categories_parent" ON "guide_categories" ("parent_category_id");
-- Create index "idx_guide_categories_slug_per_parent" to table: "guide_categories"
CREATE UNIQUE INDEX "idx_guide_categories_slug_per_parent" ON "guide_categories" ("parent_category_id", "slug");
-- Create "guide_category_conditions" table
CREATE TABLE "guide_category_conditions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "category_id" uuid NOT NULL,
  "condition_type" character varying(50) NOT NULL,
  "operator" character varying(20) NOT NULL,
  "condition_value" jsonb NOT NULL,
  "is_inverse" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_categories_conditions" FOREIGN KEY ("category_id") REFERENCES "guide_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_guide_category_conditions_category" to table: "guide_category_conditions"
CREATE INDEX "idx_guide_category_conditions_category" ON "guide_category_conditions" ("category_id");
-- Create index "idx_guide_category_conditions_deleted_at" to table: "guide_category_conditions"
CREATE INDEX "idx_guide_category_conditions_deleted_at" ON "guide_category_conditions" ("deleted_at");
-- Create "guide_category_translations" table
CREATE TABLE "guide_category_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guide_category_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_categories_translations" FOREIGN KEY ("guide_category_id") REFERENCES "guide_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_cat_trans" to table: "guide_category_translations"
CREATE UNIQUE INDEX "idx_cat_trans" ON "guide_category_translations" ("guide_category_id", "language");
-- Create "guides" table
CREATE TABLE "guides" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "category_id" uuid NOT NULL,
  "slug" character varying(200) NOT NULL,
  "icon" character varying(100) NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_categories_guides" FOREIGN KEY ("category_id") REFERENCES "guide_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_guides_category" to table: "guides"
CREATE INDEX "idx_guides_category" ON "guides" ("category_id");
-- Create index "idx_guides_deleted_at" to table: "guides"
CREATE INDEX "idx_guides_deleted_at" ON "guides" ("deleted_at");
-- Create index "idx_guides_slug_per_category" to table: "guides"
CREATE UNIQUE INDEX "idx_guides_slug_per_category" ON "guides" ("category_id", "slug");
-- Create "guide_conditions" table
CREATE TABLE "guide_conditions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "guide_id" uuid NOT NULL,
  "condition_type" character varying(50) NOT NULL,
  "operator" character varying(20) NOT NULL,
  "condition_value" jsonb NOT NULL,
  "is_inverse" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guides_conditions" FOREIGN KEY ("guide_id") REFERENCES "guides" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_guide_conditions_deleted_at" to table: "guide_conditions"
CREATE INDEX "idx_guide_conditions_deleted_at" ON "guide_conditions" ("deleted_at");
-- Create index "idx_guide_conditions_guide" to table: "guide_conditions"
CREATE INDEX "idx_guide_conditions_guide" ON "guide_conditions" ("guide_id");
-- Create "guide_steps" table
CREATE TABLE "guide_steps" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "guide_id" uuid NOT NULL,
  "slug" character varying(200) NOT NULL,
  "step_type" character varying(64) NOT NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "estimated_time" bigint NULL,
  "difficulty_level" bigint NULL,
  "is_optional" boolean NOT NULL DEFAULT false,
  "required_documents" jsonb NOT NULL DEFAULT '[]',
  "external_links" jsonb NOT NULL DEFAULT '[]',
  "fee_estimate" bigint NULL,
  "version" bigint NOT NULL DEFAULT 1,
  "effective_date" date NOT NULL DEFAULT CURRENT_DATE,
  "expiry_date" date NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guides_steps" FOREIGN KEY ("guide_id") REFERENCES "guides" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "difficulty_level_check" CHECK ((difficulty_level >= 1) AND (difficulty_level <= 5))
);
-- Create index "idx_guide_steps_deleted_at" to table: "guide_steps"
CREATE INDEX "idx_guide_steps_deleted_at" ON "guide_steps" ("deleted_at");
-- Create index "idx_guide_steps_guide" to table: "guide_steps"
CREATE INDEX "idx_guide_steps_guide" ON "guide_steps" ("guide_id");
-- Create index "idx_guide_steps_slug_per_guide" to table: "guide_steps"
CREATE UNIQUE INDEX "idx_guide_steps_slug_per_guide" ON "guide_steps" ("guide_id", "slug");
-- Create index "idx_guide_steps_sort_per_guide" to table: "guide_steps"
CREATE UNIQUE INDEX "idx_guide_steps_sort_per_guide" ON "guide_steps" ("guide_id", "sort_order");
-- Create "guide_step_translations" table
CREATE TABLE "guide_step_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guide_step_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NULL,
  "detailed_content" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_translations" FOREIGN KEY ("guide_step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_step_trans" to table: "guide_step_translations"
CREATE UNIQUE INDEX "idx_step_trans" ON "guide_step_translations" ("guide_step_id", "language");
-- Create "guide_step_versions" table
CREATE TABLE "guide_step_versions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "step_id" uuid NOT NULL,
  "version" bigint NOT NULL,
  "title" character varying(200) NOT NULL,
  "content" jsonb NOT NULL,
  "effective_date" date NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_versions" FOREIGN KEY ("step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_guide_step_versions_deleted_at" to table: "guide_step_versions"
CREATE INDEX "idx_guide_step_versions_deleted_at" ON "guide_step_versions" ("deleted_at");
-- Create index "idx_version_step_version" to table: "guide_step_versions"
CREATE UNIQUE INDEX "idx_version_step_version" ON "guide_step_versions" ("step_id", "version");
-- Create index "idx_versions_step" to table: "guide_step_versions"
CREATE INDEX "idx_versions_step" ON "guide_step_versions" ("step_id");
-- Create "guide_translations" table
CREATE TABLE "guide_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guide_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guides_translations" FOREIGN KEY ("guide_id") REFERENCES "guides" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_guide_trans" to table: "guide_translations"
CREATE UNIQUE INDEX "idx_guide_trans" ON "guide_translations" ("guide_id", "language");
-- Create "step_conditions" table
CREATE TABLE "step_conditions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "step_id" uuid NOT NULL,
  "condition_type" character varying(50) NOT NULL,
  "operator" character varying(20) NOT NULL,
  "condition_value" jsonb NOT NULL,
  "is_inverse" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_conditions" FOREIGN KEY ("step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_step_conditions_deleted_at" to table: "step_conditions"
CREATE INDEX "idx_step_conditions_deleted_at" ON "step_conditions" ("deleted_at");
-- Create index "idx_step_conditions_step" to table: "step_conditions"
CREATE INDEX "idx_step_conditions_step" ON "step_conditions" ("step_id");
-- Create "step_dependencies" table
CREATE TABLE "step_dependencies" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "step_id" uuid NOT NULL,
  "required_step_id" uuid NOT NULL,
  "dependency_type" character varying(20) NOT NULL DEFAULT 'MANDATORY',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_dependencies" FOREIGN KEY ("step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_guide_steps_reverse_dependencies" FOREIGN KEY ("required_step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_step_dependencies_deleted_at" to table: "step_dependencies"
CREATE INDEX "idx_step_dependencies_deleted_at" ON "step_dependencies" ("deleted_at");
-- Create index "idx_step_dependencies_required" to table: "step_dependencies"
CREATE INDEX "idx_step_dependencies_required" ON "step_dependencies" ("required_step_id");
-- Create index "idx_step_dependencies_step" to table: "step_dependencies"
CREATE INDEX "idx_step_dependencies_step" ON "step_dependencies" ("step_id");
-- Create index "idx_step_dependencies_unique" to table: "step_dependencies"
CREATE UNIQUE INDEX "idx_step_dependencies_unique" ON "step_dependencies" ("step_id", "required_step_id");
-- Create "user_guide_bookmarks" table
CREATE TABLE "user_guide_bookmarks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "step_id" uuid NOT NULL,
  "note" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_bookmarks" FOREIGN KEY ("step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_bookmark_user_step" to table: "user_guide_bookmarks"
CREATE UNIQUE INDEX "idx_bookmark_user_step" ON "user_guide_bookmarks" ("user_id", "step_id");
-- Create index "idx_bookmarks_step" to table: "user_guide_bookmarks"
CREATE INDEX "idx_bookmarks_step" ON "user_guide_bookmarks" ("step_id");
-- Create index "idx_bookmarks_user" to table: "user_guide_bookmarks"
CREATE INDEX "idx_bookmarks_user" ON "user_guide_bookmarks" ("user_id");
-- Create index "idx_user_guide_bookmarks_deleted_at" to table: "user_guide_bookmarks"
CREATE INDEX "idx_user_guide_bookmarks_deleted_at" ON "user_guide_bookmarks" ("deleted_at");
-- Create "user_guide_journeys" table
CREATE TABLE "user_guide_journeys" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "guide_id" uuid NOT NULL,
  "journey_hash" text NULL,
  "step_sequence" jsonb NOT NULL,
  "total_steps" bigint NOT NULL,
  "completed_steps" bigint NOT NULL DEFAULT 0,
  "estimated_total_time" bigint NULL,
  "generated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guides_journeys" FOREIGN KEY ("guide_id") REFERENCES "guides" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_journey_expires" to table: "user_guide_journeys"
CREATE INDEX "idx_journey_expires" ON "user_guide_journeys" ("expires_at");
-- Create index "idx_journey_guide" to table: "user_guide_journeys"
CREATE INDEX "idx_journey_guide" ON "user_guide_journeys" ("guide_id");
-- Create index "idx_journey_user" to table: "user_guide_journeys"
CREATE INDEX "idx_journey_user" ON "user_guide_journeys" ("user_id");
-- Create index "idx_journey_user_guide" to table: "user_guide_journeys"
CREATE UNIQUE INDEX "idx_journey_user_guide" ON "user_guide_journeys" ("user_id", "guide_id");
-- Create index "idx_user_guide_journeys_deleted_at" to table: "user_guide_journeys"
CREATE INDEX "idx_user_guide_journeys_deleted_at" ON "user_guide_journeys" ("deleted_at");
-- Create "user_guide_progresses" table
CREATE TABLE "user_guide_progresses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "step_id" uuid NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'LOCKED',
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "time_spent" bigint NULL,
  "notes" text NULL,
  "uploaded_documents" jsonb NOT NULL DEFAULT '[]',
  "last_accessed_at" timestamptz NULL,
  "version" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_guide_steps_progress" FOREIGN KEY ("step_id") REFERENCES "guide_steps" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_user_guide_progresses_deleted_at" to table: "user_guide_progresses"
CREATE INDEX "idx_user_guide_progresses_deleted_at" ON "user_guide_progresses" ("deleted_at");
-- Create index "idx_user_progress_status" to table: "user_guide_progresses"
CREATE INDEX "idx_user_progress_status" ON "user_guide_progresses" ("status");
-- Create index "idx_user_progress_step" to table: "user_guide_progresses"
CREATE INDEX "idx_user_progress_step" ON "user_guide_progresses" ("step_id");
-- Create index "idx_user_progress_user" to table: "user_guide_progresses"
CREATE INDEX "idx_user_progress_user" ON "user_guide_progresses" ("user_id");
