-- 03_add_archive_function_and_trigger.sql

CREATE OR REPLACE FUNCTION archive_completed_ticket()
RETURNS trigger AS $$
BEGIN
    -- Only archive when status transitions to 'done'
    IF NEW.status = 'done' AND OLD.status IS DISTINCT FROM 'done' THEN
        INSERT INTO ticket_history (
            id, queue_id, customer_name, status, priority,
            estimated_time, version, created_at, updated_at
        )
        VALUES (
            OLD.id, OLD.queue_id, OLD.customer_name, OLD.status, OLD.priority,
            OLD.estimated_time, OLD.version, OLD.created_at, OLD.updated_at
        );

        DELETE FROM tickets WHERE id = OLD.id;

        RETURN NULL; -- Do not update original table
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
