package service

import (
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/utils"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type AdditionalAnalysisService struct {
	analysisRepo *repository.AdditionalAnalysisRepository
	accountRepo  *repository.AccountRepository
	logger       *logrus.Logger
}

func NewAdditionalAnalysisService(analysisRepo *repository.AdditionalAnalysisRepository, accountRepo *repository.AccountRepository, logger *logrus.Logger) *AdditionalAnalysisService {
	return &AdditionalAnalysisService{
		analysisRepo: analysisRepo,
		accountRepo:  accountRepo,
		logger:       logger,
	}
}

// Create creates a new additional analysis
func (s *AdditionalAnalysisService) Create(req models.AdditionalAnalysisRequest, createdBy int) (*models.AdditionalAnalysisResponse, error) {
	// Validate if account exists
	_, err := s.accountRepo.FindByCode(req.AccountCode)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Set default values
	if req.Status == "" {
		req.Status = "active"
	}

	// Create analysis model
	analysis := &models.AdditionalAnalysis{
		AccountCode:   req.AccountCode,
		AnalysisType:  req.AnalysisType,
		AnalysisTitle: req.AnalysisTitle,
		Status:        req.Status,
		Notes:         &req.Notes,
		CreatedBy:     &createdBy,
	}

	err = s.analysisRepo.Create(analysis)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create additional analysis")
		return nil, fmt.Errorf("failed to create additional analysis: %w", err)
	}

	// Get the created analysis with account details
	result, err := s.analysisRepo.GetByID(analysis.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created additional analysis: %w", err)
	}

	s.logger.WithField("id", analysis.ID).Info("Additional analysis created successfully")
	return result, nil
}

// GetByID retrieves an additional analysis by ID
func (s *AdditionalAnalysisService) GetByID(id int) (*models.AdditionalAnalysisResponse, error) {
	analysis, err := s.analysisRepo.GetByID(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get additional analysis")
		return nil, fmt.Errorf("failed to get additional analysis: %w", err)
	}

	return analysis, nil
}

// GetAll retrieves all additional analyses with filtering and pagination
func (s *AdditionalAnalysisService) GetAll(filter models.AdditionalAnalysisFilter) ([]models.AdditionalAnalysisResponse, *utils.PaginationMeta, error) {
	// Set default pagination values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	analyses, total, err := s.analysisRepo.GetAll(filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get additional analyses")
		return nil, nil, fmt.Errorf("failed to get additional analyses: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (total + filter.Limit - 1) / filter.Limit
	meta := &utils.PaginationMeta{
		CurrentPage: filter.Page,
		PerPage:     filter.Limit,
		Total:       int64(total),
		LastPage:    totalPages,
		From:        (filter.Page-1)*filter.Limit + 1,
		To:          filter.Page * filter.Limit,
		HasMore:     filter.Page < totalPages,
	}

	return analyses, meta, nil
}

// Update updates an existing additional analysis
func (s *AdditionalAnalysisService) Update(id int, req models.AdditionalAnalysisRequest) (*models.AdditionalAnalysisResponse, error) {
	// Check if analysis exists
	existing, err := s.analysisRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("additional analysis not found: %w", err)
	}

	// Validate if account exists (if changed)
	if req.AccountCode != existing.AccountCode {
		_, err := s.accountRepo.FindByCode(req.AccountCode)
		if err != nil {
			return nil, fmt.Errorf("account not found: %w", err)
		}
	}

	// Create analysis model
	analysis := &models.AdditionalAnalysis{
		AccountCode:   req.AccountCode,
		AnalysisType:  req.AnalysisType,
		AnalysisTitle: req.AnalysisTitle,
		Status:        req.Status,
		Notes:         &req.Notes,
	}

	err = s.analysisRepo.Update(id, analysis)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update additional analysis")
		return nil, fmt.Errorf("failed to update additional analysis: %w", err)
	}

	// Get updated analysis
	result, err := s.analysisRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated additional analysis: %w", err)
	}

	s.logger.WithField("id", id).Info("Additional analysis updated successfully")
	return result, nil
}

// Delete soft deletes an additional analysis by ID
func (s *AdditionalAnalysisService) Delete(id int) error {
	// Check if analysis exists
	_, err := s.analysisRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("additional analysis not found: %w", err)
	}

	err = s.analysisRepo.Delete(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete additional analysis")
		return fmt.Errorf("failed to delete additional analysis: %w", err)
	}

	s.logger.WithField("id", id).Info("Additional analysis deleted successfully")
	return nil
}

