package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {

	connString := "postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable"

	dbpool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}
	defer dbpool.Close()

	query := "SELECT * FROM go_shady WHERE guid = '048f71f6-43c9-4906-a555-a13fecdc68b6'"

	rows, err := dbpool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Error querying data: %v", err)
	}
	defer rows.Close()

	response := []map[string]interface{}{}

	for rows.Next() {

		data := make(map[string]interface{})

		values, err := rows.Values()
		if err != nil {
			log.Fatalf("Error getting row values: %v", err)
		}

		// fmt.Println(values...)

		for i, value := range values {

			if arr, ok := value.([16]uint8); ok {

				value = ConvertGuid(arr)

				fmt.Println(value)
			}
			data[string(rows.FieldDescriptions()[i].Name)] = value
		}

		response = append(response, data)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}

func ConvertGuid(arr [16]uint8) string {
	guidString := fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		arr[0], arr[1], arr[2], arr[3],
		arr[4], arr[5],
		arr[6], arr[7],
		arr[8], arr[9],
		arr[10], arr[11], arr[12], arr[13], arr[14], arr[15])

	return guidString
}
