-- Drop the old trigger and function
DROP TRIGGER IF EXISTS entry_chunks_text_vector_update ON entry_chunks;
drop function if exists update_text_vector
;

-- Change the type of the language column to varchar
ALTER TABLE entry_chunks
DROP COLUMN language;

ALTER TABLE entry_chunks
ADD COLUMN language varchar DEFAULT 'simple'::varchar;

-- Add method to cast language to regconfig
create or replace function ts_regconfig(lang varchar)
returns regconfig
as $$
begin
  if lang is null or lang = '' then
     lang := 'simple';
  end if;
  return lang::regconfig;
exception
  when others then
    return 'simple'::regconfig;
end;
$$
language plpgsql
immutable
;

-- Add triggers to update the vectors
create function update_text_vector()
returns trigger
as $$
BEGIN
NEW.text_vector := to_tsvector(ts_regconfig(NEW.language), NEW.content);
RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER entry_chunks_text_vector_update
BEFORE INSERT OR UPDATE OF content, language ON entry_chunks
FOR EACH ROW EXECUTE FUNCTION update_text_vector();

