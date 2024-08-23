CREATE TABLE "category_groups" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "budget_id" uuid NOT NULL,
  "name" varchar NOT NULL
);

CREATE TABLE "categories" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "category_group_id" uuid NOT NULL,
  "name" varchar NOT NULL
);

CREATE TABLE "payees" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "budget_id" uuid NOT NULL,
  "name" varchar NOT NULL
);

ALTER TABLE "payees" ADD FOREIGN KEY ("budget_id") REFERENCES "budgets" ("id");

ALTER TABLE "category_groups" ADD FOREIGN KEY ("budget_id") REFERENCES "budgets" ("id");

ALTER TABLE "categories" ADD FOREIGN KEY ("category_group_id") REFERENCES "category_groups" ("id");
