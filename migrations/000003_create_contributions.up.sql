CREATE TABLE contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    edir_id UUID NOT NULL,
    member_id UUID NOT NULL,

    contribution_month DATE NOT NULL,

    amount NUMERIC(12,2) NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    due_date DATE NOT NULL,

    paid_at TIMESTAMPTZ,

    penalty_amount NUMERIC(12,2) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT contributions_edir_fk
        FOREIGN KEY (edir_id)
        REFERENCES edirs(id)
        ON DELETE CASCADE,

    CONSTRAINT contributions_member_fk
        FOREIGN KEY (member_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT contributions_amount_check
        CHECK (amount >= 0),

    CONSTRAINT contributions_penalty_check
        CHECK (penalty_amount >= 0),

    CONSTRAINT contributions_status_check
        CHECK (
            status IN (
                'PENDING',
                'SUCCESSFUL',
                'FAILED',
                'OVERDUE'
            )
        ),

    CONSTRAINT contributions_edir_member_month_unique
        UNIQUE (edir_id, member_id, contribution_month)
);

CREATE INDEX idx_contributions_edir
    ON contributions(edir_id);

CREATE INDEX idx_contributions_member
    ON contributions(member_id);

CREATE INDEX idx_contributions_status
    ON contributions(status);

CREATE INDEX idx_contributions_month
    ON contributions(contribution_month);
