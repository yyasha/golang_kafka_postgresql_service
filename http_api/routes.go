package http_api

import (
	"log"

	"fio_service/graph"
	"fio_service/graph/generated"

	"github.com/arsmn/fastgql/graphql/handler"
	"github.com/arsmn/fastgql/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// setup all routes
func StartHttpServer() {
	app := fiber.New()
	// Allow CORS
	app.Use(cors.New())
	// setup routes
	setupRoutes(app)
	// start listen
	log.Println("HTTP server started")
	err := app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}

// SetupRoutes func for describe group of api routes
func setupRoutes(app *fiber.App) {
	// graphql
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	gqlHandler := srv.Handler()
	playground := playground.Handler("GraphQL playground", "/query")

	app.All("/query", func(c *fiber.Ctx) error {
		gqlHandler(c.Context())
		return nil
	})

	app.All("/", func(c *fiber.Ctx) error {
		playground(c.Context())
		return nil
	})

	log.Printf("connect to http://localhost:3000/ for GraphQL playground")
	// get users
	app.Get("/get_users", getUsers)
	// add user
	app.Post("/add_user", addUser)
	// delete user
	app.Delete("/del_user", delUser)
	// edit user
	app.Post("/edit_user", editUser)
}
