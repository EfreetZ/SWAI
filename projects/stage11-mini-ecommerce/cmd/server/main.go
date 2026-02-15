package main

import (
	"context"
	"log"
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	httpserver "github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/http"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	p, err := a.ProductSvc.Create(context.Background(), "demo", 100)
	if err != nil {
		log.Fatal(err)
	}
	if err = a.InventorySvc.Seed(context.Background(), p.ID, 1000); err != nil {
		log.Fatal(err)
	}
	s := httpserver.New(a)
	log.Fatal(http.ListenAndServe(":18110", s.Handler()))
}
