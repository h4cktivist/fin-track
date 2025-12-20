CREATE OR REPLACE FUNCTION create_bucket_schemas(
    shard_count INT DEFAULT 2,
    buckets_per_shard INT DEFAULT 2
) RETURNS void AS $$
    DECLARE
    shard_idx INT;
        bucket_idx INT;
        schema_name TEXT;
    BEGIN
    FOR shard_idx IN 0..(shard_count - 1) LOOP
            FOR bucket_idx IN 0..(buckets_per_shard - 1) LOOP
                schema_name := format('bucket_%s_%s', shard_idx, bucket_idx);

    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);

    EXECUTE format('
                    CREATE TABLE IF NOT EXISTS %I.transactions (
                        id SERIAL PRIMARY KEY,
                        user_id INTEGER NOT NULL,
                        amount NUMERIC(14,2) NOT NULL,
                        category TEXT NOT NULL,
                        type TEXT NOT NULL CHECK (type IN (''income'', ''expense'')),
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
                    )
                ', schema_name);

    EXECUTE format('
                    CREATE INDEX IF NOT EXISTS %I_transactions_user_id_idx
                    ON %I.transactions (user_id)
                ', schema_name, schema_name);

    EXECUTE format('
                    CREATE INDEX IF NOT EXISTS %I_transactions_created_at_idx
                    ON %I.transactions (created_at)
                ', schema_name, schema_name);

    RAISE NOTICE 'Created schema %', schema_name;
    END LOOP;
    END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT create_bucket_schemas(2, 2);