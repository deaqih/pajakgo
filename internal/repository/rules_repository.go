package repository

import (
	"accounting-web/internal/models"

	"github.com/jmoiron/sqlx"
)

type RulesRepository struct {
	db *sqlx.DB
}

func NewRulesRepository(db *sqlx.DB) *RulesRepository {
	return &RulesRepository{db: db}
}

// Koreksi Rules
func (r *RulesRepository) GetKoreksiRules(limit, offset int, search string) ([]models.KoreksiRule, int, error) {
	var rules []models.KoreksiRule
	var total int

	// Build base queries
	countQuery := "SELECT COUNT(*) FROM koreksi_rules"
	selectQuery := "SELECT * FROM koreksi_rules"

	// Add search condition if provided
	whereClause := ""
	args := []interface{}{}
	if search != "" {
		whereClause = " WHERE keyword LIKE ? OR value LIKE ?"
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam)
		countQuery += whereClause
		selectQuery += whereClause
	}

	// Get total count
	if len(args) > 0 {
		err := r.db.Get(&total, countQuery, args...)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.Get(&total, countQuery)
		if err != nil {
			return nil, 0, err
		}
	}

	// Get paginated results
	selectQuery += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.db.Select(&rules, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *RulesRepository) GetActiveKoreksiRules() ([]models.KoreksiRule, error) {
	var rules []models.KoreksiRule
	query := "SELECT * FROM koreksi_rules WHERE is_active = TRUE ORDER BY id DESC"
	err := r.db.Select(&rules, query)
	return rules, err
}

func (r *RulesRepository) CreateKoreksiRule(rule *models.KoreksiRule) error {
	query := `INSERT INTO koreksi_rules (keyword, value, is_active)
	          VALUES (:keyword, :value, :is_active)`
	result, err := r.db.NamedExec(query, rule)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	rule.ID = int(id)
	return nil
}

func (r *RulesRepository) UpdateKoreksiRule(rule *models.KoreksiRule) error {
	query := `UPDATE koreksi_rules SET keyword = :keyword, value = :value,
	          is_active = :is_active WHERE id = :id`
	_, err := r.db.NamedExec(query, rule)
	return err
}

