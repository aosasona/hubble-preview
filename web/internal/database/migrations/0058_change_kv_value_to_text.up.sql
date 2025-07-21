ALTER TABLE plugins_kv
ALTER COLUMN value TYPE text
USING value::text;

