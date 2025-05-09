ALTER TABLE IF EXISTS "view" 
    ADD COLUMN IF NOT EXISTS "menu_id" UUID REFERENCES "menu"("id") ON DELETE SET NULL,
    DROP COLUMN IF EXISTS "multiple_insert_field",
    DROP COLUMN IF EXISTS "main_field";