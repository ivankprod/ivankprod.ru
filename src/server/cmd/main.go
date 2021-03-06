package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	RequestLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
	"github.com/gofiber/template/handlebars"
	"github.com/jmoiron/sqlx"
	"github.com/tarantool/go-tarantool"

	"github.com/ivankprod/ivankprod.ru/src/server/internal/router"
	"github.com/ivankprod/ivankprod.ru/src/server/pkg/db"
	BaseLogger "github.com/ivankprod/ivankprod.ru/src/server/pkg/logger"
	"github.com/ivankprod/ivankprod.ru/src/server/pkg/utils"
)

var (
	MODE_DEV  bool
	MODE_PROD bool

	logger BaseLogger.ILogger
)

// App struct
type App struct {
	*fiber.App

	DBM *sqlx.DB
	DBT *tarantool.Connection
}

// Sitemap JSON to HTML
func loadSitemap(fileSitemapJSON *os.File) *string {
	infoSitemapJSON, err := fileSitemapJSON.Stat()
	if err != nil {
		logger.Printf("Error reading sitemap.json file: %v\n", err)
		logger.Fatalln("-- Server starting failed")
	}

	bytesSitemapJSON := make([]byte, infoSitemapJSON.Size())
	if _, err = fileSitemapJSON.Read(bytesSitemapJSON); err != nil {
		logger.Printf("Error reading sitemap.json file: %v\n", err)
		logger.Fatalln("-- Server starting failed")
	}

	if err = fileSitemapJSON.Close(); err != nil {
		logger.Printf("Error closing sitemap.json file: %v\n", err)
	}

	sitemap := &utils.Sitemap{}
	if err = json.Unmarshal(bytesSitemapJSON, sitemap); err != nil {
		logger.Printf("Error unmarshalling sitemap.json file: %v\n", err)
		logger.Fatalln("-- Server starting failed")
	}

	return sitemap.Nest().ToHTMLString()
}

