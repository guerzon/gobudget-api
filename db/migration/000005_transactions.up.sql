CREATE TABLE "transactions" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
  "account_id" uuid NOT NULL,
  "date" date NOT NULL,
  "payee_id" uuid NOT NULL,
  "category_id" uuid,
  "memo" varchar,
  "amount" int NOT NULL,
  "approved" boolean NOT NULL DEFAULT true,
  "cleared" boolean NOT NULL DEFAULT false,
  "reconciled" boolean NOT NULL DEFAULT false
);

CREATE INDEX ON "transactions" ("id");

CREATE INDEX ON "transactions" ("account_id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("payee_id") REFERENCES "payees" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

CREATE VIEW transactions_view AS
select
	trans.id, trans.account_id, acc.name "account_name", acc.budget_id, trans.date, trans.payee_id, p.name "payee_name", trans.category_id, c.name "category_name", trans.memo, trans.amount, trans.approved, trans.cleared, trans.reconciled
from transactions trans, accounts acc, payees p, categories c
where trans.account_id = acc.id
and trans.payee_id = p.id
and trans.category_id = c.id;
