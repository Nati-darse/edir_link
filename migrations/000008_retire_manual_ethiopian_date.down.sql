BEGIN;

ALTER TABLE app_settings
    ADD COLUMN ethiopian_year INTEGER NOT NULL DEFAULT 2019
        CHECK (ethiopian_year >= 1),
    ADD COLUMN ethiopian_month SMALLINT NOT NULL DEFAULT 2
        CHECK (ethiopian_month BETWEEN 1 AND 13),
    ADD COLUMN ethiopian_day SMALLINT NOT NULL DEFAULT 1
        CHECK (ethiopian_day BETWEEN 1 AND 30);

ALTER TABLE app_settings
    ADD CONSTRAINT app_settings_pagume_check
    CHECK (ethiopian_month <> 13 OR ethiopian_day <= 6);

COMMIT;
