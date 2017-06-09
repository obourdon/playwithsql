--
-- CREATE
--
INSERT INTO entityone(action_id, status_id)
    VALUES (1, 1);

INSERT INTO entityone_history(entityone_id, action_id, status_id)
    VALUES (1, 1, 1);

--
-- UPDATE
--
UPDATE entityone
SET
    action_id = 1,
    status_id = 1,
    time_updated= '2017-01-01 13:13:25'
WHERE entityone_id = 1;

INSERT INTO entityone_history(entityone_id, action_id, status_id)
    VALUES (1, 1, 1);

--
-- SELECT
--
SELECT
    e.entityone_id, e.time_created,
    e.action_id, e.status_id, e.time_updated as status_time_created
FROM entityone e
WHERE 0=0
-- filter on PK
AND e.entityone_id IN (1, 2, 3)
-- filter on status
AND e.status_id IN (1, 2, 3)
LIMIT 3