-- Need to repair hastily created triggers due to typo in channel name
-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS healthcheck_change_trigger ON healthcheck;
DROP TRIGGER IF EXISTS healthcheck_change_trigger ON application_instance;

-- Recreate the function with the correct channel name
CREATE OR REPLACE FUNCTION notify_healthcheck_change()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('healthcheck_change', TG_OP || ':' || COALESCE(NEW.id::text, OLD.id::text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate the triggers with the correct function
CREATE TRIGGER healthcheck_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON healthcheck
FOR EACH ROW
EXECUTE FUNCTION notify_healthcheck_change();

-- Create a separate function for application_instance changes
CREATE OR REPLACE FUNCTION notify_application_instance_change()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('application_instance_change', TG_OP || ':' || COALESCE(NEW.id::text, OLD.id::text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger for application_instance changes
CREATE TRIGGER application_instance_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON application_instance
FOR EACH ROW
EXECUTE FUNCTION notify_application_instance_change();

-- Create the trigger for server changes
-- If server changes, we need to notify all related application instances
CREATE OR REPLACE FUNCTION notify_instances_on_server_change()
RETURNS trigger AS $$
DECLARE
  app_instance RECORD;
BEGIN
  -- Notify all related application instances
  FOR app_instance IN SELECT * FROM application_instance WHERE server_id = COALESCE(NEW.id, OLD.id)
  LOOP
    PERFORM pg_notify('application_instance_change', TG_OP || ':' || app_instance.id::text);
  END LOOP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger for server changes
CREATE TRIGGER server_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON server
FOR EACH ROW
EXECUTE FUNCTION notify_instances_on_server_change();

-- Create the trigger for application_definition changes
-- If application_definition changes, we need to notify all related application instances
CREATE OR REPLACE FUNCTION notify_instances_on_application_definition_change()
RETURNS trigger AS $$
DECLARE
  app_instance RECORD;
BEGIN
    -- Notify all related application instances
    FOR app_instance IN SELECT * FROM application_instance WHERE application_definition_id = COALESCE(NEW.id, OLD.id)
    LOOP
        PERFORM pg_notify('application_instance_change', TG_OP || ':' || app_instance.id::text);
    END LOOP;
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

-- Create the trigger for application_definition changes
CREATE TRIGGER application_definition_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON application_definition
FOR EACH ROW
EXECUTE FUNCTION notify_instances_on_application_definition_change();
