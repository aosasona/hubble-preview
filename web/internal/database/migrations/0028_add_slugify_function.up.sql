create or replace function slugify(v text)
returns text
language plpgsql
strict
immutable
as
    $function$
BEGIN
  RETURN trim(BOTH '-' FROM regexp_replace(lower(unaccent(trim(v))), '[^a-z0-9\\-_]+', '-', 'gi'));
END;
$function$
;

