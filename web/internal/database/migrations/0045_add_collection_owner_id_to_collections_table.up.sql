-- Add owner_id column to collections table
ALTER TABLE collections
ADD COLUMN owner_id INTEGER;

-- Insert the owner_id for existing collections
UPDATE collections
SET owner_id = (
	SELECT cm.user_id
	FROM collection_members cm
	WHERE cm.collection_id = collections.id AND (cm.bitmask_role & 1073741824) = 1073741824
	ORDER BY cm.created_at DESC LIMIT 1
)
WHERE owner_id IS NULL;

-- Add NOT NULL and FOREIGN KEY constraint to the owner_id column
ALTER TABLE collections
ALTER COLUMN owner_id SET NOT NULL,
ADD CONSTRAINT fk_collections_owner_id FOREIGN KEY (owner_id) REFERENCES users(id)
	ON DELETE CASCADE
	ON UPDATE CASCADE;

