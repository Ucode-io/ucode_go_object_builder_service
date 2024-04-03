--app


INSERT INTO "app"("name","description","icon","id", "tables") VALUES ('Авторизация','Авторизация','user-shield.svg','c57eedc3-a954-4262-a0af-376c65b5a283', '[
    {
    "table_id": "373e9aae-315b-456f-8ec3-0851cad46fbf",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "53edfff0-2a31-4c73-b230-06a134afa50b",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "1ab7fadc-1f2b-4934-879d-4e99772526ad",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "2546e042-af2f-4cef-be7c-834e6bde951c",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "0ade55f8-c84d-42b7-867f-6418e1314e28",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "25698624-5491-4c39-99ec-aed2eaf07b97",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "5db33db7-4524-4414-b65a-b6b8e5bba345",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "4c1f5c95-1528-4462-8d8c-cd377c23f7f7",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "961a3201-65a4-452a-a8e1-7c7ba137789c",
    "is_visible": false,
    "is_own_table": true
    },
    {
    "table_id": "5af2bfb2-6880-42ad-80c8-690e24a2523e",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "074fcb3b-038d-483d-b390-ca69490fc4c3",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "08972256-30fb-4d75-b8cf-940d8c4fc8ac",
    "is_visible": true,
    "is_own_table": true
    },
    {
    "table_id": "b1896ed7-ba00-46ae-ae53-b424d2233589",
    "is_visible": true,
    "is_own_table": true
    }
]'::jsonb);



--tables
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Разрешение для связь','view_relation_permission','Разрешение для связь которые в страница сведений',true,false,'door-closed.svg','','074fcb3b-038d-483d-b390-ca69490fc4c3');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Связь','connections','connections',true,false,'brand_connectdevelop.svg','','0ade55f8-c84d-42b7-867f-6418e1314e28');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Роли','role','role',true,false,'brand_critical-role.svg','','1ab7fadc-1f2b-4934-879d-4e99772526ad');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Пользователи','user','user',true,false,'address-card.svg','','2546e042-af2f-4cef-be7c-834e6bde951c');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Разрешение','record_permission','record permission',true,false,'record-vinyl.svg','','25698624-5491-4c39-99ec-aed2eaf07b97');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Проект','project','project',true,false,'diagram-project.svg','','373e9aae-315b-456f-8ec3-0851cad46fbf');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Автоматический фильтр','automatic_filter','Автоматический фильтр для матрица',true,false,'filter.svg','','4c1f5c95-1528-4462-8d8c-cd377c23f7f7');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Клиент платформа','client_platform','client platform',true,false,'brand_bimobject.svg','','53edfff0-2a31-4c73-b230-06a134afa50b');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Разрешение на действие','action_permission','Разрешение на действие',true,false,'eye-dropper.svg','','5af2bfb2-6880-42ad-80c8-690e24a2523e');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Логин таблица','test_login','Test Login',true,false,'blog.svg','','5db33db7-4524-4414-b65a-b6b8e5bba345');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Разрешение поля','field_permission','Разрешение поля',true,false,'clapperboard.svg','','961a3201-65a4-452a-a8e1-7c7ba137789c');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Тип клиентов','client_type','client type',true,false,'angles-right.svg','','ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Шаблон','template','Шаблоны',true,false,'arrow-right-to-bracket.svg','','08972256-30fb-4d75-b8cf-940d8c4fc8ac');
INSERT INTO "table"("label","slug","description","show_in_menu","is_changed","icon","subtitle_field_slug","id") VALUES ('Файл','file','Файлы',true,false,'file-arrow-down.svg','','b1896ed7-ba00-46ae-ae53-b424d2233589');

--fields




