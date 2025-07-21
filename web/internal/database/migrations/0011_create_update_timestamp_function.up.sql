create or replace function update_timestamp()
returns trigger
as $$
BEGIN
      NEW.updated_at = now();
      RETURN NEW;
END;
$$
language 'plpgsql'
;

