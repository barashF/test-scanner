package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	bookingRepository "github.com/internships-backend/test-backend-barashF/internal/repository/booking"
	roomRepository "github.com/internships-backend/test-backend-barashF/internal/repository/room"
	scheduleRepository "github.com/internships-backend/test-backend-barashF/internal/repository/schedule"

	authService "github.com/internships-backend/test-backend-barashF/internal/service/auth"
	bookingService "github.com/internships-backend/test-backend-barashF/internal/service/booking"
	roomService "github.com/internships-backend/test-backend-barashF/internal/service/room"
	scheduleSrvice "github.com/internships-backend/test-backend-barashF/internal/service/schedule" // Обратите внимание на опечатку в вашем коде: scheduleSrvice
	slotService "github.com/internships-backend/test-backend-barashF/internal/service/slot"

	"github.com/go-chi/chi/v5"
	_ "github.com/internships-backend/test-backend-barashF/docs"
	"github.com/internships-backend/test-backend-barashF/internal/config"
	"github.com/internships-backend/test-backend-barashF/internal/gateway"
	authHandler "github.com/internships-backend/test-backend-barashF/internal/handler/auth"
	bookingHandler "github.com/internships-backend/test-backend-barashF/internal/handler/booking"
	roomHandler "github.com/internships-backend/test-backend-barashF/internal/handler/room"
	scheduleHandler "github.com/internships-backend/test-backend-barashF/internal/handler/schedule"

	slotHandler "github.com/internships-backend/test-backend-barashF/internal/handler/slot"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/middleware"
	slotRepository "github.com/internships-backend/test-backend-barashF/internal/repository/slot"
	userRepo "github.com/internships-backend/test-backend-barashF/internal/repository/user"
	"github.com/internships-backend/test-backend-barashF/internal/repository/utils/transaction"
	"github.com/internships-backend/test-backend-barashF/pkg/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Room Booking Service
// @version         1.0
// @description     Сервис бронирования переговорок
// @termsOfService  http://swagger.io/terms/

// @host      localhost:8080
// @BasePath  /
func main() {
	appLogger, err := logger.NewZapAdapter()
	if err != nil {
		log.Fatalf("failed initialize logger: %v", err)
	}

	//nolint:errcheck
	defer appLogger.Sync()

	go func() {
		log.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
			log.Printf("pprof server failed: %v", err)
		}
	}()

	err = godotenv.Load()
	if err != nil {
		appLogger.Warn("error loading .env file", logger.F("error", err.Error()))
	}

	dbPool := database.MustInitDB()

	err = database.SeedUsers(context.Background(), dbPool, appLogger)
	if err != nil {
		appLogger.Warn("err create test users in database", logger.F("error", err.Error()))
	}

	transactionManager := transaction.NewManager(dbPool)
	userRepository := userRepo.NewRepository(transactionManager)
	slotRepository := slotRepository.NewRepository(transactionManager)
	roomRepository := roomRepository.NewRepository(transactionManager)
	scheduleRepository := scheduleRepository.NewRepository(transactionManager)
	bookingRepository := bookingRepository.NewRepository(transactionManager)

	serviceConference := gateway.NewServiceConference()

	cfg := config.Load()
	authService := authService.NewService(userRepository, cfg.JWTSecret, cfg.JWTTTL)
	scheduleService := scheduleSrvice.NewService(scheduleRepository, roomRepository, slotRepository, transactionManager, appLogger)
	roomService := roomService.NewService(roomRepository)
	slotService := slotService.NewService(slotRepository, roomRepository, appLogger)
	bookingService := bookingService.NewService(bookingRepository, slotRepository, serviceConference, transactionManager, appLogger)

	authHandler := authHandler.NewController(authService, appLogger)
	scheduleHandler := scheduleHandler.NewController(scheduleService, appLogger)
	roomHandler := roomHandler.NewController(roomService, appLogger)
	slotHandler := slotHandler.NewController(slotService, appLogger)
	bookingHandler := bookingHandler.NewController(bookingService, appLogger)

	startServer(
		dbPool,
		resolvePort(),
		authHandler,
		scheduleHandler,
		roomHandler,
		slotHandler,
		bookingHandler,
		cfg,
		appLogger,
	)
}

func startServer(
	dbPool *pgxpool.Pool,
	port string,
	authHandler *authHandler.Controller,
	scheduleHandler *scheduleHandler.Controller,
	roomHandler *roomHandler.Controller,
	slotHandler *slotHandler.Controller,
	bookingHandler *bookingHandler.Controller,
	cfg *config.Config,
	appLogger logger.Logger,
) {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: initRouter(authHandler, scheduleHandler, roomHandler, slotHandler, bookingHandler, cfg, appLogger),
	}
	appLogger.Info("Server started on ", logger.F("address", srv.Addr))

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	waitGracefullShutdown(srv, dbPool, serverErr, appLogger)
	appLogger.Info("Shutting down service-booking")
}

func resolvePort() string {
	port := os.Getenv("PORT")

	portFlag := flag.String("port", "", "укажите порт")
	flag.Parse()

	if portFlag != nil && *portFlag != "" {
		port = *portFlag
	}

	if port == "" {
		port = "8080"
	}

	return port
}

func initRouter(auth *authHandler.Controller,
	schedule *scheduleHandler.Controller,
	room *roomHandler.Controller,
	slot *slotHandler.Controller,
	booking *bookingHandler.Controller,
	cfg *config.Config, logger logger.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.MetricsMiddleware(logger))

		r.Get("/swagger/*", httpSwagger.WrapHandler)
		r.Get("/_info", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			//nolint:errcheck
			w.Write([]byte(`{"status":"OK"}`))
		})

		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/dummyLogin", auth.DummyLogin)

		r.Route("/rooms", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTSecret, logger))

			r.With(middleware.RolesAllowed(logger, "admin")).Post("/{roomId}/schedule/create", schedule.Create)
			r.With(middleware.RolesAllowed(logger, "admin")).Post("/create", room.Create)

			r.Get("/list", room.List)
			r.Get("/{roomId}/slots/list", slot.GetAvailableSlots)
		})

		r.Route("/bookings", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTSecret, logger))

			r.Group(func(r chi.Router) {
				r.Use(middleware.RolesAllowed(logger, "user"))
				r.Post("/create", booking.Create)
				r.Get("/my", booking.GetUserBookings)
				r.Post("/{bookingId}/cancel", booking.CancelBooking)
			})

			r.With(middleware.RolesAllowed(logger, "admin")).Get("/list", booking.GetAllBookings)
		})
	})

	return r
}

func waitGracefullShutdown(srv *http.Server, dbPool *pgxpool.Pool, serverErr <-chan error, appLogger logger.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var reason string
	select {
	case <-ctx.Done():
		reason = "signal"
	case err := <-serverErr:
		reason = "server error: " + err.Error()
	}
	appLogger.Info("Shutdown initiated", logger.F("reason", reason))

	appLogger.Info("Shutting down HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("HTTP server shutdown", logger.F("error", err))
	} else {
		appLogger.Info("HTTP server stopped")
	}

	appLogger.Info("Closing DB pool...")
	dbPool.Close()
	appLogger.Info("DB pool closed")
}
