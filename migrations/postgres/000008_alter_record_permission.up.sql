<<<<<<< HEAD
ALTER TABLE "record_permission"
ADD COLUMN "search_button" VARCHAR(255) DEFAULT 'Yes'
=======
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'record_permission' 
          AND column_name = 'search_button'
    ) THEN
        ALTER TABLE "record_permission"
        ADD COLUMN "search_button" VARCHAR(255) DEFAULT 'Yes';
    END IF;
END $$;
>>>>>>> a780d567eedd0064f1fd023e07c3d419e3cbbd66
