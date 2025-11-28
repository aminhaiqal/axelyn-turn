CREATE TRIGGER archive_ticket_on_done
AFTER UPDATE ON tickets
FOR EACH ROW
EXECUTE FUNCTION archive_completed_ticket();
