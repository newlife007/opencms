package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm/logger"
	
	"github.com/openwan/media-asset-management/internal/api"
	"github.com/openwan/media-asset-management/internal/database"
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

	// Initialize database
	fmt.Println("Initializing database connection...")
	
	// Get database configuration from environment variables or use defaults
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "openwan"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "openwan123"
	}
	
	dbConfig := database.Config{
		Host:            dbHost,
		Port:            3306,
		Database:        "openwan_db",
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
	storageConfig := storage.LoadConfigFromEnv()
	storageService, err := storage.NewStorageFromConfig(storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	fmt.Println("✓ Storage service initialized")

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

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\n\nShutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		if sessionStore != nil {
			sessionStore.Close()
		}

		database.Close()
		fmt.Println("Server stopped gracefully")
		os.Exit(0)
	}()

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
	
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
