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
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
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
	if err := cfg.validate(); err != nil {
		log.Fatal(err)
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database url:", err)
	}
	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
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

	var rdb *redis.Client
	var locker domain.Locker = lock.NewMemory()
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatal("redis url:", err)
		}
		opt.PoolSize = 32
		opt.MinIdleConns = 4
		client := redis.NewClient(opt)
		if err := client.Ping(ctx).Err(); err != nil {
			if cfg.Production {
				log.Fatal("redis:", err)
			}
			log.Println("redis unavailable, using in-process lock:", err)
		} else {
			rdb = client
			locker = lock.NewRedis(rdb)
			log.Println("redis lock enabled")
		}
		defer client.Close()
	} else if cfg.Production {
		log.Fatal("REDIS_URL is required in production so capacity locks work across processes")
	} else {
		log.Println("REDIS_URL empty; using in-process lock")
	}

	auth := authuc.New(db.Users(), db.Volunteers(), cfg.JWTSecret, cfg.JWTTTL)
	vol := volunteeruc.New(db.Users(), db.Volunteers(), storage, db.Notifications(), nil)
	tasks := taskuc.New(db.Tasks(), db.Volunteers(), db.Certificates(), locker, db.Notifications(), nil)
	missions := missionuc.New(db.Missions(), db.Volunteers(), db.Notifications(), nil)
	certs := certuc.New(db.Certificates(), db.Tasks(), db.Volunteers(), nil, cfg.PublicBase)

	if cfg.SeedDemo {
		postgres.Demo(ctx, db.Users(), db.Volunteers(), tasks, missions, vol, auth)
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
		InternalKey:   cfg.InternalKey,
		WebhookSecret: cfg.WebhookSecret,
		CORSOrigins:   cfg.CORSOrigins,
		Production:    cfg.Production,
		Ready: func() map[string]string {
			out := map[string]string{"status": "ready", "postgres": "ok"}
			pctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := pool.Ping(pctx); err != nil {
				out["status"] = "not_ready"
				out["postgres"] = "error"
			}
			if rdb != nil {
				out["redis"] = "ok"
				if err := rdb.Ping(pctx).Err(); err != nil {
					out["status"] = "not_ready"
					out["redis"] = "error"
				}
			}
			return out
		},
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
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
	Addr          string
	DatabaseURL   string
	JWTSecret     string
	JWTTTL        time.Duration
	StorageDir    string
	RedisURL      string
	PublicBase    string
	InternalKey   string
	WebhookSecret string
	CORSOrigins   []string
	Production    bool
	SeedDemo      bool
	DBMaxConns    int32
	DBMinConns    int32
}

func (c config) validate() error {
	if c.Production {
		if c.JWTSecret == "" || c.JWTSecret == "mahak-dev-secret-change-me" || c.JWTSecret == "change-me-in-production" {
			return errConfig("JWT_SECRET must be a strong unique value in production")
		}
		if len(c.JWTSecret) < 32 {
			return errConfig("JWT_SECRET must be at least 32 characters in production")
		}
	}
	return nil
}

type configError string

func (e configError) Error() string { return string(e) }
func errConfig(s string) error      { return configError(s) }

func loadConfig() config {
	appEnv := strings.ToLower(env("APP_ENV", "development"))
	production := appEnv == "production" || appEnv == "prod"
	seedDefault := "true"
	if production {
		seedDefault = "false"
	}
	cors := splitCSV(env("CORS_ORIGINS", ""))
	if len(cors) == 0 && !production {
		cors = []string{"*"}
	}
	return config{
		Addr:          env("HTTP_ADDR", ":8080"),
		DatabaseURL:   env("DATABASE_URL", "postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable"),
		JWTSecret:     env("JWT_SECRET", "mahak-dev-secret-change-me"),
		JWTTTL:        time.Duration(envInt("JWT_TTL_HOURS", 24)) * time.Hour,
		StorageDir:    env("STORAGE_DIR", "./data/uploads"),
		RedisURL:      env("REDIS_URL", "redis://127.0.0.1:6379"),
		PublicBase:    env("PUBLIC_BASE_URL", "http://localhost:3000"),
		InternalKey:   os.Getenv("INTERNAL_API_KEY"),
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
		CORSOrigins:   cors,
		Production:    production,
		SeedDemo:      env("SEED_DEMO", seedDefault) == "true",
		DBMaxConns:    int32(envInt("DB_MAX_CONNS", 40)),
		DBMinConns:    int32(envInt("DB_MIN_CONNS", 4)),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return d
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
