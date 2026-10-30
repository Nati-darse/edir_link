ALTER TABLE contributions
    ALTER COLUMN ethiopian_year SET NOT NULL,
    ALTER COLUMN ethiopian_month SET NOT NULL;

ALTER TABLE contributions
    ADD CONSTRAINT contributions_ethiopian_month_check
    CHECK (ethiopian_month BETWEEN 1 AND 12);

ALTER TABLE contributions
    DROP CONSTRAINT contributions_edir_member_month_unique;

ALTER TABLE contributions
    ADD CONSTRAINT contributions_edir_member_ethiopian_month_unique
    UNIQUE (edir_id, member_id, ethiopian_year, ethiopian_month);
