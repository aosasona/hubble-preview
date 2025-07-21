-- name: MfaEnabled :one
select (count(m.id) > 0) as enabled
from mfa_accounts as m
where m.user_id = $1 and m.active = true
;

-- name: FindActiveMfaAccountIdsByUserId :many
select id
from mfa_accounts
where user_id = $1 and active = true
order by created_at desc
;

-- name: FindActiveMfaAccountsByUserId :many
select id, account_type, active, meta, user_id, created_at, last_used_at, preferred
from mfa_accounts
where user_id = $1 and active = true
order by created_at asc
;

-- name: FindAllMfaAccountsByUserId :many
select id, account_type, active, meta, created_at, last_used_at, preferred
from mfa_accounts
where user_id = $1
order by created_at asc
;

-- name: FindPreferredMfaAccountByUserId :one
-- Find the one marked as preferred, if any, or the first one created and active if
-- none are marked as preferred
select id, account_type, active, meta, user_id, created_at, last_used_at, preferred
from mfa_accounts
where user_id = $1 and active = true
order by preferred desc, created_at desc
limit 1
;

-- name: FindMfaAccountByEmail :one
select id, account_type, active, meta, user_id, created_at, last_used_at, preferred
from mfa_accounts
where account_type = 'email' and (meta ->> 'email')::text = @email::text
;

-- name: FindMfaAccountById :one
select id, account_type, active, meta, user_id, created_at, last_used_at, preferred
from mfa_accounts
where id = $1
;

-- name: CreateEmailMfaAccount :one
insert into mfa_accounts (user_id, account_type, meta, active, preferred)
values ($1, 'email', $2, false, false)
returning *;

-- name: CreateTotpMfaAccount :one
insert into mfa_accounts (user_id, account_type, meta, active, preferred)
values ($1, 'totp', $2, true, false)
returning *;

-- name: ActivateMfaAccount :exec
-- Activate the MFA account and set it as preferred if it is the only one (i.e. none
-- are active)
update mfa_accounts m
set active = true, preferred = (select count(*) = 0 from mfa_accounts ml where ml.user_id = $1 and ml.active = true)
where m.id = $2 and m.user_id = $1;


-- name: DeleteMfaAccount :exec
delete from mfa_accounts
where id = $1 and user_id = $2
;

-- name: RenameMfaAccount :one
update mfa_accounts
set meta = jsonb_set(meta, '{name}', to_jsonb(@new_name::text), true)
where id = @account_id and user_id = @user_id returning *;

-- name: MfaAccountNameExists :one
select count(id) > 0 as exists
from mfa_accounts
where lower(meta ->> 'name')::text = lower(@name::text) and user_id = @user_id
;

-- name: SetPreferredMfaAccount :exec
update mfa_accounts
set preferred = case
    when id = $1 then true
    else false
end
where user_id = $2;

-- name: SetMfaAccountLastUsed :exec
update mfa_accounts
set last_used_at = now()
where id = $1;

-- name: DeleteAllBackupCodes :exec
delete from mfa_backup_tokens
where user_id = $1
;

-- name: MarkBackupCodeAsUsed :exec
update mfa_backup_tokens
set used_at = now()
where user_id = @user_id and id = @code_id;

-- name: SaveBackupCodes :exec
insert into mfa_backup_tokens(user_id, hashed_token) values ($1, $2);

-- name: FindBackupCodesByUserId :many
select id, hashed_token, used_at
from mfa_backup_tokens
where user_id = $1
;

-- name: CanGenerateBackupCodes :one
-- Check if the backup codes were generated less then 7 days ago
select count(id) = 0 as can_generate
from mfa_backup_tokens
where user_id = $1 and created_at > now() - interval '7 days'
;

