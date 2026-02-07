package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openwan/media-asset-management/internal/api"
	"github.com/openwan/media-asset-management/internal/database"
	"github.com/openwan/media-asset-management/internal/repository"
	"github.com/openwan/media-asset-management/internal/service"
	"github.com/openwan/media-asset-management/internal/session"
	"github.com/openwan/media-asset-management/internal/storage"
)

func main() {
	fmt.Println("OpenWan Media Asset Management System - API Server")
	fmt.Println("Version: 1.0.0 with Session Support")
	fmt.Println("Go Version: 1.25.5")

	// Initialize database connection
	dbConfig := database.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "openwan_user",
		Password: "openwan_password",
		Database: "openwan_db",
		Prefix:   "ow_",
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)
	fmt.Println("✓ Database connected")

	// Initialize Redis session store
	sessionStore, err := session.NewRedisStore(
		"localhost:6379", // Redis address
		"",               // No password
		0,                // DB 0
		24*time.Hour,     // TTL 24 hours
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis for sessions: %v", err)
		log.Println("Session management will not work properly!")
		// Continue anyway for testing
	} else {
		fmt.Println("✓ Redis session store connected")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	aclRepo := repository.NewACLRepository(db)
	fileRepo := repository.NewFileRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	searchRepo := repository.NewSearchRepository(db)

	// Initialize storage service (local for now)
	storageService := storage.NewLocalStorage("/tmp/openwan-storage")
	fmt.Println("✓ Storage service initialized")

	// Initialize services
	usersService := service.NewUsersService(userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	roleService := service.NewRoleService(roleRepo, permissionRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	aclService := service.NewACLService(aclRepo, userRepo, groupRepo, roleRepo, permissionRepo)
	fileService := service.NewFileService(fileRepo, categoryRepo, storageService)
	categoryService := service.NewCategoryService(categoryRepo)
	catalogService := service.NewCatalogService(catalogRepo)
	searchService := service.NewSearchService(searchRepo, aclService)

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
		StorageService:    storageService,
	}

	// Setup router
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://13.217.210.142",
		"http://13.217.210.142:3000",
	}
	router := api.SetupRouter(allowedOrigins, deps)
	fmt.Println("✓ Router configured")

	// Create and start server
	server := api.NewServer(router, ":8080")
	
	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		fmt.Println("\nShutting down server...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
		
		if sessionStore != nil {
			sessionStore.Close()
		}
		
		fmt.Println("Server stopped")
		os.Exit(0)
	}()

	fmt.Println("Starting server on :8080")
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
