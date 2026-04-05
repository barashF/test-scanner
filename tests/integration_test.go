//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/internships-backend/test-backend-barashF/internal/gateway"
	authHandler "github.com/internships-backend/test-backend-barashF/internal/handler/auth"
	bookingHandler "github.com/internships-backend/test-backend-barashF/internal/handler/booking"
	authDto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/auth"
	bkgDto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/booking"
	roomDto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/room"
	scheduleDto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/schedule"
	roomHandler "github.com/internships-backend/test-backend-barashF/internal/handler/room"
	schHandler "github.com/internships-backend/test-backend-barashF/internal/handler/schedule"
	slotHandler "github.com/internships-backend/test-backend-barashF/internal/handler/slot"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/middleware"
	bkgRep "github.com/internships-backend/test-backend-barashF/internal/repository/booking"
	roomRepo "github.com/internships-backend/test-backend-barashF/internal/repository/room"
	schRepo "github.com/internships-backend/test-backend-barashF/internal/repository/schedule"
	slotRepo "github.com/internships-backend/test-backend-barashF/internal/repository/slot"
	"github.com/internships-backend/test-backend-barashF/internal/repository/user"
	"github.com/internships-backend/test-backend-barashF/internal/repository/utils/transaction"
	authServ "github.com/internships-backend/test-backend-barashF/internal/service/auth"
	bkgServ "github.com/internships-backend/test-backend-barashF/internal/service/booking"
	roomService "github.com/internships-backend/test-backend-barashF/internal/service/room"
	schServ "github.com/internships-backend/test-backend-barashF/internal/service/schedule"
	slotServ "github.com/internships-backend/test-backend-barashF/internal/service/slot"
)

type BookingSuite struct {
	suite.Suite
	conn   *pgxpool.Pool
	server *httptest.Server
	client *http.Client
	roomID uuid.UUID

	userToken  string
	adminToken string
}

func (s *BookingSuite) SetupSuite() {
	s.conn = mustInitDB()

	l := noOpLogger{}
	txManager := transaction.NewManager(s.conn)
	repRoom := roomRepo.NewRepository(txManager)
	schRep := schRepo.NewRepository(txManager)
	slotRep := slotRepo.NewRepository(txManager)
	bkgRep := bkgRep.NewRepository(txManager)
	userRep := user.NewRepository(txManager)

	confSrv := gateway.NewServiceConference()

	servSch := schServ.NewService(schRep, repRoom, slotRep, txManager, l)
	servRoom := roomService.NewService(repRoom)
	servBkg := bkgServ.NewService(bkgRep, slotRep, confSrv, txManager, l)
	slotSrv := slotServ.NewService(slotRep, repRoom, l)
	authSrv := authServ.NewService(userRep, "test_secret", 10*time.Minute)

	ctrlAuth := authHandler.NewController(authSrv, l)
	ctrlRoom := roomHandler.NewController(servRoom, l)
	ctrlSch := schHandler.NewController(servSch, l)
	ctrlBkg := bookingHandler.NewController(servBkg, l)
	ctrlSlot := slotHandler.NewController(slotSrv, l)

	r := chi.NewRouter()
	r.Post("/dummyLogin", ctrlAuth.DummyLogin)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware("test_secret", l))

		// role admin
		r.Group(func(r chi.Router) {
			r.Use(middleware.RolesAllowed(l, "admin"))
			r.Post("/rooms/create", ctrlRoom.Create)
			r.Post("/rooms/{roomId}/schedule/create", ctrlSch.Create)
		})

		// role user
		r.Group(func(r chi.Router) {
			r.Use(middleware.RolesAllowed(l, "user", "admin"))
			r.Get("/rooms/{roomId}/slots/list", ctrlSlot.GetAvailableSlots)
			r.Post("/bookings", ctrlBkg.Create)
			r.Post("/bookings/{bookingId}/cancel", ctrlBkg.CancelBooking)
		})
	})

	s.server = httptest.NewServer(r)
	s.client = s.server.Client()
}

