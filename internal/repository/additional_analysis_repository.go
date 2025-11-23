package repository

import (
	"accounting-web/internal/models"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type AdditionalAnalysisRepository struct {
	db *sqlx.DB
}

func NewAdditionalAnalysisRepository(db *sqlx.DB) *AdditionalAnalysisRepository {
	return &AdditionalAnalysisRepository{db: db}
}

// Create creates a new additional analysis record
func (r *AdditionalAnalysisRepository) Create(analysis *models.AdditionalAnalysis) error {
	query := `
		INSERT INTO additional_analyses (
			account_code, analysis_type, analysis_title, status, notes, created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	createdAt := time.Now()
	updatedAt := createdAt

	result, err := r.db.Exec(query,
		analysis.AccountCode, analysis.AnalysisType, analysis.AnalysisTitle, analysis.Status,
		analysis.Notes, analysis.CreatedBy, createdAt, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create additional analysis: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	analysis.ID = int(id)
	analysis.CreatedAt = createdAt
	analysis.UpdatedAt = updatedAt

	return nil
}

// GetByID retrieves an additional analysis by ID with account information
func (r *AdditionalAnalysisRepository) GetByID(id int) (*models.AdditionalAnalysisResponse, error) {
	query := `
		SELECT
			aa.id, aa.account_code, a.account_name,
			aa.analysis_type, aa.analysis_title, aa.status, aa.notes, aa.created_by,
			aa.created_at, aa.updated_at
		FROM additional_analyses aa
		LEFT JOIN accounts a ON aa.account_code = a.account_code
		WHERE aa.id = ?
	`

	var analysis models.AdditionalAnalysisResponse
	err := r.db.QueryRowx(query, id).StructScan(&analysis)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("additional analysis not found")
		}
		return nil, fmt.Errorf("failed to get additional analysis: %w", err)
	}

	return &analysis, nil
}

// GetAll retrieves all additional analyses with filtering and pagination
func (r *AdditionalAnalysisRepository) GetAll(filter models.AdditionalAnalysisFilter) ([]models.AdditionalAnalysisResponse, int, error) {
	whereConditions := []string{}
	args := []interface{}{}

	// Build WHERE conditions for MySQL (use ? placeholders)
	if filter.AccountCode != "" {
		whereConditions = append(whereConditions, "aa.account_code = ?")
		args = append(args, filter.AccountCode)
	}

	if filter.AnalysisType != "" {
		whereConditions = append(whereConditions, "aa.analysis_type LIKE ?")
		args = append(args, "%"+filter.AnalysisType+"%")
	}

	
	if filter.Status != "" {
		whereConditions = append(whereConditions, "aa.status = ?")
		args = append(args, filter.Status)
	}

	if filter.Search != "" {
		whereConditions = append(whereConditions, "(aa.analysis_title LIKE ? OR a.account_name LIKE ?)")
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Build the full query
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Default sorting
	sortBy := "aa.created_at"
	sortOrder := "DESC"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder != "" {
		sortOrder = strings.ToUpper(filter.SortOrder)
	}

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM additional_analyses aa
		LEFT JOIN accounts a ON aa.account_code = a.account_code
		%s
	`, whereClause)

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count additional analyses: %w", err)
	}

	// Pagination
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Data query for MySQL
	query := fmt.Sprintf(`
		SELECT
			aa.id, aa.account_code, a.account_name,
			aa.analysis_type, aa.analysis_title, aa.status, aa.notes, aa.created_by,
			aa.created_at, aa.updated_at
		FROM additional_analyses aa
		LEFT JOIN accounts a ON aa.account_code = a.account_code
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortBy, sortOrder)

	args = append(args, limit, offset)

	var analyses []models.AdditionalAnalysisResponse
	err = r.db.Select(&analyses, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get additional analyses: %w", err)
	}

	return analyses, total, nil
}

// Update updates an existing additional analysis
func (r *AdditionalAnalysisRepository) Update(id int, analysis *models.AdditionalAnalysis) error {
	query := `
		UPDATE additional_analyses
		SET account_code = ?, analysis_type = ?, analysis_title = ?, status = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	updatedAt := time.Now()
	analysis.UpdatedAt = updatedAt

	result, err := r.db.Exec(query,
		analysis.AccountCode, analysis.AnalysisType, analysis.AnalysisTitle, analysis.Status, analysis.Notes,
		updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update additional analysis: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("additional analysis not found")
	}

	return nil
}

// Delete soft deletes an additional analysis by ID
func (r *AdditionalAnalysisRepository) Delete(id int) error {
	query := "UPDATE additional_analyses SET status = 'inactive', updated_at = ? WHERE id = ?"

	updatedAt := time.Now()
	result, err := r.db.Exec(query, updatedAt, id)
	if err != nil {
		return fmt.Errorf("failed to delete additional analysis: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("additional analysis not found")
	}

	return nil
}

// GetByAccountCode retrieves all additional analyses for a specific account
func (r *AdditionalAnalysisRepository) GetByAccountCode(accountCode string) ([]models.AdditionalAnalysis, error) {
	query := `
		SELECT id, account_code, analysis_type, analysis_title, status, notes, created_by, created_at, updated_at
		FROM additional_analyses
		WHERE account_code = ? AND status = 'active'
		ORDER BY created_at DESC
	`

	var analyses []models.AdditionalAnalysis
	err := r.db.Select(&analyses, query, accountCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get additional analyses by account code: %w", err)
	}

	return analyses, nil
}

// HardDelete permanently deletes an additional analysis by ID
func (r *AdditionalAnalysisRepository) HardDelete(id int) error {
	query := "DELETE FROM additional_analyses WHERE id = ?"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete additional analysis: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("additional analysis not found")
	}

	return nil
}

// GetAnalysisTypes retrieves all distinct analysis types
func (r *AdditionalAnalysisRepository) GetAnalysisTypes() ([]string, error) {
	query := "SELECT DISTINCT analysis_type FROM additional_analyses WHERE status = 'active' ORDER BY analysis_type"

	var types []string
	err := r.db.Select(&types, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis types: %w", err)
	}

	return types, nil
}