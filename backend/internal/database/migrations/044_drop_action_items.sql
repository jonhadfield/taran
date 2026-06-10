-- Action items are no longer extracted or surfaced anywhere in the product.
ALTER TABLE extraction DROP COLUMN IF EXISTS action_items;
ALTER TABLE weekly_summary DROP COLUMN IF EXISTS action_items;
