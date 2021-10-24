package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

var numberOfAuthorizations int64 = 0
var numberOfVerifications int64 = 0
var sumOfAuthorizationTimes int64 = 0
var sumOfVerificationTimes int64 = 0

func main() {
	app := fiber.New()

	//auth route
	app.Get("/auth/:userName", auth)

	//verify route
	app.Get("/verify", verify)

	//stats route
	app.Get("/stats", stats)

	//readme route
	app.Get("/README.txt", readme)

	app.Listen(":4000")
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func auth(c *fiber.Ctx) error {
	startTime := time.Now()
	privateBytes, err := ioutil.ReadFile("private.pem")
	fatal(err)
	publicBytes, err := ioutil.ReadFile("public.pem")
	fatal(err)
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	fatal(err)

	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = c.Params("userName")
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, err := token.SignedString(privateKey)
	fatal(err)

	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.Value = tokenString
	// Set cookie
	c.Cookie(cookie)

	numberOfAuthorizations++
	sumOfAuthorizationTimes += time.Since(startTime).Milliseconds()

	return c.SendString(string(publicBytes))
}

func verify(c *fiber.Ctx) error {
	startTime := time.Now()
	publicBytes, err := ioutil.ReadFile("public.pem")
	fatal(err)
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	fatal(err)

	tokenString := c.Cookies("token")
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return c.SendStatus(409)
	}

	numberOfVerifications++;
	sumOfVerificationTimes += time.Since(startTime).Milliseconds()

	return c.JSON(fiber.Map{"userName": claims["sub"]})
}

func stats(c *fiber.Ctx) error {
	var numberOfVerificationsString = "0"
	var averageOfVerificationsString = "n/a"
	var numberOfAuthorizationsString = "0"
	var averageOfAuthorizationsString = "n/a"

	if numberOfVerifications != 0 {
		numberOfVerificationsString = strconv.FormatInt(numberOfVerifications, 10)
		averageOfVerificationsString = strconv.FormatInt(sumOfVerificationTimes/numberOfVerifications, 10)+"ms"
	}

	if numberOfAuthorizations != 0 {
		numberOfAuthorizationsString = strconv.FormatInt(numberOfAuthorizations, 10)
		averageOfAuthorizationsString = strconv.FormatInt(sumOfAuthorizationTimes/numberOfAuthorizations, 10)+"ms"
	}

	return c.JSON(fiber.Map{
		"numberOfVerifications": numberOfVerificationsString,
		"averageOfVerifications": averageOfVerificationsString,
		"numberOfAuthorizations": numberOfAuthorizationsString,
		"averageOfAuthorizations": averageOfAuthorizationsString,
	})

}

func readme(c *fiber.Ctx) error {
	readmeBytes, err := ioutil.ReadFile("README.txt")
	fatal(err)

	return c.SendString(string(readmeBytes))

}