// HardDelete permanently deletes an additional analysis by ID
func (s *AdditionalAnalysisService) HardDelete(id int) error {
	// Check if analysis exists
	_, err := s.analysisRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("additional analysis not found: %w", err)
	}

	err = s.analysisRepo.HardDelete(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to hard delete additional analysis")
		return fmt.Errorf("failed to hard delete additional analysis: %w", err)
	}

	s.logger.WithField("id", id).Info("Additional analysis hard deleted successfully")
	return nil
}

// GetByAccountCode retrieves all additional analyses for a specific account
func (s *AdditionalAnalysisService) GetByAccountCode(accountCode string) ([]models.AdditionalAnalysis, error) {
	// Check if account exists
	_, err := s.accountRepo.FindByCode(accountCode)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	analyses, err := s.analysisRepo.GetByAccountCode(accountCode)
	if err != nil {
		s.logger.WithError(err).WithField("account_code", accountCode).Error("Failed to get additional analyses by account code")
		return nil, fmt.Errorf("failed to get additional analyses: %w", err)
	}

	return analyses, nil
}

// GetAnalysisTypes retrieves all distinct analysis types
func (s *AdditionalAnalysisService) GetAnalysisTypes() ([]string, error) {
	types, err := s.analysisRepo.GetAnalysisTypes()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get analysis types")
		return nil, fmt.Errorf("failed to get analysis types: %w", err)
	}

	return types, nil
}

// ImportFromExcel imports additional analyses from Excel file
func (s *AdditionalAnalysisService) ImportFromExcel(analyses []models.AdditionalAnalysisRequest) (*models.AdditionalAnalysisImportResult, error) {
	result := &models.AdditionalAnalysisImportResult{
		ProcessedAt: time.Now(),
		Errors:      []models.AdditionalAnalysisValidationError{},
	}

	for i, analysis := range analyses {
		row := i + 2 // Excel row numbers start from 2 (assuming header is row 1)

		// Validate required fields
		if analysis.AccountCode == "" {
			result.Failed++
			result.Errors = append(result.Errors, models.AdditionalAnalysisValidationError{
				Row:     row,
				Field:   "account_code",
				Value:   analysis.AccountCode,
				Message: "Account code is required",
				Data:    analysis,
			})
			continue
		}

		if analysis.AnalysisType == "" {
			result.Failed++
			result.Errors = append(result.Errors, models.AdditionalAnalysisValidationError{
				Row:     row,
				Field:   "analysis_type",
				Value:   analysis.AnalysisType,
				Message: "Analysis type is required",
				Data:    analysis,
			})
			continue
		}

		if analysis.AnalysisTitle == "" {
			result.Failed++
			result.Errors = append(result.Errors, models.AdditionalAnalysisValidationError{
				Row:     row,
				Field:   "analysis_title",
				Value:   analysis.AnalysisTitle,
				Message: "Analysis title is required",
				Data:    analysis,
			})
			continue
		}

		// Set default values
		if analysis.Status == "" {
			analysis.Status = "active"
		}

		// Create analysis
		_, err := s.Create(analysis, 0) // 0 for system import
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, models.AdditionalAnalysisValidationError{
				Row:     row,
				Field:   "general",
				Message: err.Error(),
				Data:    analysis,
			})
			continue
		}

		result.Success++
	}

	result.Total = result.Success + result.Failed

	s.logger.WithFields(logrus.Fields{
		"total":   result.Total,
		"success": result.Success,
		"failed":  result.Failed,
	}).Info("Additional analysis import completed")

	return result, nil
}

// ExportToExcel exports additional analyses to Excel format
func (s *AdditionalAnalysisService) ExportToExcel(filter models.AdditionalAnalysisExportRequest) ([]byte, error) {
	// Convert export filter to regular filter
	analysisFilter := models.AdditionalAnalysisFilter{
		AccountCode:  filter.AccountCode,
		AnalysisType: filter.AnalysisType,
		Status:       filter.Status,
		Search:       filter.Search,
		Limit:        10000, // Large limit for export
		Page:         1,
	}

	// Get analyses
	analyses, _, err := s.analysisRepo.GetAll(analysisFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses for export: %w", err)
	}

	// Convert to Excel (this would typically use an Excel library like excelize)
	// For now, return a placeholder
	// TODO: Implement actual Excel export functionality
	return []byte(fmt.Sprintf("Excel export functionality to be implemented. Found %d analyses.", len(analyses))), nil
}