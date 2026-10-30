BEGIN;

ALTER TABLE edirs
    RENAME COLUMN due_day TO due_ethiopian_day;

ALTER TABLE edirs
    DROP CONSTRAINT edirs_due_day_check;

ALTER TABLE edirs
    ADD CONSTRAINT edirs_due_ethiopian_day_check
    CHECK (due_ethiopian_day BETWEEN 1 AND 30);

ALTER TABLE contributions
    ADD COLUMN due_ethiopian_day SMALLINT;

UPDATE contributions c
SET due_ethiopian_day = e.due_ethiopian_day
FROM edirs e
WHERE e.id = c.edir_id;

ALTER TABLE contributions
    ALTER COLUMN due_ethiopian_day SET NOT NULL;

ALTER TABLE contributions
    ADD CONSTRAINT contributions_due_ethiopian_day_check
    CHECK (due_ethiopian_day BETWEEN 1 AND 30);

ALTER TABLE contributions
    DROP COLUMN contribution_month,
    DROP COLUMN due_date;

COMMIT;
