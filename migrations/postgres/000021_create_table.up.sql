CREATE TABLE IF NOT EXISTS "person" (
    "guid" UUID PRIMARY KEY DEFAULT gen_random_uuid(),            
    "folder_id" UUID,
    "full_name" VARCHAR(100),  
    "email" VARCHAR(255) UNIQUE,         
    "phone_number" VARCHAR(20),          
    "gender" TEXT[],                         
    "image" VARCHAR(255), 
    "date_of_birth" DATE,                
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

INSERT INTO "table" (
    "id", "label", "slug", "description", "deleted_at", "show_in_menu", "is_changed", "icon", 
    "subtitle_field_slug", "folder_id", "is_cached", "soft_delete", "order_by", "is_system", "created_at", "updated_at"
) 
VALUES (
    'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', 'Person', 'person', '', '1970-01-01 18:00:00', true, false, 'person-circle-check.svg', '', NULL, false, false, false, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)ON CONFLICT ("id") DO UPDATE SET 
    "label" = EXCLUDED."label",
    "slug" = EXCLUDED."slug",
    "description" = EXCLUDED."description",
    "deleted_at" = EXCLUDED."deleted_at",
    "show_in_menu" = EXCLUDED."show_in_menu",
    "is_changed" = EXCLUDED."is_changed",
    "icon" = EXCLUDED."icon",
    "subtitle_field_slug" = EXCLUDED."subtitle_field_slug",
    "folder_id" = EXCLUDED."folder_id",
    "is_cached" = EXCLUDED."is_cached",
    "soft_delete" = EXCLUDED."soft_delete",
    "order_by" = EXCLUDED."order_by",
    "is_system" = EXCLUDED."is_system",
    "created_at" = EXCLUDED."created_at",
    "updated_at" = EXCLUDED."updated_at";


INSERT INTO "field" (
    id, table_id, required, slug, label, "default", type, "index", attributes, 
    is_visible, is_system, "unique", automatic, enable_multilanguage, created_at, updated_at
) VALUES
('fb852fe5-0255-4d2e-8ceb-bf331ae55fb2', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'guid', 'ID', 'v4', 'UUID', 'true', '{}'::jsonb, true, true, true, false, false, NOW(), NOW()),
('f54d8076-4972-4067-9a91-c178c02c4273', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'full_name', 'Full Name', '', 'SINGLE_LINE', 'string', '{"fields": {"label_en": {"stringValue": "Full Name", "kind": "stringValue"}, "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"}, "defaultValue": {"stringValue": "", "kind": "stringValue"}, "label": {"stringValue": "", "kind": "stringValue"}}}'::jsonb, false, true, false, false, false, NOW(), NOW()),
('d868638d-35d6-4992-8216-7b2f479f722e', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'email', 'Email', '', 'EMAIL', 'string', '{"fields": {"label_en": {"stringValue": "Email", "kind": "stringValue"}, "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"}, "defaultValue": {"stringValue": "", "kind": "stringValue"}, "label": {"stringValue": "", "kind": "stringValue"}}}'::jsonb, false, true, false, false, false, NOW(), NOW()),
('eb3deeb7-6d34-4e24-b65a-f03e09efd0cf', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'phone_number', 'Phone Number', '', 'INTERNATION_PHONE', 'string', '{"fields": {"label_en": {"stringValue": "Phone Number", "kind": "stringValue"}, "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"}, "defaultValue": {"stringValue": "", "kind": "stringValue"}, "label": {"stringValue": "", "kind": "stringValue"}}}'::jsonb, false, true, false, false, false, NOW(), NOW()),
('b92b9b8c-c138-4ce6-9260-b4452a7f5ae2', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'gender', 'Gender', '', 'MULTISELECT', 'string', '{
    "fields": {
        "is_multiselect": {"boolValue": false, "kind": "boolValue"},
        "label": {"stringValue": "", "kind": "stringValue"},
        "label_en": {"stringValue": "Gender", "kind": "stringValue"},
        "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"},
        "options": {
            "listValue": {
                "values": [
                    {
                        "structValue": {
                            "fields": {
                                "value": {"stringValue": "male", "kind": "stringValue"},
                                "color": {"stringValue": "", "kind": "stringValue"},
                                "icon": {"stringValue": "", "kind": "stringValue"},
                                "id": {"stringValue": "m5nhqhnwx94lr5tvlf", "kind": "stringValue"},
                                "label": {"stringValue": "Male", "kind": "stringValue"}
                            }
                        }
                    },
                    {
                        "structValue": {
                            "fields": {
                                "value": {"stringValue": "female", "kind": "stringValue"},
                                "color": {"stringValue": "", "kind": "stringValue"},
                                "icon": {"stringValue": "", "kind": "stringValue"},
                                "id": {"stringValue": "m5nhqlmt8ivnl7ijvr", "kind": "stringValue"},
                                "label": {"stringValue": "Female", "kind": "stringValue"}
                            }
                        }
                    },
                    {
                        "structValue": {
                            "fields": {
                                "value": {"stringValue": "other", "kind": "stringValue"},
                                "color": {"stringValue": "", "kind": "stringValue"},
                                "icon": {"stringValue": "", "kind": "stringValue"},
                                "id": {"stringValue": "m5nibeydyfhz3y2lfk", "kind": "stringValue"},
                                "label": {"stringValue": "Other", "kind": "stringValue"}
                            }
                        }
                    }
                ]
            }
        }
    }
}'::jsonb, false, true, false, false, false, NOW(), NOW()),
('c5b09b80-528d-4987-9105-a2be539255ee', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'image', 'Image', '', 'PHOTO', 'string', '{"fields": {"defaultValue": {"stringValue": "", "kind": "stringValue"}, "label": {"stringValue": "", "kind": "stringValue"}, "label_en": {"stringValue": "Image", "kind": "stringValue"}, "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"}, "path": {"stringValue": "Media", "kind": "stringValue"}}}'::jsonb, false, true, false, false, false, NOW(), NOW()),
('e5a2a21e-a9e2-4e6d-87e8-57b8dd837d48', 'c1669d87-332c-41ee-84ac-9fb2ac9efdd5', false, 'date_of_birth', 'Date Of Birth', '', 'DATE', 'string', '{"fields": {"defaultValue": {"stringValue": "", "kind": "stringValue"}, "label": {"stringValue": "", "kind": "stringValue"}, "label_en": {"stringValue": "Date of birth", "kind": "stringValue"}, "number_of_rounds": {"nullValue": "NULL_VALUE", "kind": "nullValue"}}}'::jsonb, false, true, false, false, false, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    table_id = EXCLUDED.table_id,
    required = EXCLUDED.required,
    slug = EXCLUDED.slug,
    label = EXCLUDED.label,
    "default" = EXCLUDED."default",
    type = EXCLUDED.type,
    "index" = EXCLUDED."index",
    attributes = EXCLUDED.attributes,
    is_visible = EXCLUDED.is_visible,
    is_system = EXCLUDED.is_system,
    "unique" = EXCLUDED."unique",
    automatic = EXCLUDED.automatic,
    enable_multilanguage = EXCLUDED.enable_multilanguage,
    updated_at = EXCLUDED.updated_at;
