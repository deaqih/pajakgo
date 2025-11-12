package repository

import (
	"accounting-web/internal/models"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindAll(limit, offset int, search string) ([]models.Account, int, error) {
	var accounts []models.Account
	var total int

	// Build query with search
	whereClause := ""
	args := []interface{}{}

	if search != "" {
		whereClause = "WHERE account_code LIKE ? OR account_name LIKE ?"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM accounts %s", whereClause)
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data with COALESCE to handle NULL values
	query := fmt.Sprintf(`
		SELECT id,
		       account_code,
		       account_name,
		       COALESCE(account_type, '') as account_type,
		       COALESCE(nature, '') as nature,
		       COALESCE(koreksi_obyek, '') as koreksi_obyek,
		       COALESCE(analisa_tambahan, '') as analisa_tambahan,
		       is_active,
		       created_at,
		       updated_at
		FROM accounts %s
		ORDER BY account_code
		LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)
	err = r.db.Select(&accounts, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *AccountRepository) FindByID(id int) (*models.Account, error) {
	var account models.Account
	query := `
		SELECT id,
		       account_code,
		       account_name,
		       COALESCE(account_type, '') as account_type,
		       COALESCE(nature, '') as nature,
		       COALESCE(koreksi_obyek, '') as koreksi_obyek,
		       COALESCE(analisa_tambahan, '') as analisa_tambahan,
		       is_active,
		       created_at,
		       updated_at
		FROM accounts
		WHERE id = ?
		LIMIT 1`
	err := r.db.Get(&account, query, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) FindByCode(code string) (*models.Account, error) {
	var account models.Account
	query := "SELECT * FROM accounts WHERE account_code = ? LIMIT 1"
	err := r.db.Get(&account, query, code)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) Create(account *models.Account) error {
	query := `INSERT INTO accounts (account_code, account_name, account_type, nature, koreksi_obyek, analisa_tambahan, is_active)
	          VALUES (:account_code, :account_name, :account_type, :nature, :koreksi_obyek, :analisa_tambahan, :is_active)`
	result, err := r.db.NamedExec(query, account)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	account.ID = int(id)
	return nil
}

func (r *AccountRepository) Update(account *models.Account) error {
	query := `UPDATE accounts SET account_code = :account_code, account_name = :account_name,
	          account_type = :account_type, nature = :nature, koreksi_obyek = :koreksi_obyek,
	          analisa_tambahan = :analisa_tambahan, is_active = :is_active
	          WHERE id = :id`
	_, err := r.db.NamedExec(query, account)
	return err
}

func (r *AccountRepository) Delete(id int) error {
	query := "DELETE FROM accounts WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *AccountRepository) BulkInsert(accounts []models.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	query := `INSERT INTO accounts (account_code, account_name, account_type, nature, koreksi_obyek, analisa_tambahan, is_active)
	          VALUES (:account_code, :account_name, :account_type, :nature, :koreksi_obyek, :analisa_tambahan, :is_active)
	          ON DUPLICATE KEY UPDATE
	          account_name = VALUES(account_name),
	          account_type = VALUES(account_type),
	          nature = VALUES(nature),
	          koreksi_obyek = VALUES(koreksi_obyek),
	          analisa_tambahan = VALUES(analisa_tambahan),
	          is_active = VALUES(is_active)`
	_, err := r.db.NamedExec(query, accounts)
	return err
}

func (r *AccountRepository) GetAllActive() ([]models.Account, error) {
	var accounts []models.Account
	query := `
		SELECT id,
		       account_code,
		       account_name,
		       COALESCE(account_type, '') as account_type,
		       COALESCE(nature, '') as nature,
		       COALESCE(koreksi_obyek, '') as koreksi_obyek,
		       COALESCE(analisa_tambahan, '') as analisa_tambahan,
		       is_active,
		       created_at,
		       updated_at
		FROM accounts
		WHERE is_active = TRUE
		ORDER BY account_code`
	err := r.db.Select(&accounts, query)
	return accounts, err
}
