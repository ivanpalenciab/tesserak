package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/compras"
	"tesserak/backend/internal/config"
	"tesserak/backend/internal/consecutivos"
	"tesserak/backend/internal/crm"
	"tesserak/backend/internal/database"
	"tesserak/backend/internal/empresas"
	"tesserak/backend/internal/inventario"
	"tesserak/backend/internal/precios"
	"tesserak/backend/internal/productos"
	"tesserak/backend/internal/terceros"
	"tesserak/backend/internal/usuarios"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "aplica las migraciones y termina")
	skipMigrate := flag.Bool("skip-migrate", false, "no aplica migraciones al iniciar")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	gin.SetMode(cfg.GinMode)

	db, err := database.Connect(cfg.DatabaseURL, cfg.GinMode == gin.DebugMode)
	if err != nil {
		log.Fatal(err)
	}

	if !*skipMigrate {
		if err := database.Migrate(db); err != nil {
			log.Fatal(err)
		}
		log.Println("migraciones aplicadas")
	}
	if *migrateOnly {
		return
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg.CORSOrigin))

	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	auth.NewHandler(auth.NewService(db, cfg.JWTSecret)).RegisterRoutes(api)
	empresas.NewHandler(empresas.NewService(empresas.NewRepository(db))).RegisterRoutes(api)
	usuarios.NewHandler(usuarios.NewService(usuarios.NewRepository(db))).RegisterRoutes(api)
	crm.NewHandler(crm.NewService(crm.NewRepository(db))).
		WithAuthMiddleware(auth.RequireAuth(cfg.JWTSecret)).
		RegisterRoutes(api)
	productos.NewHandler(
		productos.NewService(productos.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)
	precios.NewHandler(
		precios.NewService(precios.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)
	compras.NewHandler(
		compras.NewService(compras.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)
	terceros.NewHandler(
		terceros.NewService(terceros.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)
	consecutivos.NewHandler(
		consecutivos.NewService(consecutivos.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)
	inventario.NewHandler(
		inventario.NewService(inventario.NewRepository(db)),
		auth.RequireAuth(cfg.JWTSecret),
	).RegisterRoutes(api)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("API escuchando en :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error cerrando servidor: %v", err)
	}
	log.Println("servidor detenido")
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
