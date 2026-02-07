package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SearchRepository handles Sphinx search queries using SphinxQL
type SearchRepository struct {
	sphinxDB   *sql.DB
	mainIndex  string
	deltaIndex string
}

// NewSearchRepository creates a new search repository
func NewSearchRepository(sphinxDB *sql.DB, mainIndex, deltaIndex string) *SearchRepository {
	return &SearchRepository{
		sphinxDB:   sphinxDB,
		mainIndex:  mainIndex,
		deltaIndex: deltaIndex,
	}
}

// SearchParams represents search parameters
type SearchParams struct {
	Query      string
	Type       []int
	CategoryID uint
	Status     []int
	Level      int
	GroupID    uint
	DateFrom   time.Time
	DateTo     time.Time
	Page       int
	PageSize   int
	SortBy     string // relevance, date, title
}

// SearchResultRow represents a raw search result row from Sphinx
type SearchResultRow struct {
	ID          uint64
	Title       string
	CategoryID  uint
	Type        int
	Status      int
	Level       int
	PutoutAt    int64
	Relevance   int
	CatalogInfo string
	Groups      string
}

// ExecuteSearch performs SphinxQL search query
func (r *SearchRepository) ExecuteSearch(ctx context.Context, params SearchParams) ([]SearchResultRow, error) {
	// Build SphinxQL query
	query := r.buildSearchQuery(params)
	
	// Execute query
	rows, err := r.sphinxDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sphinx query failed: %w", err)
	}
	defer rows.Close()
	
	var results []SearchResultRow
	for rows.Next() {
		var result SearchResultRow
		var putoutAt sql.NullInt64
		
		err := rows.Scan(
			&result.ID,
			&result.Title,
			&result.CategoryID,
			&result.Type,
			&result.Status,
			&result.Level,
			&putoutAt,
			&result.Relevance,
		)
		if err != nil {
			continue
		}
		
		if putoutAt.Valid {
			result.PutoutAt = putoutAt.Int64
		}
		
		results = append(results, result)
	}
	
	return results, rows.Err()
}

// GetTotalCount gets total count of search results
func (r *SearchRepository) GetTotalCount(ctx context.Context, params SearchParams) (int64, error) {
	// Build count query
	query := r.buildCountQuery(params)
	
	var total int64
	err := r.sphinxDB.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}
	
	return total, nil
}

// GenerateSnippet generates highlighted snippet using Sphinx SNIPPET() function
func (r *SearchRepository) GenerateSnippet(ctx context.Context, text, query, options string) (string, error) {
	if query == "" {
		return text, nil
	}
	
	// Escape single quotes
	text = strings.ReplaceAll(text, "'", "\\'")
	query = strings.ReplaceAll(query, "'", "\\'")
	
	// Default options if not provided
	if options == "" {
		options = "'before_match' = '<mark>', 'after_match' = '</mark>', 'limit' = 256"
	}
	
	// Build SNIPPET query
	snippetQuery := fmt.Sprintf(
		"CALL SNIPPETS(('%s'), '%s', '%s', %s)",
		text,
		r.mainIndex,
		query,
		options,
	)
	
	var snippet string
	err := r.sphinxDB.QueryRowContext(ctx, snippetQuery).Scan(&snippet)
	if err != nil {
		// Fallback to original text if snippet generation fails
		return text, nil
	}
	
	return snippet, nil
}

// CheckConnection checks if Sphinx is available
func (r *SearchRepository) CheckConnection(ctx context.Context) error {
	return r.sphinxDB.PingContext(ctx)
}

// GetIndexStatus retrieves index status information
func (r *SearchRepository) GetIndexStatus(ctx context.Context) (map[string]string, error) {
	query := fmt.Sprintf("SHOW INDEX %s STATUS", r.mainIndex)
	
	rows, err := r.sphinxDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get index status: %w", err)
	}
	defer rows.Close()
	
	stats := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err == nil {
			stats[name] = value
		}
	}
	
	return stats, rows.Err()
}

