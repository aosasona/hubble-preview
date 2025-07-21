CREATE TRIGGER set_updated_at
BEFORE UPDATE ON totp_secrets
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

