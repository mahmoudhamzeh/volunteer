package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpserver "github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/http"
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
	if cfg.AppEnv == "production" && (cfg.JWTSecret == "mahak-dev-secret-change-me" || cfg.JWTSecret == "change-me-in-production") {
		log.Fatal("JWT_SECRET must be set to a strong value when APP_ENV=production")
	}

	pcfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database url:", err)
	}
	pcfg.MaxConns = int32(cfg.MaxConns)
	pcfg.MinConns = 2
	pcfg.MaxConnLifetime = time.Hour
	pcfg.MaxConnIdleTime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
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
				probe := rdb.SetNX(ctx, "lock:write-probe", "1", time.Second)
				if probe.Err() != nil {
					log.Println("redis is not writable, using in-process lock:", probe.Err())
					locker = lock.NewMemory()
				} else {
					_ = rdb.Del(ctx, "lock:write-probe").Err()
					locker = lock.NewResilient(rdb)
					log.Println("redis lock enabled")
				}
			} else {
				log.Println("redis unavailable, using in-process lock:", err)
			}
		}
	}

	auth := authuc.New(db.Users(), db.Volunteers(), cfg.JWTSecret, cfg.JWTTTL)
	auth.SetRevealOTP(cfg.RevealOTP)
	skills := db.Skills()
	if err := skills.SeedDefaults(ctx); err != nil {
		log.Println("seed skills:", err)
	}
	vol := volunteeruc.New(db.Users(), db.Volunteers(), storage, db.Notifications(), skills, nil)
	tasks := taskuc.New(db.Tasks(), db.Volunteers(), db.Certificates(), locker, db.Notifications(), nil)
	missions := missionuc.New(db.Missions(), db.Volunteers(), db.Notifications(), nil, nil)
	certs := certuc.New(db.Certificates(), db.Tasks(), db.Volunteers(), nil, cfg.PublicBase)

	if cfg.SeedDemo {
		postgres.Demo(ctx, db.Users(), db.Volunteers(), tasks, missions, vol, auth, skills)
	}

	r := httpserver.NewRouter(httpserver.Deps{
		Auth:          auth,
		Volunteers:    vol,
		Tasks:         tasks,
		Missions:      missions,
		Certs:         certs,
		Users:         db.Users(),
		Stats:         db.Stats(),
		Notify:        db.Notifications(),
		Storage:       storage,
		InternalToken: cfg.InternalToken,
		CORSOrigins:   cfg.CORSOrigins,
		Ready:         func(c context.Context) error { return pool.Ping(c) },
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		log.Println("mahak-volunteer-api listening on", cfg.Addr, "env", cfg.AppEnv)
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
	Addr          string
	AppEnv        string
	DatabaseURL   string
	JWTSecret     string
	JWTTTL        time.Duration
	StorageDir    string
	RedisURL      string
	PublicBase    string
	InternalToken string
	CORSOrigins   []string
	SeedDemo      bool
	RevealOTP     bool
	MaxConns      int32
}

func loadConfig() config {
	ttlHours := envInt("JWT_TTL_HOURS", 24)
	maxConns := envInt("DB_MAX_CONNS", 40)
	appEnv := env("APP_ENV", "development")
	seedDefault := appEnv != "production"
	return config{
		Addr:          env("HTTP_ADDR", ":8080"),
		AppEnv:        appEnv,
		DatabaseURL:   env("DATABASE_URL", "postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable"),
		JWTSecret:     env("JWT_SECRET", "mahak-dev-secret-change-me"),
		JWTTTL:        time.Duration(ttlHours) * time.Hour,
		StorageDir:    env("STORAGE_DIR", "./data/uploads"),
		RedisURL:      env("REDIS_URL", "redis://127.0.0.1:6379"),
		PublicBase:    env("PUBLIC_BASE_URL", "http://localhost:3000"),
		InternalToken: env("INTERNAL_API_TOKEN", env("WEBHOOK_SECRET", "")),
		CORSOrigins:   splitCSV(env("CORS_ORIGINS", "*")),
		SeedDemo:      envBool("SEED_DEMO", seedDefault),
		RevealOTP:     envBool("OTP_REVEAL", appEnv != "production"),
		MaxConns:      int32(maxConns),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return d
	}
	return n
}

func envBool(k string, d bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(k)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return d
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
