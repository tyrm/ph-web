package models


const sqlUserCount = `
SELECT count(*)
FROM users
WHERE deleted_at IS NULL;`

const sqlUserEstimateCount = `
SELECT n_live_tup
FROM pg_stat_all_tables
WHERE relname = 'users';`

const sqlUserGet = `
SELECT token, username, password, email, login_count, last_login, created_at, updated_at
FROM users
WHERE token = $1 AND deleted_at IS NULL;`

const sqlUserGetByUsername = `
SELECT token, username, password, email, login_count, last_login, created_at, updated_at
FROM users
WHERE lower(username) = lower($1) AND deleted_at IS NULL;`

const sqlUserGetUsernameByID = `
SELECT username
FROM users
WHERE token = $1 AND deleted_at IS NULL;`

const sqlUserIdExists = `
SELECT exists(SELECT 1 FROM users WHERE id=$1);`

const sqlUserInsert = `
INSERT INTO "public"."users" (token, username, password, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING token;`

const sqlUserUpdateLastLogin = `
UPDATE users
SET login_count = login_count + 1, last_login = now()
WHERE token = $1
RETURNING login_count, last_login;`

const sqlUsernameExists = `
SELECT exists(SELECT 1 FROM users WHERE username=$1 AND deleted_at IS NULL);`

const sqlUsersGetPage = `
SELECT token, username, email, login_count, last_login, created_at, updated_at
FROM users WHERE deleted_at IS NULL
ORDER BY created_at asc LIMIT $1 OFFSET $2;`