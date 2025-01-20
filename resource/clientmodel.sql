CREATE TABLE clientmodel (
    deviceid SERIAL PRIMARY KEY,
    groupid INT NOT NULL,
    modelid INT NOT NULL,
    groupname TEXT NOT NULL,
    modelname TEXT NOT NULL,
    createdby TEXT NOT NULL,
    updateat BIGINT NOT NULL
);


-- CREATE OR REPLACE FUNCTION notify_clientmodel_change()
-- RETURNS TRIGGER AS $$
-- BEGIN
   
--     PERFORM pg_notify(
--         'clientmodel_change',
--         json_build_object(
--             'operation', TG_OP,                
--             'deviceid', 
--                 CASE TG_OP 
--                     WHEN 'DELETE' THEN OLD.deviceid 
--                     ELSE NEW.deviceid 
--                 END,                          
--             'updated_at', 
--                 CASE TG_OP 
--                     WHEN 'DELETE' THEN NULL   
--                     ELSE NEW.updateat 
--                 END
--         )::text
--     );
--     RETURN NULL; 
-- END;
-- $$ LANGUAGE plpgsql;

-- CREATE TRIGGER clientmodel_trigger
-- AFTER INSERT OR UPDATE OR DELETE ON clientmodel
-- FOR EACH ROW EXECUTE FUNCTION notify_clientmodel_change();

-- DROP TRIGGER IF EXISTS clientmodel_trigger ON clientmodel;
