CREATE SCHEMA IF NOT EXISTS private

CREATE TABLE IF NOT EXISTS private.connector
(
    id             SERIAL PRIMARY KEY,
    connector_type VARCHAR(255) NOT NULL,
    display_name   VARCHAR(255) NOT NULL,
    enabled        BOOLEAN               DEFAULT false,
    last_sync      TIMESTAMP,
    data           TEXT,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE FUNCTION private.notify_connector_changes() RETURNS TRIGGER AS
$$
BEGIN
    IF TG_OP = 'UPDATE' AND OLD.enabled = NEW.enabled AND OLD.data = NEW.data
    THEN RETURN NEW; -- Do nothing if only ignored fields are updated
    END IF;

    PERFORM pg_notify('connectors_changes', jsonb_build_object(
            'table', TG_TABLE_NAME,
            'action', TG_OP,
            'id', CASE WHEN TG_OP = 'DELETE' THEN OLD.id ELSE NEW.id END
    )::text); -- converts to TEXT data type because pg_notify only accept text payload
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER notify_connector_changes
    AFTER INSERT OR UPDATE OR DELETE
    ON private.connector
    FOR EACH ROW
EXECUTE FUNCTION private.notify_connector_changes()

-- Notes:
-- channels are created implicitly on LISTEN/NOTIFY
-- jsonb_pretty() can convert jsonb object to TEXT too