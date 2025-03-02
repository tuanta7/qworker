CREATE DATABASE IF NOT EXISTS qworker;

CREATE TABLE connectors (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE CHANNEL connectors_changes;

CREATE TRIGGER notify_connector_changes
AFTER INSERT OR UPDATE OR DELETE ON connectors
FOR EACH ROW
EXECUTE PROCEDURE pg_notify('connectors_changes', json_build_object(
  'table', TG_TABLE_NAME,
  'action', TG_OP,
  'id', NEW.id
)::text);