func (r *RulesRepository) GetKoreksiRuleByID(id int) (*models.KoreksiRule, error) {
	var rule models.KoreksiRule
	query := "SELECT * FROM koreksi_rules WHERE id = ?"
	err := r.db.Get(&rule, query, id)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *RulesRepository) DeleteKoreksiRule(id int) error {
	query := "DELETE FROM koreksi_rules WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *RulesRepository) BulkInsertKoreksiRules(rules []models.KoreksiRule) error {
	if len(rules) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO koreksi_rules (keyword, value, is_active) VALUES (:keyword, :value, :is_active)`

	for _, rule := range rules {
		_, err := tx.NamedExec(query, rule)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RulesRepository) GetAllActiveKoreksiRules() ([]models.KoreksiRule, error) {
	var rules []models.KoreksiRule
	query := "SELECT * FROM koreksi_rules WHERE is_active = TRUE ORDER BY id DESC"
	err := r.db.Select(&rules, query)
	return rules, err
}

// Obyek Rules
func (r *RulesRepository) GetObyekRules(limit, offset int, search string) ([]models.ObyekRule, int, error) {
	var rules []models.ObyekRule
	var total int

	// Build base queries
	countQuery := "SELECT COUNT(*) FROM obyek_rules"
	selectQuery := "SELECT * FROM obyek_rules"

	// Add search condition if provided
	whereClause := ""
	args := []interface{}{}
	if search != "" {
		whereClause = " WHERE keyword LIKE ? OR value LIKE ?"
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam)
		countQuery += whereClause
		selectQuery += whereClause
	}

	// Get total count
	if len(args) > 0 {
		err := r.db.Get(&total, countQuery, args...)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.Get(&total, countQuery)
		if err != nil {
			return nil, 0, err
		}
	}

	// Get paginated results
	selectQuery += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.db.Select(&rules, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *RulesRepository) GetActiveObyekRules() ([]models.ObyekRule, error) {
	var rules []models.ObyekRule
	query := "SELECT * FROM obyek_rules WHERE is_active = TRUE ORDER BY id DESC"
	err := r.db.Select(&rules, query)
	return rules, err
}

func (r *RulesRepository) CreateObyekRule(rule *models.ObyekRule) error {
	query := `INSERT INTO obyek_rules (keyword, value, is_active)
	          VALUES (:keyword, :value, :is_active)`
	result, err := r.db.NamedExec(query, rule)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	rule.ID = int(id)
	return nil
}

func (r *RulesRepository) UpdateObyekRule(rule *models.ObyekRule) error {
	query := `UPDATE obyek_rules SET keyword = :keyword, value = :value,
	          is_active = :is_active WHERE id = :id`
	_, err := r.db.NamedExec(query, rule)
	return err
}

func (r *RulesRepository) GetObyekRuleByID(id int) (*models.ObyekRule, error) {
	var rule models.ObyekRule
	query := "SELECT * FROM obyek_rules WHERE id = ?"
	err := r.db.Get(&rule, query, id)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *RulesRepository) DeleteObyekRule(id int) error {
	query := "DELETE FROM obyek_rules WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *RulesRepository) BulkInsertObyekRules(rules []models.ObyekRule) error {
	if len(rules) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO obyek_rules (keyword, value, is_active) VALUES (:keyword, :value, :is_active)`

	for _, rule := range rules {
		_, err := tx.NamedExec(query, rule)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RulesRepository) GetAllActiveObyekRules() ([]models.ObyekRule, error) {
	var rules []models.ObyekRule
	query := "SELECT * FROM obyek_rules WHERE is_active = TRUE ORDER BY id DESC"
	err := r.db.Select(&rules, query)
	return rules, err
}

// Withholding Tax Rules
func (r *RulesRepository) GetWithholdingTaxRules(limit, offset int) ([]models.WithholdingTaxRule, int, error) {
	var rules []models.WithholdingTaxRule
	var total int

	countQuery := "SELECT COUNT(*) FROM withholding_tax_rules"
	err := r.db.Get(&total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT * FROM withholding_tax_rules ORDER BY priority DESC, id LIMIT ? OFFSET ?"
	err = r.db.Select(&rules, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *RulesRepository) GetActiveWithholdingTaxRules() ([]models.WithholdingTaxRule, error) {
	var rules []models.WithholdingTaxRule
	query := "SELECT * FROM withholding_tax_rules WHERE is_active = TRUE ORDER BY priority DESC"
	err := r.db.Select(&rules, query)
	return rules, err
}

func (r *RulesRepository) CreateWithholdingTaxRule(rule *models.WithholdingTaxRule) error {
	query := `INSERT INTO withholding_tax_rules (keyword, tax_type, tax_rate, priority, is_active)
	          VALUES (:keyword, :tax_type, :tax_rate, :priority, :is_active)`
	result, err := r.db.NamedExec(query, rule)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	rule.ID = int(id)
	return nil
}

func (r *RulesRepository) UpdateWithholdingTaxRule(rule *models.WithholdingTaxRule) error {
	query := `UPDATE withholding_tax_rules SET keyword = :keyword, tax_type = :tax_type,
	          tax_rate = :tax_rate, priority = :priority, is_active = :is_active WHERE id = :id`
	_, err := r.db.NamedExec(query, rule)
	return err
}

func (r *RulesRepository) DeleteWithholdingTaxRule(id int) error {
	query := "DELETE FROM withholding_tax_rules WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// Tax Keywords
func (r *RulesRepository) GetTaxKeywords(limit, offset int) ([]models.TaxKeyword, int, error) {
	var keywords []models.TaxKeyword
	var total int

	countQuery := "SELECT COUNT(*) FROM tax_keywords"
	err := r.db.Get(&total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT * FROM tax_keywords ORDER BY priority DESC, id LIMIT ? OFFSET ?"
	err = r.db.Select(&keywords, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return keywords, total, nil
}

func (r *RulesRepository) GetActiveTaxKeywords() ([]models.TaxKeyword, error) {
	var keywords []models.TaxKeyword
	query := "SELECT * FROM tax_keywords WHERE is_active = TRUE ORDER BY priority DESC"
	err := r.db.Select(&keywords, query)
	return keywords, err
}

func (r *RulesRepository) CreateTaxKeyword(keyword *models.TaxKeyword) error {
	query := `INSERT INTO tax_keywords (keyword, tax_category, priority, is_active)
	          VALUES (:keyword, :tax_category, :priority, :is_active)`
	result, err := r.db.NamedExec(query, keyword)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	keyword.ID = int(id)
	return nil
}

func (r *RulesRepository) UpdateTaxKeyword(keyword *models.TaxKeyword) error {
	query := `UPDATE tax_keywords SET keyword = :keyword, tax_category = :tax_category,
	          priority = :priority, is_active = :is_active WHERE id = :id`
	_, err := r.db.NamedExec(query, keyword)
	return err
}

func (r *RulesRepository) DeleteTaxKeyword(id int) error {
	query := "DELETE FROM tax_keywords WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}
