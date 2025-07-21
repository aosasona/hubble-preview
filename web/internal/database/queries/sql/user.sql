-- name: FindUserByUsername :one
select *
from users
where username = $1 and deleted_at is null
;

-- name: FindUserByEmail :one
select *
from users
where email = $1 and deleted_at is null
;

-- name: FindUserById :one
select *
from users
where id = $1 and deleted_at is null
;

-- name: FindUserByPublicId :one
select *
from users
where public_id = $1 and deleted_at is null
;

-- name: UserExistsByUsername :one
select exists (select 1 from users where username = $1)
;

-- name: UserExistsByEmail :one
select exists (select 1 from users where email = $1)
;

-- name: CreateUser :one
insert into users (first_name, last_name, email, username, hashed_password, email_verified) values (initcap(@first_name), initcap(@last_name), lower(@email), lower(@username), @password_hash, @email_verified) returning *;

-- name: VerifyEmail :one
update users set email_verified = true where id = $1 returning *;

-- name: UpdatePassword :one
update users set hashed_password = $2 where id = $1 returning *;

-- name: UpdateProfile :one
update users set
    first_name = initcap(@first_name),
    last_name = initcap(@last_name),
    username = lower(@username)
where id = @id returning *;

-- name: UpdateEmail :one
update users set email = lower(@email) where id = @id returning *;

