-- Everything in the system is an entry, including comments, etc.
CREATE TABLE IF NOT EXISTS entries (
	id SERIAL PRIMARY KEY,
	origin UUID NOT NULL DEFAULT uuid_generate_v4(), -- origin serves as a unique identifier for the entry, since names can be different across versions of the same entry

	name VARCHAR(255) NOT NULL, -- the name of the entry
	slug TEXT GENERATED ALWAYS AS (collection_id || '-' || slugify(name) || '-v' || version) STORED, -- the unique identifier for the entry (used in the URL)
	content TEXT, -- the content of the entry (e.g. the text of a comment, the text of a direct text entry, etc.)
	file_id VARCHAR(255) DEFAULT NULL, -- the name (AKA ID) of the original file as stored in Minio (only applicable to file entries)
	size INT DEFAULT NULL, -- the size of the file in bytes (only applicable to file entries)
	version INT NOT NULL DEFAULT 1 CHECK (version > 0), -- the version of the entry (incremented each time the entry is updated i.e. a new file is uploaded)
	entry_type VARCHAR(255) NOT NULL, -- the type of the entry (e.g. "comment", "direct_text" "image", "video", "audio", "json", "spreadsheet", etc.)
	checksum VARCHAR(255) DEFAULT NULL, -- the checksum of the file (only applicable to file entries)

	parent_id INT REFERENCES entries(id) ON DELETE CASCADE, -- the parent entry (if any) i.e. the immediate parent of the entry
	collection_id INT NOT NULL REFERENCES collections(id), -- the collection the entry belongs to
	added_by INT NOT NULL REFERENCES users(id), -- the user who added the entry
	last_updated_by INT NOT NULL REFERENCES users(id), -- the user who last updated the entry


-- additional metadata about the entry, this will vary depending on the entry type
-- (e.g. the duration of a video entry, the dimensions of an image entry, etc.)
	meta JSONB DEFAULT '{}',

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NULL,
	deleted_at TIMESTAMPTZ DEFAULT NULL,
	archived_at TIMESTAMPTZ DEFAULT NULL

);

