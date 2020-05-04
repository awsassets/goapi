package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"github.com/gofiber/jwt"
	"github.com/gofiber/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goapi/config"
	"goapi/controllers"
	"goapi/repositories"
	"goapi/services"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
)

// todo user email domain verification
// todo reset user password (forgotten)

// todo admin web page

// todo docstring + warning/typo cleaning

var database = mongoDBConnect()

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// Sets MongoDB collections
	userCollection := database.Collection("users")
	houseCollection := database.Collection("houses")
	// Sets repositories
	userRepo := repositories.NewUserRepository(userCollection)
	houseRepo := repositories.NewHouseRepository(houseCollection)
	// Sets services
	userService := services.NewUserService(userRepo)
	houseService := services.NewHouseService(houseRepo, userRepo)
	authService := services.NewAuthService(userRepo)
	// Sets controllers
	userController := controllers.UserController{UserService: userService, AuthService: authService}
	houseController := controllers.HouseController{Service: houseService}
	authController := controllers.AuthController{AuthService: authService, UserService: userService}

	// Set the first groups for routes
	api := app.Group("/v" + strconv.Itoa(config.CurrentAPIVersion))
	//api := app.Group("/v0")
	users := api.Group("/users")
	houses := api.Group("/houses")
	auth := api.Group("/auth")

	// Unauthenticated routes
	users.Post("", userController.Post)       // Register route
	auth.Post("/login", authController.Login) // todo waiting time to prevent brute-force
	auth.Get("/refreshjwt", authController.RefreshJWT)

	// JWT Middleware: Routes declared below will require a valid JWT
	app.Use(jwtware.New(jwtware.Config{
		SigningKey:    []byte(config.HmacSampleSecret),
		SigningMethod: jwt.SigningMethodHS512.Name,
	})) // TODO handler to sent error in json format + TODO handler to check if not disabled user

	// Restricted routes requiring a valid JWT
	users.Get("/:id", userController.GetByID)
	users.Patch("/:id", userController.PatchBy)
	users.Delete("/:id", userController.DeleteBy)

	houses.Post("", houseController.Post)
	houses.Get("/:id", houseController.GetByID)
	houses.Get("/ofUser/:id", houseController.GetByUserID)
	houses.Patch("/:id", houseController.PatchBy)
	houses.Delete("/:id", houseController.DeleteBy)

	// Routes only activated for development
	if config.DevStatus {
		users.Get("", userController.GetAll)
		houses.Get("", houseController.GetAll)
	}

	err := app.Listen(5000)
	if err != nil {
		fmt.Print(err)
	}
}

func mongoDBConnect() *mongo.Database {
	// Load database info from yaml file
	type database struct {
		URI  string `yaml:"DatabaseURI"`
		Name string `yaml:"DatabaseName"`
	}
	var myDatabase database
	yamlFile, err := ioutil.ReadFile("config/database.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &myDatabase)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(myDatabase.URI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB! Database", myDatabase.Name)
	return client.Database(myDatabase.Name)
}
