ALTER TABLE contributions
    DROP CONSTRAINT IF EXISTS contributions_edir_member_ethiopian_month_unique;
ALTER TABLE contributions
    DROP CONSTRAINT IF EXISTS contributions_ethiopian_month_check;
ALTER TABLE contributions
    ADD CONSTRAINT contributions_edir_member_month_unique
    UNIQUE (edir_id, member_id, contribution_month);
ALTER TABLE contributions
    ALTER COLUMN ethiopian_year DROP NOT NULL,
    ALTER COLUMN ethiopian_month DROP NOT NULL;
