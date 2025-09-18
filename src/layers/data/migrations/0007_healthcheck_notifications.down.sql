-- Drop triggers
DROP TRIGGER IF EXISTS healthcheck_change_trigger ON healthcheck;
DROP TRIGGER IF EXISTS application_instance_change_trigger ON application_instance;
DROP TRIGGER IF EXISTS server_change_trigger ON server;
DROP TRIGGER IF EXISTS application_definition_change_trigger ON application_definition;

-- Drop functions
DROP FUNCTION IF EXISTS notify_healthcheck_change();
DROP FUNCTION IF EXISTS notify_application_instance_change();
DROP FUNCTION IF EXISTS notify_instances_on_server_change();
DROP FUNCTION IF EXISTS notify_instances_on_application_definition_change();

-- Recreate the original function, as in 0001_base.up.sql
CREATE OR REPLACE FUNCTION notify_healthcheck_change()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('healthcheck_changes', TG_OP || ':' || COALESCE(NEW.id::text, OLD.id::text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER healthcheck_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON healthcheck
FOR EACH ROW
EXECUTE FUNCTION notify_healthcheck_change();

CREATE TRIGGER healthcheck_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON application_instance
FOR EACH ROW
EXECUTE FUNCTION notify_healthcheck_change();