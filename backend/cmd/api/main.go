package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/http"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/lock"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/postgres"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/storage/localfs"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/authuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/certuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := loadConfig()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("postgres:", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		log.Fatal("migrate:", err)
	}

	db := postgres.New(pool)
	storage, err := localfs.New(cfg.StorageDir)
	if err != nil {
		log.Fatal(err)
	}

	var locker interface {
		Lock(context.Context, string, time.Duration) (func(), error)
	} = lock.NewMemory()
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err == nil {
			rdb := redis.NewClient(opt)
			if err := rdb.Ping(ctx).Err(); err == nil {
				locker = lock.NewRedis(rdb)
				log.Println("redis lock enabled")
			}
		}
	}

	auth := authuc.New(db.Users(), db.Volunteers(), cfg.JWTSecret, 24*time.Hour)
	skills := db.Skills()
	if err := skills.SeedDefaults(ctx); err != nil {
		log.Println("seed skills:", err)
	}
	vol := volunteeruc.New(db.Users(), db.Volunteers(), storage, db.Notifications(), skills, nil)
	tasks := taskuc.New(db.Tasks(), db.Volunteers(), db.Certificates(), locker, db.Notifications(), nil)
	missions := missionuc.New(db.Missions(), db.Volunteers(), db.Notifications(), nil)
	certs := certuc.New(db.Certificates(), db.Tasks(), db.Volunteers(), nil, cfg.PublicBase)

	postgres.Demo(ctx, db.Users(), db.Volunteers(), tasks, missions, vol, auth, skills)

	r := httpserver.NewRouter(httpserver.Deps{
		Auth:       auth,
		Volunteers: vol,
		Tasks:      tasks,
		Missions:   missions,
		Certs:      certs,
		Users:      db.Users(),
		Stats:      db.Stats(),
		Notify:     db.Notifications(),
	})

	srv := &http.Server{Addr: cfg.Addr, Handler: r}
	go func() {
		log.Println("mahak volunteers api listening on", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shctx, c := context.WithTimeout(context.Background(), 8*time.Second)
	defer c()
	_ = srv.Shutdown(shctx)
}

type config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	StorageDir  string
	RedisURL    string
	PublicBase  string
}

func loadConfig() config {
	return config{
		Addr:        env("HTTP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "mahak-dev-secret-change-me"),
		StorageDir:  env("STORAGE_DIR", "./data/uploads"),
		RedisURL:    env("REDIS_URL", "redis://127.0.0.1:6379"),
		PublicBase:  env("PUBLIC_BASE_URL", "http://localhost:3000"),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
