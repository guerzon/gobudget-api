CREATE TABLE "budgets" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "owner_username" varchar NOT NULL,
  "name" varchar NOT NULL,
  "currency_code" varchar NOT NULL DEFAULT 'USD'
);

CREATE TABLE "accounts" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "budget_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "type" varchar NOT NULL,
  "closed" boolean NOT NULL DEFAULT false,
  "note" varchar,
  "balance" int NOT NULL DEFAULT 0,
  "cleared_balance" int NOT NULL DEFAULT 0,
  "uncleared_balance" int NOT NULL DEFAULT 0,
  "last_reconciled_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "budgets" ADD FOREIGN KEY ("owner_username") REFERENCES "users" ("username");

ALTER TABLE "accounts" ADD FOREIGN KEY ("budget_id") REFERENCES "budgets" ("id");
