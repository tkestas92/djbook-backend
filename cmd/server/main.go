package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/graph"
	"github.com/djbook/backend/internal/graph/resolvers"
	apihandler "github.com/djbook/backend/internal/handler"
	"github.com/djbook/backend/internal/repository"
	"github.com/djbook/backend/internal/service"
)

func main() {
	// 1. Load .env (ignore error in production where env vars are set directly)
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	// 2. Connect MySQL
	dsn := mustEnv("MYSQL_DSN")
	db, err := connectMySQL(dsn)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer db.Close()

	// 3. Run migrations
	if err := runMigrations(db, "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// 4. Connect Redis (optional — used for token invalidation)
	var rdb *redis.Client
	if redisURL := getEnv("REDIS_URL", ""); redisURL != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr: redisURL,
		})
		if _, err := rdb.Ping(context.Background()).Result(); err != nil {
			log.Printf("warning: redis not available: %v — continuing without token invalidation", err)
			rdb.Close()
			rdb = nil
		}
	} else {
		log.Println("REDIS_URL not set — continuing without token invalidation")
	}

	// 5. Build repositories, services, and resolvers
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	eventRepo := repository.NewEventRepository(db)
	financeRepo := repository.NewFinanceRepository(db)

	userSvc := service.NewUserService(userRepo)
	profileSvc := service.NewProfileService(profileRepo)
	eventSvc := service.NewEventService(eventRepo)
	financeSvc := service.NewFinanceService(financeRepo)

	rootResolver := resolvers.NewResolver(userSvc, profileSvc, eventSvc, financeSvc)

	// 6. Setup gqlgen server
	jwtSecret := getEnv("JWT_SECRET", "djbook-default-secret-change-in-production")

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: rootResolver,
	}))

	// 7. Ensure upload directory exists
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	photoDir := filepath.Join(uploadDir, "photos")
	if err := os.MkdirAll(photoDir, 0755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	// 8. Apply JWT middleware and register routes
	jwtMiddleware := auth.JWTMiddleware(jwtSecret, rdb)

	mux := http.NewServeMux()

	// Auth endpoints
	authHandler := auth.NewHandler(userSvc, profileSvc, jwtSecret)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/google", handleGoogleAuth(userSvc, jwtSecret))
	mux.HandleFunc("/auth/apple", handleAppleAuth(userSvc, jwtSecret))

	// Authenticated user endpoint
	meHandler := apihandler.NewMeHandler(userSvc, profileSvc, eventSvc)
	mux.Handle("/me", jwtMiddleware(http.HandlerFunc(meHandler.GetMe)))

	// Photo upload endpoint
	uploadHandler := apihandler.NewUploadHandler(profileSvc, photoDir)
	mux.Handle("/upload/photo", jwtMiddleware(http.HandlerFunc(uploadHandler.UploadPhoto)))

	// Static file serving for uploaded photos
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	// Public SoundCloud profile sets (no auth)
	soundcloudHandler := apihandler.NewSoundCloudHandler()
	mux.HandleFunc("/soundcloud/sets", soundcloudHandler.GetSets)

	// GraphQL playground (unauthenticated) and query endpoint (JWT optional — resolvers enforce auth)
	mux.Handle("/", playground.Handler("DJBook API", "/query"))
	mux.Handle("/query", jwtMiddleware(srv))

	port := getEnv("PORT", "8080")
	log.Printf("DJBook backend listening on :%s", port)
	log.Printf("GraphQL playground: http://localhost:%s/", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// connectMySQL opens a MySQL connection with retry logic.
func connectMySQL(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if err = db.Ping(); err == nil {
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(10)
			db.SetConnMaxLifetime(5 * time.Minute)
			return db, nil
		}
		log.Printf("waiting for MySQL (attempt %d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to MySQL after retries: %w", err)
}

// runMigrations executes all .sql migration files in order.
func runMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		for _, stmt := range splitSQLStatements(string(content)) {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("execute migration %s: %w", f, err)
			}
		}
		log.Printf("applied migration: %s", filepath.Base(f))
	}
	return nil
}

// handleGoogleAuth handles Google Sign In token exchange.
func handleGoogleAuth(userSvc *service.UserService, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idToken := r.FormValue("id_token")
		if idToken == "" {
			http.Error(w, "id_token required", http.StatusBadRequest)
			return
		}

		googleID, email, err := auth.VerifyGoogleToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "invalid google token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := userSvc.GetOrCreateByGoogle(r.Context(), googleID, email)
		if err != nil {
			http.Error(w, "user error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateToken(user.ID, jwtSecret)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":%q,"userId":%q}`, token, user.ID)
	}
}

// handleAppleAuth handles Apple Sign In token exchange.
func handleAppleAuth(userSvc *service.UserService, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idToken := r.FormValue("id_token")
		if idToken == "" {
			http.Error(w, "id_token required", http.StatusBadRequest)
			return
		}

		appleID, email, err := auth.VerifyAppleToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "invalid apple token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := userSvc.GetOrCreateByApple(r.Context(), appleID, email)
		if err != nil {
			http.Error(w, "user error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateToken(user.ID, jwtSecret)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":%q,"userId":%q}`, token, user.ID)
	}
}

func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	var statements []string
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
