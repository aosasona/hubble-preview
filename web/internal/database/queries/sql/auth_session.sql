-- name: FindValidSessionByToken :one
select *
from auth_sessions
where token = $1 and expires_at > now() + interval '1 second'
;


-- name: FindValidSessionByUserId :one
select *
from auth_sessions
where user_id = $1 and expires_at > now() + interval '1 second'
;

-- name: CreateSession :one
insert into auth_sessions (user_id, token, expires_at) values ($1, $2, $3) returning *;

