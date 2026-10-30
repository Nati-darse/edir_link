ALTER TABLE contributions
    ADD COLUMN ethiopian_year INTEGER,
    ADD COLUMN ethiopian_month SMALLINT;

-- Existing rows will be backfilled before these columns become mandatory.
