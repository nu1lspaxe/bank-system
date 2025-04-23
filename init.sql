CREATE TABLE IF NOT EXISTS "BK_User" (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(20) NOT NULL,
    email VARCHAR(256) NOT NULL UNIQUE,
    password VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE "BK_User" ENABLE ROW LEVEL SECURITY;

CREATE POLICY "BK_User_policy"
ON "BK_User"
FOR SELECT, UPDATE, DELETE
USING (
    id = current_setting('app.current_user_id')::BIGINT
);

CREATE TABLE IF NOT EXISTS "BK_Account" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    id_number VARCHAR(20) NOT NULL UNIQUE,
    currency_code VARCHAR(3) NOT NULL,
    balance NUMERIC(100, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(10) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) 
        REFERENCES "BK_User"(id) ON DELETE CASCADE,
    CONSTRAINT positive_balance 
        CHECK (balance >= 0),
    CONSTRAINT valid_currency_code 
        CHECK (currency_code IN ('USD', 'EUR', 'TWD')),
    CONSTRAINT valid_status 
        CHECK (status IN ('ACTIVE', 'INACTIVE', 'CLOSED', 'FROZEN'))
);

ALTER TABLE "BK_Account" ENABLE ROW LEVEL SECURITY;

CREATE POLICY "BK_Account_policy"
ON "BK_Account"
FOR SELECT, UPDATE, DELETE
USING (
    user_id = current_setting('app.current_user_id')::BIGINT
);

CREATE INDEX idx_bk_account_user_id ON "BK_Account" (user_id);
CREATE INDEX idx_bk_account_id_number ON "BK_Account" (id_number);

CREATE TYPE TX_TYPE AS ENUM (
    'WITHDRAW',
    'DEPOSIT',
    'TRANSFER',
    'INTEREST',
);

CREATE TABLE IF NOT EXISTS "BK_Transaction" (
    id BIGSERIAL PRIMARY KEY,
    account_from BIGINT NOT NULL,  
    account_to BIGINT,     
    amount NUMERIC(20, 2) NOT NULL,
    balance_after NUMERIC(100, 2) NOT NULL,
    tx_type TX_TYPE NOT NULL,
    detail TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (account_from) 
    REFERENCES "BK_Account"(id) ON DELETE SET NULL,
    FOREIGN KEY (account_to)
    REFERENCES "BK_Account"(id) ON DELETE SET NULL
);

ALTER TABLE "BK_Transaction" ENABLE ROW LEVEL SECURITY;

CREATE POLICY "BK_Transaction_policy"
ON "BK_Transaction"
FOR SELECT, UPDATE, DELETE
USING (
    EXISTS (
        SELECT 1 FROM "BK_Account" 
        WHERE id = account_from 
            AND user_id = current_setting('app.current_user_id')::BIGINT
    ) OR EXISTS (
        SELECT 1 FROM "BK_Account" 
        WHERE id = account_to 
            AND user_id = current_setting('app.current_user_id')::BIGINT
    )
);

CREATE INDEX idx_bk_transaction_account_from ON "BK_Transaction" (account_from);
CREATE INDEX idx_bk_transaction_account_to ON "BK_Transaction" (account_to);

CREATE OR REPLACE VIEW v_user_transactions AS
SELECT 
    t.*,
    a.user_id
FROM "BK_Transaction" t
JOIN "BK_Account" a 
    ON t.account_from = a.id OR t.account_to = a.id;

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_bk_user_update 
BEFORE UPDATE ON "BK_User"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trig_bk_account_update
BEFORE UPDATE ON "BK_Account"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- Generate a unique account number
CREATE OR REPLACE FUNCTION generate_account_number()
RETURNS VARCHAR(20) AS $$
DECLARE
    result VARCHAR(20);
BEGIN
    result := (
        SELECT STRING_AGG(FLOOR(RANDOM() * 10)::TEXT, '')
        FROM generate_series(1, 20)
    );

    WHILE EXISTS (
        SELECT 1 FROM "BK_Account" WHERE id_number = result
    ) LOOP 
        result := (
            SELECT STRING_AGG(FLOOR(RANDOM() * 10)::TEXT, '')
            FROM generate_series(1, 20)
        );
    END LOOP;

    RETURN result;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION before_insert_bk_account()
RETURNS TRIGGER AS $$
BEGIN
    NEW.id_number := generate_account_number();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to generate account number before inserting a new account
CREATE TRIGGER trig_bk_account_insert
BEFORE INSERT ON "BK_Account"
FOR EACH ROW
EXECUTE FUNCTION before_insert_bk_account();

