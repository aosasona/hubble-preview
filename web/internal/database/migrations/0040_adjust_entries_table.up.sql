ALTER TABLE entries
DROP COLUMN slug;


ALTER TABLE entries
ALTER COLUMN name TYPE TEXT,
ADD COLUMN slug TEXT GENERATED ALWAYS AS (collection_id || '-' || slugify(name) || '-v' || version) STORED;  -- the unique identifier for the entry (used in the URL)

