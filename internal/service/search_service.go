package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
)

// SearchService handles search operations using Sphinx
type SearchService struct {
	searchRepo    *repository.SearchRepository
	filesRepo     repository.FilesRepository
	categoryRepo  repository.CategoryRepository
	sphinxDB      *sql.DB
	mainIndex     string
	deltaIndex    string
}

// NewSearchService creates a new search service
func NewSearchService(
	sphinxDB *sql.DB,
	mainIndex string,
	deltaIndex string,
	filesRepo repository.FilesRepository,
	categoryRepo repository.CategoryRepository,
) *SearchService {
	var searchRepo *repository.SearchRepository
	if sphinxDB != nil {
		searchRepo = repository.NewSearchRepository(sphinxDB, mainIndex, deltaIndex)
	}
	
	return &SearchService{
		searchRepo:   searchRepo,
		filesRepo:    filesRepo,
		categoryRepo: categoryRepo,
		sphinxDB:     sphinxDB,
		mainIndex:    mainIndex,
		deltaIndex:   deltaIndex,
	}
}

// SearchResult represents a search result with highlighting
type SearchResult struct {
	ID           uint64            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Type         int               `json:"type"`
	TypeName     string            `json:"type_name"`
	CategoryID   uint              `json:"category_id"`
	CategoryName string            `json:"category_name"`
	CategoryPath string            `json:"category_path"`
	Status       int               `json:"status"`
	StatusName   string            `json:"status_name"`
	Level        int               `json:"level"`
	UploadAt     string            `json:"upload_at"`
	PutoutAt     string            `json:"putout_at"`
	Snippet      string            `json:"snippet"`
	Relevance    int               `json:"relevance"`
	ThumbnailURL string            `json:"thumbnail_url"`
	PreviewURL   string            `json:"preview_url"`
	Size         int64             `json:"size"`
	Ext          string            `json:"ext"`
	CatalogInfo  map[string]string `json:"catalog_info"`
}

// SearchParams represents search request parameters
type SearchParams struct {
	Query      string
	Type       []int
	CategoryID uint
	Status     []int
	Level      int
	GroupID    uint
	DateFrom   string
	DateTo     string
	Page       int
	PageSize   int
	SortBy     string
}

// Search performs full-text search using Sphinx
func (s *SearchService) Search(ctx context.Context, params SearchParams) ([]SearchResult, int64, map[string]map[interface{}]int, error) {
	// If Sphinx is not configured, fallback to database search
	if s.searchRepo == nil {
		return s.fallbackSearch(ctx, params)
	}

	// Parse date filters
	var dateFrom, dateTo time.Time
	var err error
	if params.DateFrom != "" {
		dateFrom, err = time.Parse("2006-01-02", params.DateFrom)
		if err != nil {
			dateFrom = time.Time{}
		}
	}
	if params.DateTo != "" {
		dateTo, err = time.Parse("2006-01-02", params.DateTo)
		if err != nil {
			dateTo = time.Time{}
		}
	}

	// Build repository search params
	// IMPORTANT: 搜索只能搜索已发布的内容，强制status=2
	repoParams := repository.SearchParams{
		Query:      params.Query,
		Type:       params.Type,
		CategoryID: params.CategoryID,
		Status:     []int{2}, // 强制只搜索已发布内容 (status=2)
		Level:      params.Level,
		GroupID:    params.GroupID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Page:       params.Page,
		PageSize:   params.PageSize,
		SortBy:     params.SortBy,
	}

	// Execute search
	rows, err := s.searchRepo.ExecuteSearch(ctx, repoParams)
	if err != nil {
		// Fallback to database search on error
		return s.fallbackSearch(ctx, params)
	}

	// Get total count
	total, err := s.searchRepo.GetTotalCount(ctx, repoParams)
	if err != nil {
		total = int64(len(rows))
	}

	// Get facets
	facets, _ := s.searchRepo.GetFacets(ctx, repoParams)

	// Enrich results with additional data from database
	results := make([]SearchResult, 0, len(rows))
	for _, row := range rows {
		result := SearchResult{
			ID:         row.ID,
			Title:      row.Title,
			Type:       row.Type,
			CategoryID: row.CategoryID,
			Status:     row.Status,
			Level:      row.Level,
			Relevance:  row.Relevance,
		}

		// Format timestamps
		if row.PutoutAt > 0 {
			result.PutoutAt = time.Unix(row.PutoutAt, 0).Format("2006-01-02 15:04:05")
		}

		// Get full file details from database
		if file, err := s.filesRepo.FindByID(ctx, row.ID); err == nil {
			result.Description = file.Title // Or extract from catalog_info
			result.Size = file.Size
			result.Ext = file.Ext
			result.UploadAt = time.Unix(int64(file.UploadAt), 0).Format("2006-01-02 15:04:05")
			result.ThumbnailURL = s.getThumbnailURL(file)
			result.PreviewURL = s.getPreviewURL(file)

			// Parse catalog_info JSON
			if file.CatalogInfo != "" {
				var catalogMap map[string]string
				if err := json.Unmarshal([]byte(file.CatalogInfo), &catalogMap); err == nil {
					result.CatalogInfo = catalogMap
					// Extract description if available
					if desc, ok := catalogMap["description"]; ok {
						result.Description = desc
					}
				}
			}
		}

		// Get category info
		if category, err := s.categoryRepo.FindByID(ctx, int(row.CategoryID)); err == nil {
			result.CategoryName = category.Name
			result.CategoryPath = category.Path
		}

		// Add type name
		result.TypeName = s.getTypeName(row.Type)

		// Add status name
		result.StatusName = s.getStatusName(row.Status)

		// Generate snippet with highlighting
		if params.Query != "" {
			textForSnippet := result.Title
			if result.Description != "" {
				textForSnippet = result.Title + " " + result.Description
			}
			snippet, _ := s.searchRepo.GenerateSnippet(ctx, textForSnippet, params.Query, "")
			result.Snippet = snippet
		} else {
			// Return description as snippet if no query
			if len(result.Description) > 200 {
				result.Snippet = result.Description[:200] + "..."
			} else {
				result.Snippet = result.Description
			}
		}

		results = append(results, result)
	}

	return results, total, facets, nil
}

