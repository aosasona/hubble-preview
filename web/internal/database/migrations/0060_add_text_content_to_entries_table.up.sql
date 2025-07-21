-- Add the new text_content column to the entries table
ALTER TABLE entries
ADD COLUMN text_content text DEFAULT NULL;

-- Add comment on the new column
COMMENT ON COLUMN entries.text_content IS 'The plain text version of the entry. This is used for chunks and indexing.';

