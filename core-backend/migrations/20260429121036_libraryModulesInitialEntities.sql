-- Create "library_categories" table
CREATE TABLE "library_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "slug" character varying(200) NOT NULL,
  "icon" character varying(100) NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "parent_category_id" uuid NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_library_categories_child_categories" FOREIGN KEY ("parent_category_id") REFERENCES "library_categories" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);
-- Create index "idx_library_categories_deleted_at" to table: "library_categories"
CREATE INDEX "idx_library_categories_deleted_at" ON "library_categories" ("deleted_at");
-- Create index "idx_library_categories_parent" to table: "library_categories"
CREATE INDEX "idx_library_categories_parent" ON "library_categories" ("parent_category_id");
-- Create index "idx_library_categories_slug_per_parent" to table: "library_categories"
CREATE UNIQUE INDEX "idx_library_categories_slug_per_parent" ON "library_categories" ("parent_category_id", "slug");
-- Create "library_template_downloads" table
CREATE TABLE "library_template_downloads" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "template_id" uuid NOT NULL,
  "group_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_library_template_downloads_account_id" to table: "library_template_downloads"
CREATE INDEX "idx_library_template_downloads_account_id" ON "library_template_downloads" ("account_id");
-- Create index "idx_library_template_downloads_deleted_at" to table: "library_template_downloads"
CREATE INDEX "idx_library_template_downloads_deleted_at" ON "library_template_downloads" ("deleted_at");
-- Create index "idx_library_template_downloads_group_id" to table: "library_template_downloads"
CREATE INDEX "idx_library_template_downloads_group_id" ON "library_template_downloads" ("group_id");
-- Create index "idx_library_template_downloads_template_id" to table: "library_template_downloads"
CREATE INDEX "idx_library_template_downloads_template_id" ON "library_template_downloads" ("template_id");
-- Create "library_category_translations" table
CREATE TABLE "library_category_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "library_category_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_library_categories_translations" FOREIGN KEY ("library_category_id") REFERENCES "library_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_library_cat_trans" to table: "library_category_translations"
CREATE UNIQUE INDEX "idx_library_cat_trans" ON "library_category_translations" ("library_category_id", "language");
-- Create "library_template_groups" table
CREATE TABLE "library_template_groups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "slug" character varying(200) NOT NULL,
  "category_id" uuid NOT NULL,
  "format" character varying(20) NOT NULL,
  "tier_access" character varying(10) NOT NULL DEFAULT 'basic',
  "requires_auth" boolean NOT NULL DEFAULT true,
  "is_active" boolean NOT NULL DEFAULT true,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "default_language" character varying(10) NOT NULL DEFAULT 'en',
  "thumbnail_url" character varying(512) NULL,
  "download_count" bigint NOT NULL DEFAULT 0,
  "created_by" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_library_categories_template_groups" FOREIGN KEY ("category_id") REFERENCES "library_categories" ("id") ON UPDATE CASCADE ON DELETE RESTRICT
);
-- Create index "idx_library_template_groups_category" to table: "library_template_groups"
CREATE INDEX "idx_library_template_groups_category" ON "library_template_groups" ("category_id");
-- Create index "idx_library_template_groups_created_by" to table: "library_template_groups"
CREATE INDEX "idx_library_template_groups_created_by" ON "library_template_groups" ("created_by");
-- Create index "idx_library_template_groups_deleted_at" to table: "library_template_groups"
CREATE INDEX "idx_library_template_groups_deleted_at" ON "library_template_groups" ("deleted_at");
-- Create index "idx_library_template_groups_slug_per_cat" to table: "library_template_groups"
CREATE UNIQUE INDEX "idx_library_template_groups_slug_per_cat" ON "library_template_groups" ("category_id", "slug");
-- Create "library_templates" table
CREATE TABLE "library_templates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "group_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NULL,
  "file_key" character varying(512) NOT NULL,
  "file_url" character varying(512) NULL,
  "file_size" bigint NOT NULL,
  "content_type" character varying(100) NOT NULL,
  "version" bigint NOT NULL DEFAULT 1,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_library_template_groups_templates" FOREIGN KEY ("group_id") REFERENCES "library_template_groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_library_templates_deleted_at" to table: "library_templates"
CREATE INDEX "idx_library_templates_deleted_at" ON "library_templates" ("deleted_at");
-- Create index "idx_library_templates_group" to table: "library_templates"
CREATE INDEX "idx_library_templates_group" ON "library_templates" ("group_id");
-- Create index "idx_library_templates_group_lang" to table: "library_templates"
CREATE UNIQUE INDEX "idx_library_templates_group_lang" ON "library_templates" ("group_id", "language");
-- Create "library_interactive_forms" table
CREATE TABLE "library_interactive_forms" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "template_id" uuid NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "form_layout" jsonb NOT NULL,
  "version" bigint NOT NULL DEFAULT 1,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_library_templates_interactive_form" FOREIGN KEY ("template_id") REFERENCES "library_templates" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_library_interactive_forms_deleted_at" to table: "library_interactive_forms"
CREATE INDEX "idx_library_interactive_forms_deleted_at" ON "library_interactive_forms" ("deleted_at");
-- Create index "idx_library_interactive_forms_template" to table: "library_interactive_forms"
CREATE UNIQUE INDEX "idx_library_interactive_forms_template" ON "library_interactive_forms" ("template_id");