// buildSearchQuery builds the SphinxQL SELECT query
func (r *SearchRepository) buildSearchQuery(params SearchParams) string {
	var conditions []string
	
	// Add MATCH clause for full-text search
	if params.Query != "" {
		// Escape query
		query := r.escapeQuery(params.Query)
		conditions = append(conditions, fmt.Sprintf("MATCH('%s')", query))
	}
	
	// Add filter conditions
	if len(params.Type) > 0 {
		typeList := make([]string, len(params.Type))
		for i, t := range params.Type {
			typeList[i] = fmt.Sprintf("%d", t)
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typeList, ",")))
	}
	
	if params.CategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = %d", params.CategoryID))
	}
	
	if len(params.Status) > 0 {
		statusList := make([]string, len(params.Status))
		for i, s := range params.Status {
			statusList[i] = fmt.Sprintf("%d", s)
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusList, ",")))
	} else {
		// Default to only published files (status=2)
		conditions = append(conditions, "status = 2")
	}
	
	if params.Level > 0 {
		conditions = append(conditions, fmt.Sprintf("level <= %d", params.Level))
	}
	
	// Date range filtering
	if !params.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("putout_at >= %d", params.DateFrom.Unix()))
	}
	if !params.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("putout_at <= %d", params.DateTo.Unix()))
	}
	
	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// Build ORDER BY clause
	orderBy := r.buildOrderByClause(params.SortBy)
	
	// Build LIMIT clause
	offset := (params.Page - 1) * params.PageSize
	limitClause := fmt.Sprintf("LIMIT %d, %d", offset, params.PageSize)
	
	// Construct final query
	query := fmt.Sprintf(
		"SELECT id, title, category_id, type, status, level, putout_at, WEIGHT() as relevance FROM %s %s %s %s",
		r.mainIndex,
		whereClause,
		orderBy,
		limitClause,
	)
	
	return query
}

// buildCountQuery builds the SphinxQL COUNT query
func (r *SearchRepository) buildCountQuery(params SearchParams) string {
	var conditions []string
	
	// Add MATCH clause
	if params.Query != "" {
		query := r.escapeQuery(params.Query)
		conditions = append(conditions, fmt.Sprintf("MATCH('%s')", query))
	}
	
	// Add filter conditions (same as search query)
	if len(params.Type) > 0 {
		typeList := make([]string, len(params.Type))
		for i, t := range params.Type {
			typeList[i] = fmt.Sprintf("%d", t)
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typeList, ",")))
	}
	
	if params.CategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = %d", params.CategoryID))
	}
	
	if len(params.Status) > 0 {
		statusList := make([]string, len(params.Status))
		for i, s := range params.Status {
			statusList[i] = fmt.Sprintf("%d", s)
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusList, ",")))
	} else {
		conditions = append(conditions, "status = 2")
	}
	
	if params.Level > 0 {
		conditions = append(conditions, fmt.Sprintf("level <= %d", params.Level))
	}
	
	if !params.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("putout_at >= %d", params.DateFrom.Unix()))
	}
	if !params.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("putout_at <= %d", params.DateTo.Unix()))
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	return fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.mainIndex, whereClause)
}

// buildOrderByClause builds the ORDER BY clause
func (r *SearchRepository) buildOrderByClause(sortBy string) string {
	switch sortBy {
	case "date":
		return "ORDER BY putout_at DESC"
	case "title":
		return "ORDER BY title ASC"
	case "relevance":
		fallthrough
	default:
		return "ORDER BY WEIGHT() DESC, putout_at DESC"
	}
}

// escapeQuery escapes special characters in search query
func (r *SearchRepository) escapeQuery(query string) string {
	// Escape single quotes
	query = strings.ReplaceAll(query, "'", "\\'")
	// Escape backslashes
	query = strings.ReplaceAll(query, "\\", "\\\\")
	return query
}

// GetFacets returns faceted search results (counts by category, type, etc.)
func (r *SearchRepository) GetFacets(ctx context.Context, params SearchParams) (map[string]map[interface{}]int, error) {
	facets := make(map[string]map[interface{}]int)
	
	// Get type facets
	typeFacets, err := r.getFacetCounts(ctx, params, "type")
	if err == nil {
		facets["type"] = typeFacets
	}
	
	// Get category facets
	categoryFacets, err := r.getFacetCounts(ctx, params, "category_id")
	if err == nil {
		facets["category"] = categoryFacets
	}
	
	return facets, nil
}

// getFacetCounts gets count for a specific facet field
func (r *SearchRepository) getFacetCounts(ctx context.Context, params SearchParams, field string) (map[interface{}]int, error) {
	// Build base query without the facet field filter
	var conditions []string
	
	if params.Query != "" {
		query := r.escapeQuery(params.Query)
		conditions = append(conditions, fmt.Sprintf("MATCH('%s')", query))
	}
	
	// Add status filter
	if len(params.Status) > 0 {
		statusList := make([]string, len(params.Status))
		for i, s := range params.Status {
			statusList[i] = fmt.Sprintf("%d", s)
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusList, ",")))
	} else {
		conditions = append(conditions, "status = 2")
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// Build facet query with GROUP BY
	query := fmt.Sprintf(
		"SELECT %s, COUNT(*) as cnt FROM %s %s GROUP BY %s ORDER BY cnt DESC LIMIT 100",
		field,
		r.mainIndex,
		whereClause,
		field,
	)
	
	rows, err := r.sphinxDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	facetCounts := make(map[interface{}]int)
	for rows.Next() {
		var fieldValue int
		var count int
		if err := rows.Scan(&fieldValue, &count); err == nil {
			facetCounts[fieldValue] = count
		}
	}
	
	return facetCounts, rows.Err()
}
