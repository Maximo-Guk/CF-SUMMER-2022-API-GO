package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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

	app.Listen(":80")
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func auth(c *fiber.Ctx) error {
	startTime := time.Now()
	privateBytes, err := os.ReadFile("private.pem")
	fatal(err)
	publicBytes, err := os.ReadFile("public.pem")
	fatal(err)
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	fatal(err)

	// create new JWT RSA256 Token and add claims to it
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = c.Params("userName")
	// expiration date of 1 day
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, err := token.SignedString(privateKey)
	fatal(err)

	// Create jwt cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.HTTPOnly = true
	cookie.Value = tokenString
	// Set cookie
	c.Cookie(cookie)

	numberOfAuthorizations++
	sumOfAuthorizationTimes += time.Since(startTime).Microseconds()

	// return public key
	return c.SendString(string(publicBytes))
}

func verify(c *fiber.Ctx) error {
	startTime := time.Now()
	publicBytes, err := os.ReadFile("public.pem")
	fatal(err)
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	fatal(err)

	tokenString := c.Cookies("token")
	// check if token exists within cookie
	if tokenString == "" {
		c.Status(400)
		return c.SendString("Token not found in cookie")
	}
	claims := jwt.MapClaims{}
	// parse cookie from request and get claims, verify cookie with public key
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	// if verification fails return 401
	if err != nil {
		c.Status(401)
		return c.SendString("Token is invalid")
	}

	numberOfVerifications++
	sumOfVerificationTimes += time.Since(startTime).Microseconds()

	// return userName claim as string
	return c.SendString(claims["sub"].(string))
}

func stats(c *fiber.Ctx) error {
	var numberOfVerificationsString = "0"
	var averageOfVerificationsString = "n/a"
	var numberOfAuthorizationsString = "0"
	var averageOfAuthorizationsString = "n/a"

	// μs are the unit for microseconds
	if numberOfVerifications != 0 {
		numberOfVerificationsString = strconv.FormatInt(numberOfVerifications, 10)
		averageOfVerificationsString = strconv.FormatInt(sumOfVerificationTimes/numberOfVerifications, 10) + "μs"
	}

	if numberOfAuthorizations != 0 {
		numberOfAuthorizationsString = strconv.FormatInt(numberOfAuthorizations, 10)
		averageOfAuthorizationsString = strconv.FormatInt(sumOfAuthorizationTimes/numberOfAuthorizations, 10) + "μs"
	}

	return c.JSON(fiber.Map{
		"numberOfVerifications":   numberOfVerificationsString,
		"averageOfVerifications":  averageOfVerificationsString,
		"numberOfAuthorizations":  numberOfAuthorizationsString,
		"averageOfAuthorizations": averageOfAuthorizationsString,
	})

}

func readme(c *fiber.Ctx) error {
	readmeBytes, err := os.ReadFile("README.txt")
	fatal(err)

	return c.SendString(string(readmeBytes))

}