-- Withdraw from account
CREATE OR REPLACE FUNCTION withdraw_from_account(
    input_account_id BIGINT, 
    amount NUMERIC(20,2), 
    tx_detail TEXT
) RETURNS TABLE (
    new_balance NUMERIC(100, 2),
    transaction_id BIGINT
) AS $$
DECLARE
    tx_id BIGINT;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM "BK_Account" 
        WHERE id = input_account_id     
            AND status = 'ACTIVE'
    ) THEN
        RAISE EXCEPTION 'Account % not active', input_account_id USING ERRCODE = 'P0001';
    END IF;

    IF amount <= 0 THEN
        RAISE EXCEPTION 'Amount must be positive, got %', amount USING ERRCODE = 'P0001';
    END IF;

    WITH updated AS (
        UPDATE "BK_Account"
        SET balance = balance - amount
        WHERE id = input_account_id 
            AND balance >= amount
        RETURNING balance
    )

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Insufficient funds for account %', account_id USING ERRCODE = 'P0001';
    END IF;

    SELECT balance INTO new_balance FROM updated;

    INSERT INTO "BK_Transaction" (
        account_id, 
        amount, 
        balance_after, 
        tx_type, detail
    ) VALUES (
        input_account_id, 
        amount, 
        new_balance, 
        'WITHDRAW', 
        tx_detail
    ) RETURNING id INTO tx_id;

    RETURN new_balance, tx_id;
END;
$$ LANGUAGE plpgsql;

-- Deposit to account
CREATE OR REPLACE FUNCTION deposit_to_account(
    input_account_id BIGINT, 
    amount NUMERIC(20, 2), 
    tx_detail TEXT
) RETURNS TABLE (
    new_balance NUMERIC(100, 2),
    transaction_id BIGINT
) AS $$
DECLARE
    new_balance NUMERIC(100, 2);
    tx_id BIGINT;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM "BK_Account" 
        WHERE id = input_account_id
            AND status = 'ACTIVE'
    ) THEN
        RAISE EXCEPTION 'Account % not active', input_account_id USING ERRCODE = 'P0001';
    END IF;

    IF amount <= 0 THEN
        RAISE EXCEPTION 'Amount must be positive, got %', amount USING ERRCODE = 'P0001';
    END IF;

    UPDATE "BK_Account"
    SET balance = balance + amount
    WHERE id = account_id
    RETURNING balance INTO new_balance;

    INSERT INTO "BK_Transaction" (
        account_id, 
        amount, 
        balance_after, 
        tx_type, 
        detail
    ) VALUES (
        input_account_id, 
        amount, 
        new_balance, 
        'DEPOSIT', 
        tx_detail
    ) RETURNING id INTO tx_id;

    RETURN new_balance, tx_id;
END;
$$ LANGUAGE plpgsql;

-- Transfer from one account to another
CREATE OR REPLACE FUNCTION transfer_between_accounts(
    from_account_id BIGINT, 
    to_account_id BIGINT, 
    amount NUMERIC(20, 2), 
    tx_detail TEXT
) RETURNS TABLE (
    new_balance_from NUMERIC(100, 2),
    transaction_id BIGINT
) AS $$
DECLARE
    new_balance_from NUMERIC(100, 2);
    tx_id BIGINT;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM "BK_Account" 
        WHERE id IN (from_account_id, to_account_id) 
        AND status = 'ACTIVE'
        HAVING COUNT(DISTINCT id) = 2
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM "BK_Account" 
            WHERE id = from_account_id AND status = 'ACTIVE'
        ) THEN
            RAISE EXCEPTION 'Account % not active', from_account_id USING ERRCODE = 'P0001';
        END IF;
        RAISE EXCEPTION 'Account % not active', to_account_id USING ERRCODE = 'P0001';
    END IF;

    IF from_account_id = to_account_id THEN
        RAISE EXCEPTION 'Cannot transfer to the same account %', from_account_id USING ERRCODE = 'P0001';
    END IF;

    IF amount <= 0 THEN
        RAISE EXCEPTION 'Amount must be positive, got %', amount USING ERRCODE = 'P0001';
    END IF;

    WITH updated_from AS (
        UPDATE "BK_Account"
        SET balance = balance - amount
        WHERE id = from_account_id 
            AND balance >= amount
        RETURNING balance
    ),
    updated_to AS (
        UPDATE "BK_Account"
        SET balance = balance + amount
        WHERE id = to_account_id
        RETURNING balance
    )
    SELECT updated_from.balance INTO new_balance_from FROM updated_from;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Insufficient funds for account %', from_account_id USING ERRCODE = 'P0001';
    END IF;

    INSERT INTO "BK_Transaction" (
        account_from, 
        account_to, 
        amount, 
        balance_after, 
        tx_type, 
        detail
    ) VALUES (
        from_account_id, 
        to_account_id, 
        amount, 
        new_balance_from, 
        'TRANSFER', 
        tx_detail
    ) RETURNING id INTO tx_id;

    RETURN new_balance_from, tx_id;
END;
$$ LANGUAGE plpgsql;