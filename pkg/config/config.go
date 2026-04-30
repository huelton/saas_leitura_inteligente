package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/huelton/leitura-inteligente/internal/limiter"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	AIAPIKey      string
	AIModel       string
	Environment   string
	HTTPAddr      string
	PageChunkSize int
	JWTSecret     string
	JWTExpiry     time.Duration
	Limits        limiter.Config
	Storage       StorageConfig
	Security      SecurityConfig
}

// SecurityConfig — CORS, rate limit, métricas e upload.
type SecurityConfig struct {
	CORSAllowedOrigins []string
	MetricsEnabled     bool
	MetricsBasicUser   string
	MetricsBasicPass   string
	AuthRatePerMin     int
	UploadRatePerMin   int
	MaxUploadBytes     int64
}

type StorageConfig struct {
	Provider          string
	Region            string
	Bucket            string
	Endpoint          string
	AccessKeyID       string
	SecretAccessKey   string
	UsePathStyle      bool
	UploadPartSizeMB  int64
	UploadConcurrency int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	chunk := 1000
	if v := os.Getenv("PAGE_CHUNK_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			chunk = n
		}
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	hours := 72
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			hours = n
		}
	}

	freeBooks := 2
	if v := os.Getenv("FREE_MAX_BOOKS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			freeBooks = n
		}
	}

	freeQ := 10
	if v := os.Getenv("FREE_MAX_QUESTIONS_PER_DAY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			freeQ = n
		}
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Em ambiente local/dev, não bloquear fluxo por limite freemium.
	if env != "production" {
		freeBooks = 0
		freeQ = 0
	}

	partSizeMB := int64(8)
	if v := os.Getenv("S3_UPLOAD_PART_SIZE_MB"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 5 {
			partSizeMB = n
		}
	}
	uploadConcurrency := 4
	if v := os.Getenv("S3_UPLOAD_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			uploadConcurrency = n
		}
	}
	usePathStyle := true
	if v := os.Getenv("S3_USE_PATH_STYLE"); v != "" {
		usePathStyle = v == "1" || v == "true" || v == "TRUE" || v == "True"
	}

	origins := parseCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(origins) == 0 && env == "development" {
		origins = []string{"http://localhost:3000"}
	}
	metricsEnabled := true
	if v := os.Getenv("METRICS_ENABLED"); v != "" {
		metricsEnabled = v == "1" || v == "true" || v == "TRUE" || v == "True"
	}
	authR := 30
	if v := os.Getenv("AUTH_RATE_LIMIT_PER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			authR = n
		}
	}
	uploadR := 20
	if v := os.Getenv("UPLOAD_RATE_LIMIT_PER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			uploadR = n
		}
	}
	maxUp := int64(50 * 1024 * 1024)
	if v := os.Getenv("MAX_UPLOAD_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			maxUp = n
		}
	}

	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AIAPIKey:      apiKey,
		AIModel:       model,
		Environment:   env,
		HTTPAddr:      addr,
		PageChunkSize: chunk,
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiry:     time.Duration(hours) * time.Hour,
		Limits: limiter.Config{
			FreeMaxBooks:        freeBooks,
			FreeMaxQuestionsDay: freeQ,
		},
		Storage: StorageConfig{
			Provider:          getenvDefault("STORAGE_PROVIDER", "s3"),
			Region:            getenvDefault("AWS_REGION", "us-east-1"),
			Bucket:            os.Getenv("S3_BUCKET"),
			Endpoint:          os.Getenv("S3_ENDPOINT"),
			AccessKeyID:       os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
			UsePathStyle:      usePathStyle,
			UploadPartSizeMB:  partSizeMB,
			UploadConcurrency: uploadConcurrency,
		},
		Security: SecurityConfig{
			CORSAllowedOrigins: origins,
			MetricsEnabled:     metricsEnabled,
			MetricsBasicUser:   os.Getenv("METRICS_BASIC_USER"),
			MetricsBasicPass:   os.Getenv("METRICS_BASIC_PASSWORD"),
			AuthRatePerMin:     authR,
			UploadRatePerMin:   uploadR,
			MaxUploadBytes:     maxUp,
		},
	}, nil
}

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
