--ALTER TABLE healthcheck DROP COLUMN ssl_expiry_alert
--SELECT * FROM healthcheck

CREATE OR REPLACE FUNCTION notify_healthcheck_change()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('healthcheck_changes', TG_OP || ':' || COALESCE(NEW.id::text, OLD.id::text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS healthcheck_change_trigger ON healthcheck;
CREATE TRIGGER healthcheck_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON healthcheck
FOR EACH ROW
EXECUTE FUNCTION notify_healthcheck_change();