BEGIN;

CREATE TABLE app_settings (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    ethiopian_year INTEGER NOT NULL CHECK (ethiopian_year >= 1),
    ethiopian_month SMALLINT NOT NULL
        CHECK (ethiopian_month BETWEEN 1 AND 13),
    ethiopian_day SMALLINT NOT NULL
        CHECK (ethiopian_day BETWEEN 1 AND 30),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ethiopian_month <> 13 OR ethiopian_day <= 6)
);

INSERT INTO app_settings (
    id,
    ethiopian_year,
    ethiopian_month,
    ethiopian_day
)
VALUES (TRUE, 2019, 2, 1);

ALTER TABLE memberships
    ADD COLUMN joined_ethiopian_year INTEGER,
    ADD COLUMN joined_ethiopian_month SMALLINT,
    ADD COLUMN joined_ethiopian_day SMALLINT;

ALTER TABLE memberships
    ADD CONSTRAINT memberships_joined_ethiopian_year_check
        CHECK (joined_ethiopian_year IS NULL OR joined_ethiopian_year >= 1),
    ADD CONSTRAINT memberships_joined_ethiopian_month_check
        CHECK (
            joined_ethiopian_month IS NULL
            OR joined_ethiopian_month BETWEEN 1 AND 13
        ),
    ADD CONSTRAINT memberships_joined_ethiopian_day_check
        CHECK (
            joined_ethiopian_day IS NULL
            OR joined_ethiopian_day BETWEEN 1 AND 30
        ),
    ADD CONSTRAINT memberships_joined_ethiopian_date_complete_check
        CHECK (
            (
                joined_ethiopian_year IS NULL
                AND joined_ethiopian_month IS NULL
                AND joined_ethiopian_day IS NULL
            )
            OR
            (
                joined_ethiopian_year IS NOT NULL
                AND joined_ethiopian_month IS NOT NULL
                AND joined_ethiopian_day IS NOT NULL
            )
        );

COMMIT;