func (s *BookingSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *BookingSuite) TestCreateBooking_SuccessFlow() {
	// auth
	dummyLoginAdmin := &authDto.DummyLoginRequest{
		Role: "admin",
	}
	dummyLoginUser := &authDto.DummyLoginRequest{
		Role: "user",
	}

	body, _ := json.Marshal(dummyLoginAdmin)
	resp, err := s.client.Post(s.server.URL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var AccesToken authDto.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&AccesToken)
	s.Require().NoError(err)
	s.adminToken = AccesToken.AccessToken

	body, _ = json.Marshal(dummyLoginUser)
	resp, err = s.client.Post(s.server.URL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&AccesToken)
	s.Require().NoError(err)
	s.userToken = AccesToken.AccessToken

	// create room
	roomReq := roomDto.CreateRequest{
		Name:        "Conference Room 404",
		Capacity:    10,
		Description: "For testing",
	}

	body, _ = json.Marshal(roomReq)
	req, _ := http.NewRequest("POST", s.server.URL+"/rooms/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var roomResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&roomResp)
	s.Require().NoError(err)

	s.roomID, err = uuid.Parse(roomResp["id"].(string))
	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, s.roomID)

	// create schedule
	scheduleReq := scheduleDto.CreateRequest{
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	body, _ = json.Marshal(scheduleReq)
	url := fmt.Sprintf("%s/rooms/%s/schedule/create", s.server.URL, s.roomID.String())
	req, _ = http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	// get slots of room
	targetDate := "2026-04-06"

	getSlotsURL := fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", s.server.URL, s.roomID.String(), targetDate)

	req, _ = http.NewRequest("GET", getSlotsURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var listResult struct {
		Slots []map[string]any `json:"slots"`
	}

	err = json.NewDecoder(resp.Body).Decode(&listResult)
	s.Require().NoError(err)

	s.Require().NotEmpty(listResult.Slots, "No available slots returned for the created schedule!")

	slotIDStr, ok := listResult.Slots[0]["id"].(string)
	s.Require().True(ok, "Slot ID is missing or not a string")

	slotID, err := uuid.Parse(slotIDStr)
	s.Require().NoError(err)

	// booking room
	bookingReq := bkgDto.CreateRequest{
		SlotID: slotID,
	}

	body, _ = json.Marshal(bookingReq)
	req, _ = http.NewRequest("POST", s.server.URL+"/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)
}

func (s *BookingSuite) TestCancelBooking_SuccessFlow() {
	// auth
	dummyLoginAdmin := &authDto.DummyLoginRequest{
		Role: "admin",
	}
	dummyLoginUser := &authDto.DummyLoginRequest{
		Role: "user",
	}

	body, _ := json.Marshal(dummyLoginAdmin)
	resp, err := s.client.Post(s.server.URL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var AccesToken authDto.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&AccesToken)
	s.Require().NoError(err)
	s.adminToken = AccesToken.AccessToken

	body, _ = json.Marshal(dummyLoginUser)
	resp, err = s.client.Post(s.server.URL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&AccesToken)
	s.Require().NoError(err)
	s.userToken = AccesToken.AccessToken

	// create room
	roomReq := roomDto.CreateRequest{
		Name:        "Conference Room 404",
		Capacity:    10,
		Description: "For testing",
	}

	body, _ = json.Marshal(roomReq)
	req, _ := http.NewRequest("POST", s.server.URL+"/rooms/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var roomResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&roomResp)
	s.Require().NoError(err)

	s.roomID, err = uuid.Parse(roomResp["id"].(string))
	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, s.roomID)

	// create schedule
	scheduleReq := scheduleDto.CreateRequest{
		DaysOfWeek: []int{1, 2, 3, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	body, _ = json.Marshal(scheduleReq)
	url := fmt.Sprintf("%s/rooms/%s/schedule/create", s.server.URL, s.roomID.String())
	req, _ = http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	// get slots of room
	targetDate := "2026-04-07"

	getSlotsURL := fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", s.server.URL, s.roomID.String(), targetDate)

	req, _ = http.NewRequest("GET", getSlotsURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var listResult struct {
		Slots []map[string]any `json:"slots"`
	}

	err = json.NewDecoder(resp.Body).Decode(&listResult)
	s.Require().NoError(err)

	s.Require().NotEmpty(listResult.Slots, "No available slots returned for the created schedule!")

	slotIDStr, ok := listResult.Slots[0]["id"].(string)
	s.Require().True(ok, "Slot ID is missing or not a string")

	slotID, err := uuid.Parse(slotIDStr)
	s.Require().NoError(err)

	// booking room
	bookingReq := bkgDto.CreateRequest{
		SlotID: slotID,
	}
	fmt.Println(slotID)
	body, _ = json.Marshal(bookingReq)
	req, _ = http.NewRequest("POST", s.server.URL+"/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	var createResult struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}

	err = json.NewDecoder(resp.Body).Decode(&createResult)
	s.Require().NoError(err)

	bookingIDStr := createResult.Booking.ID
	s.Require().NotEmpty(bookingIDStr, "empty bookingID")

	s.Require().True(ok, "failed get ID booking")
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	fmt.Println(bookingIDStr)

	url = fmt.Sprintf("%s/bookings/%s/cancel", s.server.URL, bookingIDStr)
	req, _ = http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *BookingSuite) TearDownTest() {
	ctx := context.Background()

	s.conn.Exec(
		ctx,
		`DELETE FROM rooms WHERE id = $1 CASCADE`,
		s.roomID,
	)
}

func TestBookingSuite(t *testing.T) {
	suite.Run(t, new(BookingSuite))
}

func mustInitDB() *pgxpool.Pool {
	var dbPool *pgxpool.Pool

	config, err := pgxpool.ParseConfig(getConnectionString())
	if err != nil {
		log.Fatalf("Unable parse connection string: %v\n", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	pingAttemptsLimit := 3
	var pingErr error

	for i := 1; i <= pingAttemptsLimit; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr = dbPool.Ping(pingCtx)
		pingCancel()

		if pingErr == nil {
			break
		}
		log.Printf("db ping attempt %d failed: %v", i, pingErr)
		if i < pingAttemptsLimit {
			time.Sleep(300 * time.Millisecond)
		}
	}

	if pingErr != nil {
		log.Fatalf("Unable ti ping databse")
	}

	log.Println("Database connection pool established")
	return dbPool
}

func getConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		"postgres",
		"postgres",
		"localhost",
		"5432",
		"booking_db",
	)
}

type noOpLogger struct{}

func (n noOpLogger) Debug(m string, f ...logger.Field) {}
func (n noOpLogger) Info(m string, f ...logger.Field)  {}
func (n noOpLogger) Warn(m string, f ...logger.Field)  {}
func (n noOpLogger) Error(m string, f ...logger.Field) {}
func (n noOpLogger) Sync() error                       { return nil }
