package initialsetup

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDefaultFieldPermission(conn *pgxpool.Pool, roleId string) error {
	query := `INSERT INTO field_permission (guid, role_id, label, table_slug, field_id)
    VALUES 
    ('5a44c527-dde1-427b-bbfd-574664a65b10', $1, 'Название', 'language_setting', '4946ffb5-9079-4ef8-acdf-0b1d2ffbc36b'),
    ('7c98f01c-5a59-4e8c-bb6d-e376513d78a6', $1, 'Название', 'language_setting', '5842bc93-2943-4cab-b603-c1a8032d7c55'),
    ('9552edd6-bc64-4b6a-ac60-60eb28a48438', $1, 'Название', 'language_setting', '366a5ecd-993f-44b5-9d98-0155f12bb01b'),
    ('d85ac643-99ae-4517-815e-0dbf5aac993d', $1, 'Название', 'timezone_setting', '45cfc41f-ad65-4fde-ad5b-73b8a8d7bd3e'),
    ('af405097-9c3a-4669-9e06-6982314eb007', $1, 'Название', 'timezone_setting', 'da225ef5-7d7c-4e27-a036-6ff895749fa3'),
    ('802b4f04-9d85-45cc-9016-3d8b1ca5bfa2', $1, 'Название', 'project', '37137f5f-ef9b-4710-a6df-fb920750fdfb'),
    ('a2c47645-c470-4add-8cf4-fe771ebb4a10', $1, 'Домен проекта', 'project', 'dfbf6a89-9c78-4922-9a00-0e1555c23ece'),
    ('4fc8f5d3-d7d7-4290-9017-c8d47e6d1644', $1, 'ID', 'project', '8265459c-ab41-45b5-a79d-cbfa299ddaa7'),
    ('f59f17ae-8873-42ca-9854-e2122b7b2ff1', $1, 'Субдомен платформы', 'client_platform', '948500db-538e-412b-ba36-09f5e9f0eccc'),
    ('adc920b0-8501-4186-a094-b2f7117b92d0', $1, 'Название платформы', 'client_platform', 'c818bc89-c2e9-4181-9db4-06fdf837d6e2'),
    ('3ca58e92-5ae0-4ef0-9e45-c072cc228377', $1, 'IT\'S RELATION', 'client_platform', 'd95156ba-d443-4c95-8383-c122747330c5'),
    ('c27601cf-a256-4009-9006-3b20a14c976d', $1, 'ID', 'client_platform', '6c812f3d-1aae-4b9e-8c28-55019ede57f8'),
    ('d553f68f-ebe4-4a0c-96fc-40123ba3a9ce', $1, 'IT\'S RELATION', 'client_platform', 'f7220ec5-d9cb-485b-9652-f3429132375d'),
    ('86ee9f3f-7066-4ab5-903a-57519db4d34f', $1, 'Подтверждено', 'client_type', 'd99ac785-1d1a-49d8-af23-4d92774d15b6'),
    ('861e6969-5881-4438-88ba-4e021b1d12f0', $1, 'ID', 'client_type', '5bcd3857-9f9e-4ab9-97da-52dbdcb3e5d7'),
    ('26294b4f-9559-4d85-82ee-8828d6ac7d37', $1, 'IT\'S RELATION', 'client_type', 'faa90368-d201-4322-82b7-e370f788d248'),
    ('442f5e92-5893-4707-920d-afe571d54d3c', $1, 'Самовосстановление', 'client_type', 'd37e08d6-f7d0-441e-b7af-6034e5c2a77f'),
    ('1f4c2339-2613-42e3-85e9-c438fe796b5f', $1, 'IT\'S RELATION', 'client_type', '4eb81779-7529-420f-991f-a194e2010563'),
    ('1588aeaa-9354-4b0e-91c9-96a479ed2d9b', $1, 'Название типа', 'client_type', '04d0889a-b9ba-4f5c-8473-c8447aab350d'),
    ('39ddc721-2844-4d50-9720-0d4d019b6516', $1, 'Саморегистрация', 'client_type', '763a0625-59d7-4fd1-ad4b-7ef303c3aadf'),
    ('efa21a0f-3bbe-443b-aa57-6f43a5b33239', $1, 'ID', 'role', '3bb6863b-5024-4bfb-9fa0-6ed5bf8d2406'),
    ('7db28cca-d21c-485d-bcb2-f72d40e043c9', $1, 'ID', 'user', 'a0e1ad16-d06d-4a3a-b73b-5a60c43abce1'),
    ('768e8e97-d272-4860-a32e-7f38b925f52b', $1, 'Логин', 'user', '5b7ab2c2-cb07-4fe8-9c19-14d31f1ac11b'),
    ('75085348-7475-41e4-8345-93dbd18a8db2', $1, 'Эл. почта', 'user', 'd826b95a-8c9e-47e1-9a24-3ff1bcb60728'),
    ('9d164b91-1006-47e0-a41b-b4ac6cc2822f', $1, 'Имя', 'user', '8d741e76-4403-4d08-89de-44964e8f282e'),
    ('a61529a0-8e8a-42ad-b28c-1038e1d5a7fc', $1, 'Фамилия', 'user', 'ac1c03d9-dc48-4fbb-8146-1409d4e00eb8')`

	_, err := conn.Exec(context.Background(), query, roleId)
	if err != nil {
		return err
	}

	return nil
}
