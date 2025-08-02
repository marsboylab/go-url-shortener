package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-url-shortener/internal/domain"
	"go-url-shortener/internal/repository/interfaces"
)

// urlRepository는 PostgreSQL을 사용한 URL 저장소 구현입니다
type urlRepository struct {
	db *sql.DB
}

// NewURLRepository는 새로운 URL 저장소를 생성합니다
func NewURLRepository(db *sql.DB) interfaces.URLRepository {
	return &urlRepository{db: db}
}

// Create는 새로운 URL을 생성합니다
func (r *urlRepository) Create(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (id, original_url, description, expires_at, created_at, updated_at, 
						 click_count, is_active, created_by_api_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.OriginalURL,
		url.Description,
		url.ExpiresAt,
		url.CreatedAt,
		url.UpdatedAt,
		url.ClickCount,
		url.IsActive,
		url.CreatedByAPIKey,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("URL with ID '%s' already exists", url.ID)
		}
		return fmt.Errorf("failed to create URL: %w", err)
	}
	
	return nil
}

// GetByID는 ID로 URL을 조회합니다
func (r *urlRepository) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	query := `
		SELECT id, original_url, description, expires_at, created_at, updated_at,
			   click_count, is_active, last_accessed_at, created_by_api_key
		FROM urls 
		WHERE id = $1 AND is_active = true`
	
	url := &domain.URL{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Description,
		&url.ExpiresAt,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
		&url.IsActive,
		&url.LastAccessedAt,
		&url.CreatedByAPIKey,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL with ID '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	
	return url, nil
}

// Update는 URL을 업데이트합니다
func (r *urlRepository) Update(ctx context.Context, url *domain.URL) error {
	query := `
		UPDATE urls 
		SET original_url = $2, description = $3, expires_at = $4, updated_at = $5,
			click_count = $6, is_active = $7, last_accessed_at = $8
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.OriginalURL,
		url.Description,
		url.ExpiresAt,
		url.UpdatedAt,
		url.ClickCount,
		url.IsActive,
		url.LastAccessedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("URL with ID '%s' not found", url.ID)
	}
	
	return nil
}

// Delete는 URL을 삭제합니다 (soft delete)
func (r *urlRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE urls SET is_active = false, updated_at = $1 WHERE id = $2`
	
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("URL with ID '%s' not found", id)
	}
	
	return nil
}

// List는 URL 목록을 조회합니다
func (r *urlRepository) List(ctx context.Context, apiKey string, options domain.URLListOptions) ([]domain.URL, int64, error) {
	// 기본값 설정
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}
	if options.Sort == "" {
		options.Sort = "created_at"
	}
	if options.Order == "" {
		options.Order = "desc"
	}
	
	// WHERE 절 구성
	whereClause := "WHERE created_by_api_key = $1"
	args := []interface{}{apiKey}
	argIndex := 2
	
	if options.IsActive != nil {
		whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *options.IsActive)
		argIndex++
	}
	
	// 총 개수 조회
	countQuery := "SELECT COUNT(*) FROM urls " + whereClause
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count URLs: %w", err)
	}
	
	// 목록 조회
	offset := (options.Page - 1) * options.Limit
	query := fmt.Sprintf(`
		SELECT id, original_url, description, expires_at, created_at, updated_at,
			   click_count, is_active, last_accessed_at, created_by_api_key
		FROM urls 
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, options.Sort, options.Order, argIndex, argIndex+1)
	
	args = append(args, options.Limit, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer rows.Close()
	
	var urls []domain.URL
	for rows.Next() {
		var url domain.URL
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.Description,
			&url.ExpiresAt,
			&url.CreatedAt,
			&url.UpdatedAt,
			&url.ClickCount,
			&url.IsActive,
			&url.LastAccessedAt,
			&url.CreatedByAPIKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return urls, totalCount, nil
}

// ExistsByID는 ID가 이미 존재하는지 확인합니다
func (r *urlRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)"
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check URL existence: %w", err)
	}
	
	return exists, nil
}

// IncrementClickCount는 클릭 수를 증가시킵니다
func (r *urlRepository) IncrementClickCount(ctx context.Context, id string) error {
	query := `
		UPDATE urls 
		SET click_count = click_count + 1, 
			last_accessed_at = $1,
			updated_at = $1
		WHERE id = $2 AND is_active = true`
	
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("URL with ID '%s' not found or inactive", id)
	}
	
	return nil
}

// UpdateLastAccessed는 마지막 접근 시간을 업데이트합니다
func (r *urlRepository) UpdateLastAccessed(ctx context.Context, id string) error {
	query := `
		UPDATE urls 
		SET last_accessed_at = $1, updated_at = $1
		WHERE id = $2 AND is_active = true`
	
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last accessed: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("URL with ID '%s' not found or inactive", id)
	}
	
	return nil
}

// GetExpiredURLs는 만료된 URL 목록을 조회합니다
func (r *urlRepository) GetExpiredURLs(ctx context.Context, limit int) ([]domain.URL, error) {
	query := `
		SELECT id, original_url, description, expires_at, created_at, updated_at,
			   click_count, is_active, last_accessed_at, created_by_api_key
		FROM urls 
		WHERE expires_at < $1 AND is_active = true
		ORDER BY expires_at ASC
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired URLs: %w", err)
	}
	defer rows.Close()
	
	var urls []domain.URL
	for rows.Next() {
		var url domain.URL
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.Description,
			&url.ExpiresAt,
			&url.CreatedAt,
			&url.UpdatedAt,
			&url.ClickCount,
			&url.IsActive,
			&url.LastAccessedAt,
			&url.CreatedByAPIKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired URL: %w", err)
		}
		urls = append(urls, url)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return urls, nil
}

// DeleteExpiredURLs는 만료된 URL들을 삭제합니다
func (r *urlRepository) DeleteExpiredURLs(ctx context.Context, before time.Time) (int64, error) {
	query := `UPDATE urls SET is_active = false, updated_at = $1 WHERE expires_at < $2 AND is_active = true`
	
	result, err := r.db.ExecContext(ctx, query, time.Now(), before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired URLs: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return rowsAffected, nil
}