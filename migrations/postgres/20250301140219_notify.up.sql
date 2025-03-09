CREATE TABLE connectors
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE FUNCTION notify_connector_changes() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('connectors_changes', jsonb_build_object(
            'table', TG_TABLE_NAME,
            'action', TG_OP,
            'new_id', NEW.id,
            'old_id', OLD.id
)::text); -- converts to TEXT data type because pg_notify only accept text payload
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER notify_connector_changes
AFTER INSERT OR UPDATE OR DELETE
ON connectors
FOR EACH ROW
EXECUTE FUNCTION notify_connector_changes()

-- channels are created implicitly
-- jsonb_pretty() can convert jsonb object to TEXT too