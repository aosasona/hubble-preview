-- name: FindOutdatedTotpHashes :many
select id, account_id, hash, version
from totp_secrets
where version < @max_version::int
;

-- name: FindTotpSecretByAccountId :one
select id, account_id, hash, version
from totp_secrets
where account_id = $1
;

-- name: UpdateTotpHash :exec
-- Update the hash and version of the TOTP secret only if the version if lesser than
-- the new version (to prevent redundant updates)
update totp_secrets
set hash = @hash, version = @version
where id = @id and version < @version;

-- name: CreateTotpSecret :one
insert into totp_secrets (account_id, hash, version) values ($1, $2, $3) returning *;

-- name: DeleteTotpSecret :exec
delete from totp_secrets
where id = $1 and account_id = $2
;

