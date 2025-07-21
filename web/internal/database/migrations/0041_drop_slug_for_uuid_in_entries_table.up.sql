ALTER TABLE entries
DROP COLUMN slug,
ADD COLUMN public_id UUID NOT NULL DEFAULT gen_random_uuid();

