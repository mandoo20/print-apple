package boot

import (
	"log"
	"net/http"

	"print-apple/internal/config"

	// "github.com/jmoiron/sqlx"

	appleData "print-apple/internal/data/apple"
	// userData "print-apple/internal/data/user"
	server "print-apple/internal/delivery/http"
	appleHandler "print-apple/internal/delivery/http/apple"
	appleService "print-apple/internal/service/apple"

	// userService "print-apple/internal/service/user"
	kConsumer "print-apple/internal/delivery/kafka"
	firebaseclient "print-apple/pkg/firebaseClient"
	"print-apple/pkg/kafka"
)

// HTTP will load configuration, do dependency injection and then start the HTTP server
func HTTP() error {
	var (
		s   server.Server         // HTTP Server Object
		ad  appleData.Data        // User domain data layer
		as  appleService.Service  // User domain service layer
		ah  *appleHandler.Handler // User domain handler
		cfg *config.Config        // Configuration object
		fb  *firebaseclient.Client
		k   *kafka.Kafka // Kafka Producer
	)

	// Get configuration
	err := config.Init()
	if err != nil {
		log.Fatalf("[CONFIG] Failed to initialize config: %v", err)
	}
	cfg = config.Get()

	fb, err = firebaseclient.NewClient(cfg)
	if err != nil {
		log.Fatalf("[DB] Failed to initialize database connection: %v", err)
	}

	log.Println(cfg.Kafka.Brokers)
	k, err = kafka.New(cfg.Kafka.Username, cfg.Kafka.Password, cfg.Kafka.Brokers)
	if err != nil {
		log.Fatalf("[KAFKA] Failed to initialize kafka producer: %v", err)
	}

	// Apple domain initialization
	ad = appleData.New(fb)
	as = appleService.New(ad)
	ah = appleHandler.New(as)

	// Inject service used on handler
	s = server.Server{
		Apple: ah,
	}

	// Error Handling
	if err := s.Serve(cfg.Server.Port); err != http.ErrServerClosed {
		return err
	}

	go kConsumer.New(as, k, cfg.Kafka.Subscriptions)
	// Error Handling
	if err := s.Serve(cfg.Server.Port); err != http.ErrServerClosed {
		return err
	}

	return nil
}