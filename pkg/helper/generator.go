package helper

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateRandomNumber(prefix string, digitNumber int) string {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate a random number with the specified number of digits
	min := int64(1)
	max := int64(10)
	for i := 1; i < digitNumber; i++ {
		min *= 10
		max *= 10
	}
	randomNumber := prefix + "-" + strconv.FormatInt(min+rand.Int63n(max-min), 10)

	return randomNumber
}

func GenerateRandomString(perfix string, length int) string {
	rand.Seed(time.Now().UnixNano())

	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(characters))
		result[i] = characters[randomIndex]
	}
	return perfix + string(result)
}
