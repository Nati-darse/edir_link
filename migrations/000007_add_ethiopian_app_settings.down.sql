BEGIN;

ALTER TABLE memberships
    DROP CONSTRAINT IF EXISTS memberships_joined_ethiopian_date_complete_check,
    DROP CONSTRAINT IF EXISTS memberships_joined_ethiopian_day_check,
    DROP CONSTRAINT IF EXISTS memberships_joined_ethiopian_month_check,
    DROP CONSTRAINT IF EXISTS memberships_joined_ethiopian_year_check,
    DROP COLUMN IF EXISTS joined_ethiopian_day,
    DROP COLUMN IF EXISTS joined_ethiopian_month,
    DROP COLUMN IF EXISTS joined_ethiopian_year;

DROP TABLE IF EXISTS app_settings;

COMMIT;
