package registry

const sqlCreateEntry = `
INSERT INTO "public"."registry" (parent_id, key, value, secure, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING  id, parent_id, key, value, secure, created_at, updated_at;`

const sqlDeleteByID = `
DELETE FROM registry
WHERE id = $1`

const sqlGetPathByID = `
WITH RECURSIVE subdirectories AS (
SELECT id, parent_id, key
FROM registry WHERE id = $1
UNION  SELECT r.id, r.parent_id, r.key
FROM registry r INNER JOIN subdirectories s ON s.parent_id = r.id
) SELECT * FROM subdirectories ORDER BY id desc;`

const sqlGetRegistryByID = `
SELECT r1.id, r1.parent_id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.id = $1
GROUP BY r1.id;`

const sqlGetChildrenByParentID = `
SELECT r1.id, r1.parent_id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.parent_id = $1
GROUP BY r1.id;`

const sqlGetRegistryByKeyParentID = `
SELECT r1.id, r1.parent_id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.key = $1 AND r1.parent_id = $2
GROUP BY r1.id;`

const sqlGetRegistryRoot = `
SELECT r1.id, r1.key, r1.value, r1.secure, r1.created_at, r1.updated_at,
sum(case when r2.parent_id = r1.id then 1 else 0 end) as children
FROM registry r1 LEFT JOIN registry r2 ON r1.id = r2.parent_id
WHERE r1.key = '{ROOT}' AND r1.parent_id IS NULL
GROUP BY r1.id;`