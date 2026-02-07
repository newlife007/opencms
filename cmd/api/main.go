package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"
	
	"github.com/openwan/media-asset-management/internal/api"
	"github.com/openwan/media-asset-management/internal/config"
	"github.com/openwan/media-asset-management/internal/database"
	"github.com/openwan/media-asset-management/internal/queue"
	"github.com/openwan/media-asset-management/internal/repository"
	"github.com/openwan/media-asset-management/internal/service"
	"github.com/openwan/media-asset-management/internal/session"
	"github.com/openwan/media-asset-management/internal/storage"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("OpenWan Media Asset Management System")
	fmt.Println("API Server with Session Authentication")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Go Version: 1.25.5")
	fmt.Println("========================================")
	fmt.Println()

	// Load configuration from YAML file
	fmt.Println("Loading configuration...")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v", err)
		log.Println("Using environment variables and defaults...")
		cfg = nil // Will use defaults below
	} else {
		fmt.Printf("✓ Configuration loaded from %s\n", configPath)
	}

	// Initialize database
	fmt.Println("Initializing database connection...")
	
	// Get database configuration from config file or environment variables
	var dbHost, dbUser, dbPassword, dbName string
	var dbPort int
	
	if cfg != nil {
		dbHost = cfg.Database.Host
		dbPort = cfg.Database.Port
		dbName = cfg.Database.Database
		dbUser = cfg.Database.Username
		dbPassword = cfg.Database.Password
	} else {
		dbHost = os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "127.0.0.1"
		}
		dbPort = 3306
		dbName = "openwan_db"
		dbUser = os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "openwan"
		}
		dbPassword = os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			dbPassword = "openwan123"
		}
	}
	
	dbConfig := database.Config{
		Host:            dbHost,
		Port:            dbPort,
		Database:        dbName,
		Username:        dbUser,
		Password:        dbPassword,
		Prefix:          "ow_",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        logger.Warn,
	}
	
	if err := database.Initialize(dbConfig); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	db := database.GetDB()
	fmt.Println("✓ Database connected")

	// Initialize Redis session store
	fmt.Println("Initializing Redis session store...")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	sessionStore, err := session.NewRedisStore(
		redisAddr,
		"",
		0,
		24*time.Hour,
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Session management will not work properly!")
		sessionStore = nil
	} else {
		fmt.Println("✓ Redis session store connected")
	}

	// Initialize storage service
	fmt.Println("Initializing storage service...")
	// Load storage config from YAML config file or environment variables
	var storageConfig storage.Config
	
	if cfg != nil {
		storageConfig = storage.Config{
			Type:         cfg.Storage.Type,
			LocalPath:    cfg.Storage.LocalPath,
			S3Bucket:     cfg.Storage.S3Bucket,
			S3Region:     cfg.Storage.S3Region,
			S3AccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),     // From env for security
			S3SecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"), // From env for security
			S3Prefix:     cfg.Storage.S3Prefix,
			S3UseIAMRole: os.Getenv("S3_USE_IAM_ROLE") == "true",
		}
	} else {
		// Fallback to environment variables
		storageConfig = storage.LoadConfigFromEnv()
	}
	
	// Override with environment variables if set
	if envType := os.Getenv("STORAGE_TYPE"); envType != "" {
		storageConfig.Type = envType
	}
	if envBucket := os.Getenv("S3_BUCKET"); envBucket != "" {
		storageConfig.S3Bucket = envBucket
	}
	
	storageService, err := storage.NewStorageFromConfig(storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	fmt.Printf("✓ Storage service initialized (Type: %s)\n", storageConfig.Type)
	if storageConfig.Type == "s3" {
		fmt.Printf("  S3 Bucket: %s\n", storageConfig.S3Bucket)
		fmt.Printf("  S3 Region: %s\n", storageConfig.S3Region)
		fmt.Printf("  S3 Prefix: %s\n", storageConfig.S3Prefix)
		fmt.Printf("  Using IAM Role: %v\n", storageConfig.S3UseIAMRole)
	} else {
		fmt.Printf("  Local Path: %s\n", storageConfig.LocalPath)
	}

	// Initialize repositories
	fmt.Println("Initializing repositories...")
	mainRepo := repository.NewRepository(db)
	usersRepo := repository.NewUsersRepository(db)
	filesRepo := repository.NewFilesRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	groupsRepo := repository.NewGroupsRepository(db)
	levelsRepo := repository.NewLevelsRepository(db)
	fmt.Println("✓ Repositories initialized")

	// Initialize services
	fmt.Println("Initializing services...")
	aclService := service.NewACLService(mainRepo)
	usersService := service.NewUsersService(usersRepo, groupsRepo, levelsRepo)
	fileService := service.NewFileService(mainRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	catalogService := service.NewCatalogService(db)
	searchService := service.NewSearchService(nil, "", "", filesRepo, categoryRepo)
	groupService := service.NewGroupService(mainRepo)
	roleService := service.NewRoleService(mainRepo)
	permissionService := service.NewPermissionService(mainRepo)
	levelsService := service.NewLevelsService(levelsRepo)
	fmt.Println("✓ Services initialized")

	// Initialize queue service for transcoding
	fmt.Println("Initializing message queue...")
	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		queueURL = "amqp://guest:guest@localhost:5672/"
	}
	queueService, err := queue.NewRabbitMQQueue(queueURL)
	if err != nil {
		log.Printf("⚠ Warning: Failed to initialize queue service: %v", err)
		log.Printf("⚠ Transcoding will be disabled")
		queueService = nil
	} else {
		fmt.Println("✓ Queue service initialized")
	}

	// Setup router dependencies
	deps := &api.RouterDependencies{
		SessionStore:      sessionStore,
		ACLService:        aclService,
		UsersService:      usersService,
		FileService:       fileService,
		CategoryService:   categoryService,
		CatalogService:    catalogService,
		SearchService:     searchService,
		GroupService:      groupService,
		RoleService:       roleService,
		PermissionService: permissionService,
		LevelsService:     levelsService,
		StorageService:    storageService,
		QueueService:      queueService,
	}

	// Setup router
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://13.217.210.142",
		"http://13.217.210.142:3000",
	}
	
	// Add additional allowed origins from environment variable
	if extraOrigins := os.Getenv("ALLOWED_ORIGINS"); extraOrigins != "" {
		// Can be comma-separated list
		allowedOrigins = append(allowedOrigins, extraOrigins)
	}
	
	router := api.SetupRouter(allowedOrigins, deps)
	fmt.Println("✓ Router configured")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Create server
	server := api.NewServer(router, ":"+port)

	// Start server
	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("Server starting on :%s\n", port)
	fmt.Printf("Health check: http://localhost:%s/health\n", port)
	fmt.Printf("API endpoint: http://localhost:%s/api/v1/ping\n", port)
	fmt.Printf("Database: %s@%s:3306/openwan_db\n", dbUser, dbHost)
	fmt.Printf("Redis: %s\n", redisAddr)
	fmt.Printf("Storage: %s\n", storageConfig.Type)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("========================================")
	fmt.Println()
	
	// server.Start() handles signal listening and graceful shutdown internally
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		// Cleanup on error
		if sessionStore != nil {
			sessionStore.Close()
		}
		database.Close()
		os.Exit(1)
	}
	
	// Cleanup after server stops
	if sessionStore != nil {
		sessionStore.Close()
	}
	database.Close()
	fmt.Println("Server exited")
}
