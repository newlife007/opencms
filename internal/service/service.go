package service

import (
	"github.com/openwan/media-asset-management/internal/repository"
)

// Services holds all service instances
type Services struct {
	Files *FilesService
	ACL   *ACLService
}

// NewServices creates all services with dependency injection
func NewServices(repo repository.Repository) *Services {
	return &Services{
		Files: NewFilesService(repo),
		ACL:   NewACLService(repo),
	}
}