INSERT INTO "field" ("id", "required", "slug", "label", "default", "type", "index", "is_visible", "table_id", "relation_id","commit_id", "attributes") 
    VALUES ('d8127cf2-2d60-474e-94ba-317d3b1ba18a',false,'table_slug','Название таблица','','SINGLE_LINE','string',false,'074fcb3b-038d-483d-b390-ca69490fc4c3', '', '', '{
        "fields": {
            "disabled": {
                "boolValue": false,
                "kind": "boolValue"
            },
            "icon": {
                "stringValue": "",
                "kind": "stringValue"
            },
            "placeholder": {
                "stringValue": "",
                "kind": "stringValue"
            },
            "showTooltip": {
                "boolValue": false,
                "kind": "boolValue"
            },
            "creatable": {
                "boolValue": false,
                "kind": "boolValue"
            },
            "defaultValue": {
                "stringValue": "",
                "kind": "stringValue"
            }
        }}'::jsonb);


INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('076c519a-5503-4bff-99f1-c741ed7d47b8',false,'relation_id','Ид свяьза','','SINGLE_LINE','string',false,'074fcb3b-038d-483d-b390-ca69490fc4c3','','', '{
            "fields": {
                "defaultValue": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "disabled": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "icon": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "creatable": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('c5962e1c-2687-46a5-b2dd-d46d41a038c1',false,'view_permission','Разрешение на просмотр','','SWITCH','string',false,'074fcb3b-038d-483d-b390-ca69490fc4c3','','', '{
            "fields": {
                "defaultValue": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "disabled": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "icon": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "creatable": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") 
    VALUES ('a73fd453-3c21-4ab8-9e21-59d85acd106c',false,'guid','ID','v4','UUID', '',true,'074fcb3b-038d-483d-b390-ca69490fc4c3','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('6f344830-819c-40a3-a255-f11cdb515c2e',false,'role_id','FROM view_relation_permission TO role','','LOOKUP','',true,'074fcb3b-038d-483d-b390-ca69490fc4c3','158213ef-f38d-4c0d-b9ec-815e4d27db7e','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('9d53673d-4df3-4679-91be-8a787bdff724',false,'table_slug','Table Slug','','SINGLE_LINE','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('546320ae-8d9f-43cb-afde-3df5701e4b49',false,'icon','Icon','','ICON','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('a9767595-8863-414e-9220-f6499def0276',false,'view_slug','View Slug','','SINGLE_LINE','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('b73c268c-9b91-47e4-9cb8-4f1d4ad14605',false,'view_label','View label','','SINGLE_LINE','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('c71928df-22d1-408c-8d63-7784cbec4a1d',false,'name','Название','','SINGLE_LINE','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('71a33f28-002e-42a9-95fe-934a1f04b789',false,'type','Тип','','MULTISELECT','string',false,'0ade55f8-c84d-42b7-867f-6418e1314e28','','', '{
            "fields": {
                "options": {
                    "listValue": {
                        "values": []
                    },
                    "kind": "listValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                },
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('f6962da4-bc72-4236-ac7b-18589c2fc029',false,'client_type_id','IT''S RELATION','','LOOKUP','', true,'0ade55f8-c84d-42b7-867f-6418e1314e28','65a2d42f-5479-422f-84db-1a98547dfa04','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('fbcf919a-25b8-486b-a110-342c8c47385f',false,'guid','ID','v4','UUID','',true,'0ade55f8-c84d-42b7-867f-6418e1314e28','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('3bb6863b-5024-4bfb-9fa0-6ed5bf8d2406',false,'guid','ID','v4','UUID','',true,'1ab7fadc-1f2b-4934-879d-4e99772526ad','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('c12adfef-2991-4c6a-9dff-b4ab8810f0df',false,'name','Название роли','','SINGLE_LINE','string',false,'1ab7fadc-1f2b-4934-879d-4e99772526ad','','', '{
            "fields": {
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('cb677e25-ddb3-4a64-a0cd-5aa6653417ed',false,'client_platform_id','IT''S RELATION','','LOOKUP','',true,'1ab7fadc-1f2b-4934-879d-4e99772526ad','ca008469-cfe2-4227-86db-efdf69680310','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('123cd75b-2da5-458f-8020-8176a18b54ce',false,'project_id','IT''S RELATION','','LOOKUP','',true,'1ab7fadc-1f2b-4934-879d-4e99772526ad','a56d0ad8-4645-42d8-9fbb-77e22526bd17','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('645d85b8-67e0-4594-96c7-540b63d6b018',false,'active','Актив','','NUMBER','string',false,'2546e042-af2f-4cef-be7c-834e6bde951c','','', '{
            "fields": {
                "maxLength": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "placeholder": {
                    "stringValue": "",
                    "kind": "stringValue"
                },
                "showTooltip": {
                    "boolValue": false,
                    "kind": "boolValue"
                }
            }
        }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('110055ac-75ab-4c1f-ae35-67098d1816a5',false,'client_type_id','IT''S RELATION','','LOOKUP','',true,'1ab7fadc-1f2b-4934-879d-4e99772526ad','8ab28259-800d-4079-8572-a0f033d70e35','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('2a77fd05-5278-4188-ba34-a7b8d13e2e51',false,'guid','ID','v4','UUID','', true,'2546e042-af2f-4cef-be7c-834e6bde951c','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('87ddadf0-689b-4285-9fc7-5cb76bdd4a7c',false,'expires_at','Срок годности','','DATE_TIME','string',false,'2546e042-af2f-4cef-be7c-834e6bde951c','', '','{
                "fields": {
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d84b1431-f7ae-43c5-b2e1-83f82ec1f979',false,'client_platform_id','IT''S RELATION','','LOOKUP','',true,'2546e042-af2f-4cef-be7c-834e6bde951c','e03071ed-a3e1-417d-a654-c0998a7c74bc','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('be11f4ac-1f91-4e04-872d-31cef96954e9',false,'project_id','IT''S RELATION','','LOOKUP', '',true,'2546e042-af2f-4cef-be7c-834e6bde951c','6d2f94cb-0de4-455e-8dfc-97800eac7579','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('bd5f353e-52d6-4b07-946c-678534a624ec',false,'role_id','IT''S RELATION','','LOOKUP','',true,'2546e042-af2f-4cef-be7c-834e6bde951c','63b54109-5476-43c1-bf26-24e2266a33f0','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5ca9db39-f165-4877-a191-57b5e8fedaf5',false,'client_type_id','IT''S RELATION','','LOOKUP','',true,'2546e042-af2f-4cef-be7c-834e6bde951c','8f123dec-dfe4-4b89-956c-f607c84a84bd','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('9bdbb8eb-334b-4515-87dc-20abd0da129a',false,'table_slug','Название таблица','','SINGLE_LINE','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('f932bf71-9049-462b-9342-8347bca7598d',false,'update','Изменение','','SINGLE_LINE','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('4d1bb99b-d2f0-4ac7-8c50-2f5dd7932755',false,'guid','ID','v4','UUID','',true,'25698624-5491-4c39-99ec-aed2eaf07b97','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('1e71486b-1438-4170-8883-50b505b8bb14',false,'write','Написать','','SINGLE_LINE','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d95c1242-63ab-45c1-bd23-86f23ee72971',false,'delete','Удаление','','SINGLE_LINE','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('27355d70-1c11-4fb9-9192-a1fffd93d9db',false,'read','Чтение','','SINGLE_LINE','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5f099f9f-8217-4790-a8ee-954ec165b8d8',false,'is_have_condition','Есть условия','','SWITCH','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "creatable": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('8e748044-1b00-485c-b445-63e44380a6b1',false,'role_id','IT''S RELATION','','LOOKUP','',true,'25698624-5491-4c39-99ec-aed2eaf07b97','82e93baf-2e02-432a-942b-2c93cbe26b89','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('dfbf6a89-9c78-4922-9a00-0e1555c23ece',false,'domain','Домен проекта','','SINGLE_LINE','string',false,'373e9aae-315b-456f-8ec3-0851cad46fbf','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('37137f5f-ef9b-4710-a6df-fb920750fdfb',false,'name','Название','','SINGLE_LINE','string',false,'373e9aae-315b-456f-8ec3-0851cad46fbf','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('8368fc76-0e80-409c-b64e-2275304411d8',false,'table_slug','Название таблица','','SINGLE_LINE','string',false,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('8265459c-ab41-45b5-a79d-cbfa299ddaa7',false,'guid','ID','v4','UUID','',true,'373e9aae-315b-456f-8ec3-0851cad46fbf','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('957ffe32-714a-41d2-9bd8-e6b6b30fef67',false,'object_field','Полья объекты','','SINGLE_LINE','string',false,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('6d5d18cd-255d-49fd-a08e-5a6b0f1b093f',false,'custom_field','Пользавательские полья','','SINGLE_LINE','string',false,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('a1ece772-a8e0-41ae-8060-e1f667d0d96e',false,'role_id','IT''S RELATION','','LOOKUP','',true,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','697fbd16-97d8-4233-ab21-4ce12dd6c5c6','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('2ca6eec7-faea-4afd-a75f-980c18164f3c',false,'guid','ID','v4','UUID','',true,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('948500db-538e-412b-ba36-09f5e9f0eccc',false,'subdomain','Субдомен платформы','','SINGLE_LINE','string',false,'53edfff0-2a31-4c73-b230-06a134afa50b','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('c818bc89-c2e9-4181-9db4-06fdf837d6e2',false,'name','Название платформы','','SINGLE_LINE','string',false,'53edfff0-2a31-4c73-b230-06a134afa50b','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('f7220ec5-d9cb-485b-9652-f3429132375d',false,'project_id','IT''S RELATION','','LOOKUP','',true,'53edfff0-2a31-4c73-b230-06a134afa50b','c1492b03-8e76-4a09-9961-f61d413dbe68','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('6c812f3d-1aae-4b9e-8c28-55019ede57f8',false,'guid','ID','v4','UUID','',true,'53edfff0-2a31-4c73-b230-06a134afa50b','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('1e39a65d-9709-4c5a-99e4-dde67191d95a',false,'custom_event_id','Ид действия','','SINGLE_LINE','string',false,'5af2bfb2-6880-42ad-80c8-690e24a2523e','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d95156ba-d443-4c95-8383-c122747330c5',false,'client_type_ids','IT''S RELATION','','LOOKUPS','', true,'53edfff0-2a31-4c73-b230-06a134afa50b','426a0cd6-958d-4317-bf23-3b4ea4720e53','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('b84f052c-c407-48c5-a4bf-6bd54869fbd7',false,'permission','Разрешение','','SWITCH','string',false,'5af2bfb2-6880-42ad-80c8-690e24a2523e','','', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "creatable": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('34abee63-37ad-48c1-95d2-f4a032c373a1',false,'table_slug','Название таблица','','SINGLE_LINE','string',false,'5af2bfb2-6880-42ad-80c8-690e24a2523e','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('f58a6d67-7254-474c-af2a-052596bb0513',false,'role_id','FROM action_permission TO role','','LOOKUP','',true,'5af2bfb2-6880-42ad-80c8-690e24a2523e','d522a2ac-7fb4-413d-b5bb-8d1d34b65b98','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('cf504dda-a34a-4422-843b-afaa32efbe49',false,'guid','ID','v4','UUID','',true,'5af2bfb2-6880-42ad-80c8-690e24a2523e','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('f5486957-e804-4050-a3c5-cfdcdaca0a16',false,'table_slug','Table Slug','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d02ddb83-ad98-47f5-bc0a-6ee7586d9a8e',false,'login_strategy','Login strategy','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5591515f-8212-4bd5-b13b-fffd9751e9ce',false,'login_label','Login label','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('fbc9b3e9-0507-48f5-9772-d42febfb4d30',false,'login_view','Login view','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('7ab42774-6ca9-4e10-a71b-77871009e0a2',false,'object_id','Ид обьеткта','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('35ddf13d-d724-42ab-a1bb-f3961c7db9d6',false,'password_view','Password view','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('14e3ed9a-384e-45ac-897f-0fb4174adfaf',false,'guid','ID','v4','UUID','',true,'5db33db7-4524-4414-b65a-b6b8e5bba345','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('276c0e0c-2b79-472a-bcdf-ac0eed115ebe',false,'password_label','Password label','','SINGLE_LINE','string',false,'5db33db7-4524-4414-b65a-b6b8e5bba345','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('7587ed1d-a8b9-4aa8-b845-56dbb9333e25',false,'field_id','Ид поля','','SINGLE_LINE','string',false,'961a3201-65a4-452a-a8e1-7c7ba137789c','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d5fda673-95b2-492a-97c2-afd0466f1e32',false,'client_type_id','IT''S RELATION','','LOOKUP','',true,'5db33db7-4524-4414-b65a-b6b8e5bba345','79bdd075-eef0-48d1-b763-db8dfd819043','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('00787831-04b4-4a08-b74d-14f80a219b86',false,'view_permission','Разрешение на просмотр','','SWITCH','string',false,'961a3201-65a4-452a-a8e1-7c7ba137789c','','', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "creatable": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('27634c7a-1de8-4d54-9f57-5ff7c77d0af9',false,'table_slug','Название таблица','','SINGLE_LINE','string',false,'961a3201-65a4-452a-a8e1-7c7ba137789c','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('6040f51f-7b41-4e7a-87b1-b48286a00bea',false,'guid','ID','v4','UUID','',true,'961a3201-65a4-452a-a8e1-7c7ba137789c','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('9ae553a2-edca-41f7-ba8a-557dc24cb74a',false,'edit_permission','Изменить разрешение','','SWITCH','string',false,'961a3201-65a4-452a-a8e1-7c7ba137789c','','', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "creatable": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('285ceb40-6267-4f5e-9327-f75fe79e8bfe',false,'label','Название поля','','SINGLE_LINE','string',false,'961a3201-65a4-452a-a8e1-7c7ba137789c','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('1267fb0d-2788-4171-ab69-b9573d3974a2',false,'role_id','IT''S RELATION','','LOOKUP','',true,'961a3201-65a4-452a-a8e1-7c7ba137789c','8283449e-7978-4e75-83d6-1b6f3a194683','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d99ac785-1d1a-49d8-af23-4d92774d15b6',false,'confirm_by','Подтверждено','','MULTISELECT','string',false,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','','', '{
                "fields": {
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "options": {
                        "listValue": {
                            "values": [
                                {
                                    "stringValue": "UNDECIDED",
                                    "kind": "stringValue"
                                },
                                {
                                    "stringValue": "PHONE",
                                    "kind": "stringValue"
                                },
                                {
                                    "stringValue": "EMAIL",
                                    "kind": "stringValue"
                                }
                            ]
                        },
                        "kind": "listValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('04d0889a-b9ba-4f5c-8473-c8447aab350d',false,'name','Название типа','','SINGLE_LINE','string',false,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5bcd3857-9f9e-4ab9-97da-52dbdcb3e5d7',false,'guid','ID','v4','UUID','',true,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('763a0625-59d7-4fd1-ad4b-7ef303c3aadf',false,'self_register','Саморегистрация','','SWITCH','string',false,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('faa90368-d201-4322-82b7-e370f788d248',false,'project_id','IT''S RELATION','','LOOKUP','',true,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','094c33df-5556-45f3-a74c-7f589412bcc8','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('d37e08d6-f7d0-441e-b7af-6034e5c2a77f',false,'self_recover','Самовосстановление','','SWITCH','string',false,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','','', '{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('834df8ef-edb7-4170-996c-9bd5431d9a62',false,'table_slug','Таблица','','SINGLE_LINE','string',false,'08972256-30fb-4d75-b8cf-940d8c4fc8ac','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('4eb81779-7529-420f-991f-a194e2010563',false,'client_platform_ids','IT''S RELATION','','LOOKUPS','',true,'ed3bf0d9-40a3-4b79-beb4-52506aa0b5ea','426a0cd6-958d-4317-bf23-3b4ea4720e53','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('9772b679-33ec-4004-b527-317a1165575e',false,'title','Название','','SINGLE_LINE','string',false,'08972256-30fb-4d75-b8cf-940d8c4fc8ac','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5dda58a1-84ac-4c50-8993-02e2cefcb29a',false,'size','Размер','','MULTISELECT','string',false,'08972256-30fb-4d75-b8cf-940d8c4fc8ac','','', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "options": {
                        "listValue": {
                            "values": [
                                {
                                    "structValue": {
                                        "fields": {
                                            "color": {
                                                "stringValue": "",
                                                "kind": "stringValue"
                                            },
                                            "icon": {
                                                "stringValue": "",
                                                "kind": "stringValue"
                                            },
                                            "id": {
                                                "stringValue":
                                                    "l8pqwm99uvafrgkle38",
                                                "kind": "stringValue"
                                            },
                                            "label": {
                                                "stringValue": "A4",
                                                "kind": "stringValue"
                                            },
                                            "value": {
                                                "stringValue": "A4",
                                                "kind": "stringValue"
                                            }
                                        }
                                    },
                                    "kind": "structValue"
                                },
                                {
                                    "structValue": {
                                        "fields": {
                                            "color": {
                                                "stringValue": "",
                                                "kind": "stringValue"
                                            },
                                            "icon": {
                                                "stringValue": "",
                                                "kind": "stringValue"
                                            },
                                            "id": {
                                                "stringValue":
                                                    "l8pqwp8kktxrtw6xebi",
                                                "kind": "stringValue"
                                            },
                                            "label": {
                                                "stringValue": "A5",
                                                "kind": "stringValue"
                                            },
                                            "value": {
                                                "stringValue": "A5",
                                                "kind": "stringValue"
                                            }
                                        }
                                    },
                                    "kind": "structValue"
                                }
                            ]
                        },
                        "kind": "listValue"
                    },
                    "placeholder": {
                        "stringValue": "size",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "has_color": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "has_icon": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "is_multiselect": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('494e1ad3-fce8-4e6c-921f-850d0ec73cc4',false,'guid','ID','v4','UUID','',true,'08972256-30fb-4d75-b8cf-940d8c4fc8ac','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('98279b02-10c0-409e-8303-14224fd76ec6',false,'html','HTML','','SINGLE_LINE','string',false,'08972256-30fb-4d75-b8cf-940d8c4fc8ac','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('a99106a9-32dc-446b-9850-8713d687804a',false,'date','Время создание','','DATE_TIME','string',false,'b1896ed7-ba00-46ae-ae53-b424d2233589','','', '{
                "fields": {
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('61870278-3195-4874-9c0c-28104bbfb19a',false,'type','Тип файла','','SINGLE_LINE','string',false,'b1896ed7-ba00-46ae-ae53-b424d2233589','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('b7e00be4-70f2-4a57-bf77-91d3834d0520',false,'size','Размер','','NUMBER','string',false,'b1896ed7-ba00-46ae-ae53-b424d2233589','','', '{
                "fields": {
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('afb99f72-106d-4939-b831-9e4b025afb9f',false,'name','Название','','SINGLE_LINE','string',false,'b1896ed7-ba00-46ae-ae53-b424d2233589','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('a9199d67-72bc-42b0-bcd3-04d8a18b9441',false,'guid','ID','v4','UUID','',true,'b1896ed7-ba00-46ae-ae53-b424d2233589','','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('5c0efd79-60f4-455b-b1df-e51e3b56b140',false,'file_link','Линк фaйла','','SINGLE_LINE','string',false,'b1896ed7-ba00-46ae-ae53-b424d2233589','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
            --checked all of up
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('4be7708d-e8af-4687-ad3d-8df0f7be1566',false,'cashbox_id','FROM file TO cash_transaction','','LOOKUP','string',true,'b1896ed7-ba00-46ae-ae53-b424d2233589','dae09c03-247a-4353-8f17-fc35e545a44e','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('bbe64ec5-9b8b-4b8a-886f-fa96342f29ab',false,'news_id','FROM file TO news','','LOOKUP','',true,'b1896ed7-ba00-46ae-ae53-b424d2233589','89afc0b2-431b-4243-a22f-53539f50deff','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('673a828f-3955-40f7-8968-dfee6eb6442d',false,'specialities_id','FROM file TO specialities','','LOOKUP','',true,'b1896ed7-ba00-46ae-ae53-b424d2233589','df8b1c3b-7f1a-43f4-8368-b37f191c888d','','{}');
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('7ae7cc2b-d0eb-4e76-9a9f-8b72c8dc9a71',false,'is_public','Общедоступ','','SWITCH','string',false,'25698624-5491-4c39-99ec-aed2eaf07b97','','5ed7c5af-9433-441d-ad45-5fc6bdf2bbd9', '{
                "fields": {
                    "icon": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "creatable": {
                        "boolValue": false,
                        "kind": "boolValue"
                    },
                    "defaultValue": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "disabled": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);
INSERT INTO "field"("id","required","slug","label","default","type","index","is_visible","table_id","relation_id","commit_id", "attributes") VALUES ('eb0c1ac2-eff3-4ca2-b383-9c801a2992ff',false,'method','Метод','','SINGLE_LINE','string',false,'4c1f5c95-1528-4462-8d8c-cd377c23f7f7','','', '{
                "fields": {
                    "maxLength": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "placeholder": {
                        "stringValue": "",
                        "kind": "stringValue"
                    },
                    "showTooltip": {
                        "boolValue": false,
                        "kind": "boolValue"
                    }
                }
            }'::jsonb);

--relations

INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('426a0cd6-958d-4317-bf23-3b4ea4720e53','client_type','client_platform_ids','client_platform','client_type_ids','Many2Many','{c818bc89-c2e9-4181-9db4-06fdf837d6e2}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('ca008469-cfe2-4227-86db-efdf69680310','role','client_platform_id','client_platform','id','Many2One','{c818bc89-c2e9-4181-9db4-06fdf837d6e2}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('e03071ed-a3e1-417d-a654-c0998a7c74bc','user','client_platform_id','client_platform','id','Many2One','{c818bc89-c2e9-4181-9db4-06fdf837d6e2}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('8ab28259-800d-4079-8572-a0f033d70e35','role','client_type_id','client_type','id','Many2One','{04d0889a-b9ba-4f5c-8473-c8447aab350d}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('8f123dec-dfe4-4b89-956c-f607c84a84bd','user','client_type_id','client_type','id','Many2One','{04d0889a-b9ba-4f5c-8473-c8447aab350d}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('65a2d42f-5479-422f-84db-1a98547dfa04','connections','client_type_id','client_type','id','Many2One','{04d0889a-b9ba-4f5c-8473-c8447aab350d}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('79bdd075-eef0-48d1-b763-db8dfd819043','test_login','client_type_id','client_type','id','Many2One','{04d0889a-b9ba-4f5c-8473-c8447aab350d}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('63b54109-5476-43c1-bf26-24e2266a33f0','user','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('82e93baf-2e02-432a-942b-2c93cbe26b89','record_permission','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('697fbd16-97d8-4233-ab21-4ce12dd6c5c6','automatic_filter','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('8283449e-7978-4e75-83d6-1b6f3a194683','field_permission','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('d522a2ac-7fb4-413d-b5bb-8d1d34b65b98','action_permission','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('158213ef-f38d-4c0d-b9ec-815e4d27db7e','view_relation_permission','role_id','role','id','Many2One','{c12adfef-2991-4c6a-9dff-b4ab8810f0df}','',false,'','',false);
INSERT INTO "relation"("id","table_from","field_from","table_to","field_to","type","view_fields","relation_field_slug","is_user_id_default","cascading_tree_field_slug","cascading_tree_table_slug","object_id_from_jwt") VALUES ('0ab05a13-d077-4086-9b7e-a4029d451acd','cashbox','user_id','user','id','Many2One','{22144ff4-7c1c-4102-9697-80f3ccaf3941}','',false,'','',false);
