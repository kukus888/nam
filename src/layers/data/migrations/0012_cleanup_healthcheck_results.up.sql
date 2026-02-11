-- Simple stored procedure to cleanup healthcheck_results table
-- Keeps only status changes (is_successful true→false or false→true) and their previous records
-- Deletes all duplicate/unchanged status records to reduce table size

CREATE OR REPLACE FUNCTION cleanup_healthcheck_results()
RETURNS TABLE (
    records_before BIGINT,
    records_after BIGINT,
    records_deleted BIGINT
) 
LANGUAGE plpgsql AS $$
DECLARE
    rec_before_count BIGINT;
    rec_after_count BIGINT;
    rec_deleted_count BIGINT;
BEGIN
    -- Get initial record count
    SELECT COUNT(*) INTO rec_before_count FROM healthcheck_results;
    
    -- Delete records that are NOT status changes or their previous records
    DELETE FROM healthcheck_results 
    WHERE id NOT IN (
        WITH ordered_results AS (
            SELECT 
                id,
                healthcheck_id,
                application_instance_id,
                is_successful,
                time_start,
                LAG(is_successful) OVER (
                    PARTITION BY healthcheck_id, application_instance_id 
                    ORDER BY time_start
                ) AS prev_is_successful,
                LAG(id) OVER (
                    PARTITION BY healthcheck_id, application_instance_id 
                    ORDER BY time_start
                ) AS prev_id,
                ROW_NUMBER() OVER (
                    PARTITION BY healthcheck_id, application_instance_id 
                    ORDER BY time_start
                ) AS rn
            FROM healthcheck_results
        )
        SELECT DISTINCT record_id 
        FROM (
            -- Keep first record in each group
            SELECT id as record_id 
            FROM ordered_results 
            WHERE rn = 1
            
            UNION
            
            -- Keep records where status changed
            SELECT id as record_id 
            FROM ordered_results 
            WHERE prev_is_successful IS DISTINCT FROM is_successful
            
            UNION
            
            -- Keep the previous record before each status change
            SELECT prev_id as record_id
            FROM ordered_results 
            WHERE prev_is_successful IS DISTINCT FROM is_successful 
              AND prev_id IS NOT NULL
        ) records_to_keep
        WHERE record_id IS NOT NULL
    );
    
    -- Get final record count
    SELECT COUNT(*) INTO rec_after_count FROM healthcheck_results;
    rec_deleted_count := rec_before_count - rec_after_count;
    
    -- Return statistics
    records_before := rec_before_count;
    records_after := rec_after_count;
    records_deleted := rec_deleted_count;
    
    RETURN NEXT;
    
    RAISE NOTICE 'Cleanup completed: % records before, % records after, % deleted', 
        rec_before_count, rec_after_count, rec_deleted_count;
END;
$$;

-- Usage example:
-- SELECT * FROM cleanup_healthcheck_results();