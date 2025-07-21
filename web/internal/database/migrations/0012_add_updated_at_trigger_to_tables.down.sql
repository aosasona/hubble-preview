DO $$
DECLARE
    tbl_name TEXT;
BEGIN
    FOR tbl_name IN
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_type = 'BASE TABLE'
          AND table_name IN ('users', 'tags', 'workspaces', 'collections', 'collection_members', 'mfa_accounts', 'workspace_members')
    LOOP
	EXECUTE format('DROP TRIGGER IF EXISTS set_updated_at ON %I;', tbl_name);
    END LOOP;
END;
$$ LANGUAGE 'plpgsql';
