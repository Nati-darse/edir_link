-- The Ethiopian date is now computed from the system clock by
-- internal/ethiopiancalendar. The manually maintained date columns are retired.
-- The app_settings table itself is kept for future, unrelated settings.
BEGIN;

ALTER TABLE app_settings
    DROP COLUMN ethiopian_year,
    DROP COLUMN ethiopian_month,
    DROP COLUMN ethiopian_day;

COMMIT;