// fallbackSearch performs database search when Sphinx is unavailable
func (s *SearchService) fallbackSearch(ctx context.Context, params SearchParams) ([]SearchResult, int64, map[string]map[interface{}]int, error) {
	// Use database repository to perform basic search
	// This is a fallback with limited functionality
	
	var results []SearchResult
	facets := make(map[string]map[interface{}]int)
	
	// Build filters map for FindAll
	filters := make(map[string]interface{})
	
	// IMPORTANT: 搜索只能搜索已发布的内容 (status=2)
	filters["status"] = 2
	
	// Add search query filter (IMPORTANT: This was missing!)
	if params.Query != "" {
		filters["search_query"] = params.Query
	}
	
	if len(params.Type) > 0 && params.Type[0] > 0 {
		filters["type"] = params.Type[0]
	}
	// 注意：不再使用params.Status，强制只搜索已发布内容
	if params.CategoryID > 0 {
		filters["category_id"] = params.CategoryID
	}
	if params.Level > 0 {
		filters["level"] = params.Level
	}
	if params.DateFrom != "" {
		filters["upload_date_from"] = params.DateFrom
	}
	if params.DateTo != "" {
		filters["upload_date_to"] = params.DateTo
	}
	
	// Calculate offset
	offset := (params.Page - 1) * params.PageSize
	if offset < 0 {
		offset = 0
	}
	
	// Build basic query using files repository
	files, total, err := s.filesRepo.FindAll(ctx, filters, params.PageSize, offset)
	if err != nil {
		return results, 0, facets, fmt.Errorf("fallback search failed: %w", err)
	}
	
	// Convert to search results
	for _, file := range files {
		result := SearchResult{
			ID:           uint64(file.ID),
			Title:        file.Title,
			Type:         file.Type,
			TypeName:     s.getTypeName(file.Type),
			CategoryID:   uint(file.CategoryID),
			CategoryName: file.CategoryName,
			Status:       file.Status,
			StatusName:   s.getStatusName(file.Status),
			Level:        file.Level,
			UploadAt:     time.Unix(int64(file.UploadAt), 0).Format("2006-01-02 15:04:05"),
			Size:         file.Size,
			Ext:          file.Ext,
			ThumbnailURL: s.getThumbnailURL(file),
			PreviewURL:   s.getPreviewURL(file),
		}
		
		if file.PutoutAt != nil {
			result.PutoutAt = time.Unix(int64(*file.PutoutAt), 0).Format("2006-01-02 15:04:05")
		}
		
		// Parse catalog info
		if file.CatalogInfo != "" {
			var catalogMap map[string]string
			if err := json.Unmarshal([]byte(file.CatalogInfo), &catalogMap); err == nil {
				result.CatalogInfo = catalogMap
				if desc, ok := catalogMap["description"]; ok {
					result.Description = desc
				}
			}
		}
		
		// Generate simple snippet highlighting if query is present
		if params.Query != "" {
			snippet := s.generateSimpleSnippet(result.Title, result.Description, params.Query, 200)
			result.Snippet = snippet
		} else {
			// Return description as snippet if no query
			if len(result.Description) > 200 {
				result.Snippet = result.Description[:200] + "..."
			} else {
				result.Snippet = result.Description
			}
		}
		
		results = append(results, result)
	}
	
	return results, total, facets, nil
}

