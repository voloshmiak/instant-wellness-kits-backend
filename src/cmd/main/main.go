package main

import (
	"InstantWellnessKits/src/config"
	"InstantWellnessKits/src/controller"
	"InstantWellnessKits/src/repository/geocoder"
	"InstantWellnessKits/src/repository/postgres"
	"InstantWellnessKits/src/repository/postgres/order"
	tax_rate "InstantWellnessKits/src/repository/postgres/tax-rate"
	"InstantWellnessKits/src/usecase"
	"github.com/rs/cors"
	"log"
	"net/http"
	"time"
)

const (
	writeTimeout = 15 * time.Second
	readTimeout  = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	router := http.NewServeMux()

	geocoderApi := geocoder.NewApi(cfg.GeocodingAPIKey)

	conn, err := postgres.InitDb(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = postgres.ApplyMigrations(conn)
	if err != nil {
		return err
	}

	err = postgres.SeedTaxRates(conn)
	if err != nil {
		return err
	}

	taxRateRepo := tax_rate.NewRepository(conn)
	orderRepo := order.NewRepository(conn)

	createUsecase := usecase.NewCreateOrderUseCase(geocoderApi, orderRepo, taxRateRepo)
	listUsecase := usecase.NewListOrdersUseCase(orderRepo)
	importUsecase := usecase.NewImportOrdersUseCase(geocoderApi, orderRepo, taxRateRepo)

	importController := controller.NewImportHandler(importUsecase)
	createController := controller.NewCreateController(createUsecase)
	getController := controller.NewGetHandler(listUsecase)

	router.Handle("POST /orders/import", importController)
	router.Handle("POST /orders", createController)
	router.Handle("GET /orders", getController)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(router)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		Handler:      handler,
	}

	log.Println("Listening on", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
