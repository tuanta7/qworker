DROP TRIGGER IF EXISTS notify_connector_changes ON private.connector;
DROP FUNCTION IF EXISTS private.notify_connector_changes;
DROP TABLE IF EXISTS private.connector;
DROP SCHEMA IF EXISTS private;