func main() {
	// Logging file
	f, err := os.OpenFile("./logs/"+utils.DateMSK_ToLocaleSepString()+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.SetPrefix(utils.TimeMSK_ToLocaleString() + " ")
		log.Printf("Error opening log file: %v\n", err)
		log.Fatalln("-- Server starting failed")
	}

	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.SetPrefix(utils.TimeMSK_ToLocaleString() + " ")
			log.Printf("Error closing log file: %v\n", err)
		}
	}(f)

	logger = BaseLogger.New(io.MultiWriter(f, os.Stdout))
	logger.Println("-- Server starting...")

	// Load STAGE_MODE configuration
	if os.Getenv("STAGE_MODE") == "dev" {
		MODE_DEV = true
		MODE_PROD = false
	} else if os.Getenv("STAGE_MODE") == "prod" {
		MODE_DEV = false
		MODE_PROD = true
	}

	// Open sitemap.json file for reading
	fileSitemapJSON, err := os.Open("./misc/sitemap.json")
	if err != nil {
		logger.Printf("Error opening sitemap.json file: %v\n", err)
		logger.Fatalln("-- Server starting failed")
	}

	// Templates engine
	engine := handlebars.New("./views", ".hbs")

	// App
	app := App{
		App: fiber.New(fiber.Config{
			Prefork:       false,
			ErrorHandler:  router.HandleError,
			Views:         engine,
			StrictRouting: true,
		}),
	}

	// DB MySQL connect
	/*dbm, err := db.ConnectMySQL()
	if dbm == nil {
		if err == nil {
			logger.Println("Failed connecting to MySQL database")
		} else {
			logger.Printf("Error connecting to MySQL database: %v\n", err)
		}

		app.fail("-- Server starting failed\n")
	} else {
		app.DBM = dbm
	}*/

	// DB Tarantool connect
	dbt, err := db.ConnectTarantool()
	if dbt == nil {
		if err == nil {
			logger.Println("Failed connecting to Tarantool database")
		} else {
			logger.Printf("Error connecting to Tarantool database: %v\n", err)
		}

		app.fail("-- Server starting failed")
	} else {
		app.DBT = dbt
	}

	// Safe panic
	app.Use(recover.New())

	// Logger
	app.Use(RequestLogger.New(RequestLogger.Config{
		Format:     "${time} | ${method} | IP: ${ip} | STATUS: ${status} | URL: ${url}\n",
		TimeFormat: "02.01.2006 15:04:05",
		TimeZone:   "Europe/Moscow",
		Output:     f,
	}))

	// ContentSecurityPolicy
	var csp string

	if MODE_DEV {
		csp = "default-src 'self'; base-uri 'self'; block-all-mixed-content; font-src 'self' https: data:; frame-ancestors 'self'; img-src 'self' *.userapi.com *.fbsbx.com *.yandex.net *.googleusercontent.com data:; object-src 'none'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; script-src-attr 'none'; style-src 'self' https: 'unsafe-inline'; upgrade-insecure-requests"
	} else if MODE_PROD {
		csp = "default-src 'self'; base-uri 'self'; block-all-mixed-content; font-src 'self' https: data:; frame-ancestors 'self'; img-src 'self' *.userapi.com *.fbsbx.com *.yandex.net *.googleusercontent.com data:; object-src 'none'; script-src 'self' 'unsafe-inline'; script-src-attr 'none'; style-src 'self' https: 'unsafe-inline'; upgrade-insecure-requests"
	}

	app.Use(helmet.New(helmet.Config{
		ContentSecurityPolicy: csp,
	}))

	// Compression
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))

	// HTTP->HTTPS, without www & sitemap.xml
	app.Use(func(c *fiber.Ctx) error {
		if c.Protocol() == "http" || (len(c.Subdomains()) > 0) {
			return c.Redirect("https://"+os.Getenv("SERVER_HOST")+c.OriginalURL(), 301)
		}

		if c.OriginalURL() == "/sitemap.xml" {
			return c.SendFile("./sitemap.xml", true)
		}

		return c.Next()
	})

	// favicon
	app.Use(favicon.New(favicon.Config{
		File: "./favicon.ico",
	}))

	// Prometheus
	prometheus := fiberprometheus.New("ivankprodru_app")
	prometheus.RegisterAt(app.App, "/metrics")
	app.Use(prometheus.Middleware)

	// Static files
	app.Static("/static/", "./static", fiber.Static{Compress: true, MaxAge: 86400})

	// Setup router
	router.Router(app.App, app.DBT, loadSitemap(fileSitemapJSON))

	// HTTP listener
	go func() {
		logger.Printf("-- Attempt starting at %s:%s\n", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT_HTTP"))

		if err := app.Listen(":" + os.Getenv("SERVER_PORT_HTTP")); err != nil {
			logger.Println(err)
			app.fail(fmt.Sprintf("-- Server starting at %s:%s failed\n", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT_HTTP")))
		}
	}()

	// HTTPS certs
	cer, err := tls.LoadX509KeyPair(os.Getenv("SERVER_SSL_CERT"), os.Getenv("SERVER_SSL_KEY"))
	if err != nil {
		logger.Println(err)
		app.fail("-- Server starting failed\n")
	}

	// HTTPS listener
	config := &tls.Config{Certificates: []tls.Certificate{cer}, MinVersion: tls.VersionTLS13}
	ln, err := tls.Listen("tcp", ":"+os.Getenv("SERVER_PORT_HTTPS"), config)
	if err != nil {
		logger.Println(err)
		app.fail("-- Server starting failed\n")
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		app.exit()
	}()

	// LISTEN
	logger.Println("-- Attempt starting at " + os.Getenv("SERVER_HOST") + ":" + os.Getenv("SERVER_PORT_HTTPS"))

	if err = app.Listener(ln); err != nil {
		logger.Println(err)
		app.fail(fmt.Sprintf("-- Server starting at %s:%s failed\n", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT_HTTPS")))
	}
}

// App fail
func (app *App) fail(msg ...string) {
	if app.DBM != nil {
		if err := app.DBM.Close(); err != nil {
			logger.Println(err)
		}
	}
	if app.DBT != nil {
		if err := app.DBT.Close(); err != nil {
			logger.Println(err)
		}
	}

	if len(msg) > 0 {
		logger.Fatalln(strings.Trim(fmt.Sprint(msg), "[]"))
	} else {
		os.Exit(1)
	}
}

// App exit
func (app *App) exit(msg ...string) {
	if app.DBM != nil {
		if err := app.DBM.Close(); err != nil {
			logger.Println(err)
		}
	}
	if app.DBT != nil {
		if err := app.DBT.Close(); err != nil {
			logger.Println(err)
		}
	}

	if len(msg) > 0 {
		logger.Println(strings.Trim(fmt.Sprint(msg), "[]"))
	}

	logger.Println("-- Server terminated")
	_ = app.Shutdown()
}
