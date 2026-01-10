// package main - starting point of the service
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mihailtudos/microservices/auth/internal/auth"
	desc "github.com/mihailtudos/microservices/auth/pkg/auth_v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrMissingDBString = errors.New("failed to connect - DB string")
	ErrPingDB          = errors.New("failed to ping DB")
)

type server struct {
	desc.UnimplementedAuthV1Server
	queries *auth.Queries
	db      *pgxpool.Pool
}

// HashPassword generates a bcrypt hash for the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	ur := req.GetInfo()

	fmt.Printf("user request: %+v\n", ur)
	// userRoleID := "3face0d0-9f17-44be-87e8-ece2504bd2cd"
	// roleUUID, err := uuid.Parse(userRoleID)
	// if err != nil {
	// 	return nil, err
	// }

	if ur.Password != ur.PasswordConfirm {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"password not valid",
		)
	}

	hash, _ := HashPassword(ur.Password)
	fmt.Println("valid pw")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Internal server error",
		)
	}

	defer tx.Rollback(ctx)

	q := s.queries.WithTx(tx)
	fmt.Println("start tx")

	adminRoleId, err := q.GetRoleByName(ctx, "admin")
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Internal server error",
		)
	}

	newUser, err := q.CreateUser(ctx, auth.CreateUserParams{
		ID:           uuid.New(),
		Name:         ur.Name,
		Email:        ur.Email,
		RoleID:       adminRoleId,
		PasswordHash: hash,
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to create new user",
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to persist",
		)
	}

	return &desc.CreateResponse{
		Id: newUser.ID.String(),
	}, nil
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	userID := req.GetId()
	uID, err := uuid.Parse(userID)

	log.Printf("User id: %s", userID)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to get user: %v",
			err,
		)
	}

	user, err := s.queries.GetUserById(ctx, uID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(
				codes.NotFound,
				"user with id %s not found",
				userID,
			)
		}

		return nil, status.Errorf(
			codes.Internal,
			"failed to get user: %v",
			err,
		)
	}

	return &desc.GetResponse{
		User: &desc.User{
			Id: user.ID.String(),
			User: &desc.UserInfo{
				Name:  user.Name,
				Email: user.Email,
				Role:  desc.Role_ADMIN,
			},
			CreatedAt: timestamppb.New(user.CreatedAt.Time),
			UpdatedAt: timestamppb.New(user.UpdatedAt.Time),
		},
	}, nil
}

func setupDB(ctx context.Context) (*pgxpool.Pool, error) {
	dbString := os.Getenv("DATABASE_URL")
	if dbString == "" {
		dbString = os.Getenv("GOOSE_DBSTRING")
	}

	cfg, err := pgxpool.ParseConfig(dbString)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Sensible defaults (tune for prod)
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

func main() {
	if os.Getenv("PORT") == "" {
		log.Fatal("port not provided\n")
	}
	
	ctx := context.Background()

	db, err := setupDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}

	defer db.Close()

	queries := auth.New(db)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatal("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	authServer := &server{
		queries: queries,
		db:      db,
	}

	desc.RegisterAuthV1Server(s, authServer)

	log.Printf("server listening at %s", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
