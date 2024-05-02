package initialsetup

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/config"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func CreateFiles(conn *pgxpool.Pool, projectId string) error {

	count := 0
	query := `SELECT COUNT(*) FROM "menu" WHERE id = '8a6f913a-e3d4-4b73-9fc0-c942f343d0b9'`

	err := conn.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		err = CreateMinioBucket(projectId)
		if err != nil {
			return err
		}
	} else {
		err = CreateMinioBucket(projectId)
		if err != nil {
			return err
		}

		query = `INSERT INTO "menu" 
		(id, icon, label, parent_id, table_id, layout_id, type, bucket_path) 
		VALUES 
		('8a6f913a-e3d4-4b73-9fc0-c942f343d0b9', 'file-pdf.svg', 'Files', 'c57eedc3-a954-4262-a0af-376c65b5a284', '', '', 'FOLDER', $1)`

		_, err = conn.Exec(context.Background(), query, projectId)
		if err != nil {
			return err
		}

		query = `SELECT COUNT(*) FROM "menu" WHERE parent_id = '8a6f913a-e3d4-4b73-9fc0-c942f343d0b9' AND label = 'Media'`

		defaultCount := 0

		err = conn.QueryRow(context.Background(), query).Scan(&defaultCount)
		if err != nil {
			return err
		}

		attributes := map[string]interface{}{
			"label_aa": "Media",
			"label_ak": "Media",
			"path":     "Media",
		}

		attr, err := json.Marshal(attributes)
		if err != nil {
			return err
		}

		if defaultCount > 0 {
			query := `UPDATE "menu" SET attributes = $1 WHERE id = 'f4089a64-4f6f-4604-a57a-b1c99f4d16a8'`

			_, err = conn.Exec(context.Background(), query, attr)
			if err != nil {
				return err
			}
		} else if defaultCount == 0 {
			err = CreateFolderToBucket(projectId, "Media")
			if err != nil {
				return err
			}

			query = `INSERT INTO "menu" 
			(id, icon, label, parent_id, type, attributes) 
			VALUES 
			('f4089a64-4f6f-4604-a57a-b1c99f4d16a8', '', 'Media', '8a6f913a-e3d4-4b73-9fc0-c942f343d0b9', 'MINIO_FOLDER', $1)`

			_, err = conn.Exec(context.Background(), query, attr)
			if err != nil {
				return err
			}
		}
	}

	query = `INSERT INTO "menu" 
		(id, icon, label, parent_id, table_id, layout_id, type) 
		VALUES 
		('9e988322-cffd-484c-9ed6-460d8701551b', 'folder.svg', 'Users', 'c57eedc3-a954-4262-a0af-376c65b5a284', '', '', 'FOLDER')`

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func CreateMinioBucket(bucketName string) error {

	cfg := config.Load()

	fmt.Println("MINIOOO --->>>>>>>>")
	fmt.Println(cfg.MinioAccessKeyID)
	fmt.Println(cfg.MinioSecretKey)

	minioClient, err := minio.New("172.20.20.17:9001", &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	exists, err := minioClient.BucketExists(context.Background(), "")
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("we already own %s", bucketName)
	}

	err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: ""})
	if err != nil {
		return err
	}

	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"AWS": "*",
				},
				"Action": []string{
					"s3:GetBucketLocation",
					"s3:ListBucket",
					"s3:GetObject",
				},
				"Resource": []string{
					"arn:aws:s3:::" + bucketName,
					"arn:aws:s3:::" + bucketName + "/*",
				},
			},
		},
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	err = minioClient.SetBucketPolicy(context.Background(), bucketName, string(policyBytes))
	if err != nil {
		return err
	}

	return nil
}

func CreateFolderToBucket(bucketName, folderName string) error {
	cfg := config.Load()

	minioClient, err := minio.New("172.20.20.17:9001", &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	fullFolderName := folderName + "/"

	_, err = minioClient.StatObject(context.Background(), bucketName, fullFolderName, minio.StatObjectOptions{})
	if err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code == "NoSuchKey" {
			_, err = minioClient.PutObject(context.Background(), bucketName, fullFolderName, strings.NewReader(""), 0, minio.PutObjectOptions{ContentType: "application/octet-stream"})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	return nil
}
