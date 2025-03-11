CREATE TABLE connectors
(
    id             SERIAL PRIMARY KEY,
    connector_type VARCHAR(255) NOT NULL,
    display_name   VARCHAR(255) NOT NULL,
    enabled        BOOLEAN DEFAULT false,
    last_sync      TIMESTAMP,
    data           TEXT,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()


CREATE FUNCTION notify_connector_changes() RETURNS TRIGGER AS
$$
BEGIN
    PERFORM pg_notify('connectors_changes', jsonb_build_object(
            'table', TG_TABLE_NAME,
            'action', TG_OP,
            'connector_id', CASE WHEN TG_OP = 'DELETE' THEN OLD.id ELSE NEW.id END,
    )::text); -- converts to TEXT data type because pg_notify only accept text payload
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER notify_connector_changes
AFTER INSERT OR UPDATE OR DELETE 
ON connectors
FOR EACH ROW EXECUTE FUNCTION notify_connector_changes()

-- Notes:
-- channels are created implicitly on LISTEN/NOTIFY
-- jsonb_pretty() can convert jsonb object to TEXT too