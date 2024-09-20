-- Проверка существования базы данных и создание её, если она не существует
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'orderdb') THEN
        PERFORM dblink_exec('dbname=postgres', 'CREATE DATABASE orderdb');
    END IF;
END $$;