// generateSimpleSnippet creates a highlighted snippet without Sphinx
func (s *SearchService) generateSimpleSnippet(title, description, query string, maxLen int) string {
	// Combine title and description
	text := title
	if description != "" {
		text = title + " " + description
	}
	
	// Simple case-insensitive search for query term
	// In production, you'd want more sophisticated highlighting
	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)
	
	// Find position of query in text
	pos := strings.Index(textLower, queryLower)
	
	if pos >= 0 {
		// Extract context around the match
		start := pos - 50
		if start < 0 {
			start = 0
		}
		
		end := pos + len(query) + 150
		if end > len(text) {
			end = len(text)
		}
		
		snippet := text[start:end]
		
		// Add ellipsis if truncated
		if start > 0 {
			snippet = "..." + snippet
		}
		if end < len(text) {
			snippet = snippet + "..."
		}
		
		// Highlight the query term with <b> tags
		// Case-insensitive replacement
		snippetLower := strings.ToLower(snippet)
		pos = strings.Index(snippetLower, queryLower)
		if pos >= 0 {
			highlighted := snippet[:pos] + "<b>" + snippet[pos:pos+len(query)] + "</b>" + snippet[pos+len(query):]
			return highlighted
		}
		
		return snippet
	}
	
	// Query not found, return beginning of text
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}

// Helper methods

func (s *SearchService) getTypeName(fileType int) string {
	switch fileType {
	case 1:
		return "Video"
	case 2:
		return "Audio"
	case 3:
		return "Image"
	case 4:
		return "Rich Media"
	default:
		return "Unknown"
	}
}

func (s *SearchService) getStatusName(status int) string {
	switch status {
	case 0:
		return "New"
	case 1:
		return "Pending"
	case 2:
		return "Published"
	case 3:
		return "Rejected"
	case 4:
		return "Deleted"
	default:
		return "Unknown"
	}
}

func (s *SearchService) getThumbnailURL(file *models.Files) string {
	if file == nil {
		return ""
	}
	
	// For images, use the image itself as thumbnail
	if file.Type == 3 {
		return fmt.Sprintf("/api/v1/files/%d/download", file.ID)
	}
	
	// For videos and audio, return a default thumbnail
	// In production, you might have actual thumbnail files generated
	switch file.Type {
	case 1:
		return "/static/thumbnails/video-default.png"
	case 2:
		return "/static/thumbnails/audio-default.png"
	case 4:
		return "/static/thumbnails/doc-default.png"
	default:
		return "/static/thumbnails/file-default.png"
	}
}

func (s *SearchService) getPreviewURL(file *models.Files) string {
	if file == nil {
		return ""
	}
	
	// For video and audio files, return the preview URL
	if file.Type == 1 || file.Type == 2 {
		return fmt.Sprintf("/api/v1/files/%d/preview", file.ID)
	}
	
	// For images, return the download URL
	if file.Type == 3 {
		return fmt.Sprintf("/api/v1/files/%d/download", file.ID)
	}
	
	return ""
}

// Reindex triggers search index rebuild
func (s *SearchService) Reindex(ctx context.Context) error {
	if s.sphinxDB == nil {
		return fmt.Errorf("Sphinx not configured")
	}
	
	// In production, this would trigger indexer command
	// For now, just return success
	// Execute: indexer --config sphinx.conf --all --rotate
	
	return fmt.Errorf("Reindex functionality requires indexer command execution - not implemented in this version")
}

// GetIndexStatus returns indexing status
func (s *SearchService) GetIndexStatus(ctx context.Context) (map[string]interface{}, error) {
	if s.sphinxDB == nil {
		return map[string]interface{}{
			"configured": false,
			"status": "Sphinx not configured",
		}, nil
	}
	
	// Check if we can connect to Sphinx
	err := s.sphinxDB.PingContext(ctx)
	if err != nil {
		return map[string]interface{}{
			"configured": true,
			"status": "Sphinx unavailable",
			"error": err.Error(),
		}, nil
	}
	
	// Get index stats using SHOW INDEX STATUS
	rows, err := s.sphinxDB.QueryContext(ctx, fmt.Sprintf("SHOW INDEX %s STATUS", s.mainIndex))
	if err != nil {
		return map[string]interface{}{
			"configured": true,
			"status": "Connected",
			"index_status": "Unknown",
		}, nil
	}
	defer rows.Close()
	
	stats := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err == nil {
			stats[name] = value
		}
	}
	
	return map[string]interface{}{
		"configured": true,
		"status": "Connected",
		"index": s.mainIndex,
		"stats": stats,
	}, nil
}


// GetSuggestions returns search suggestions for autocomplete
// Returns popular search terms or recent file titles matching the query
func (s *SearchService) GetSuggestions(ctx context.Context, query string) ([]string, error) {
	// Simple implementation: return file titles that match the query
	// In production, this could be enhanced with:
	// - Popular search terms from search history
	// - Recent searches by user
	// - Weighted suggestions based on file popularity
	
	var suggestions []string
	
	// Use filesRepo to search for matching titles
	// Build filters for title search
	filters := map[string]interface{}{
		"title":  query,  // Will be used as LIKE %query%
		"status": 2,      // Only published files
	}
	
	files, _, err := s.filesRepo.FindAll(ctx, filters, 10, 0)
	if err != nil {
		return suggestions, err
	}
	
	for _, file := range files {
		if file.Title != "" {
			suggestions = append(suggestions, file.Title)
		}
	}
	
	return suggestions, nil
}

