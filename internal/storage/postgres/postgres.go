package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"test-task/internal/domain/filters"
	"test-task/internal/domain/models"
	"test-task/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PeopleTable = "people"

	IdColumn          = "id"
	NameColumn        = "name"
	SurnameColumn     = "surname"
	PatronymicColumn  = "patronymic"
	AgeColumn         = "age"
	GenderColumn      = "gender"
	NationalityColumn = "nationality"
	CreatedColumn     = "created_at"
	UpdatedColum      = "updated_at"
)

var (
	ErrConnectString = errors.New("can't connect to Postgres")
	ErrTxBegin       = errors.New("can't start transaction")
	ErrTxCommit      = errors.New("can't commit transaction")
	ErrQuery         = errors.New("can't do query")
)

type PostgreStorage struct {
	conn *pgxpool.Pool
	log  *slog.Logger
}

type StoragePerson struct {
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}

func New(ctx context.Context, log *slog.Logger, connString string) (*PostgreStorage, error) {
	log.Debug("Connecting to database", "Connect String", connString)

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Error(ErrConnectString.Error(), "err", err.Error())

		return nil, fmt.Errorf("%w:%w", ErrConnectString, err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Error(ErrConnectString.Error(), "err", err.Error())

		return nil, fmt.Errorf("%w:%w", ErrConnectString, err)
	}

	log.Debug("Database is connected")

	return &PostgreStorage{
		conn: conn,
		log:  log,
	}, nil
}

func (s *PostgreStorage) Close() {
	s.conn.Close()
}

func (s *PostgreStorage) Ping(ctx context.Context) error {
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	err := s.conn.Ping(ctxTimeOut)
	if err != nil {
		return fmt.Errorf("no database connection:%w", err)
	}

	return nil
}

func (s *PostgreStorage) Save(ctx context.Context, entity *models.Person) (int64, error) {

	var id int64

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.log.Error(ErrTxBegin.Error(), "err", err.Error())

		return 0, fmt.Errorf("%w:%w", ErrTxBegin, err)
	}

	tx.Begin(ctx)
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`
	INSERT INTO %s
	(%s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING %s
	`, PeopleTable,
		NameColumn, SurnameColumn, PatronymicColumn, AgeColumn, GenderColumn, NationalityColumn,
		IdColumn,
	)

	err = s.conn.QueryRow(ctx, query,
		entity.Name,
		entity.Surname,
		entity.Patronymic,
		entity.Age,
		entity.Gender,
		entity.Nationality).Scan(&id)

	if err != nil {
		s.log.Error(ErrQuery.Error(), "err", err.Error())
		s.log.Debug(ErrQuery.Error(), "err", err.Error(), "query", query)

		return 0, fmt.Errorf("%s:%w", ErrQuery, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error(ErrTxCommit.Error(), "err", err.Error())
		s.log.Debug(ErrTxCommit.Error(), "err", err.Error(), "query", query)

		return 0, fmt.Errorf("%s:%w", ErrTxCommit, err)
	}

	return id, nil
}

func (s *PostgreStorage) Delete(ctx context.Context, id int64) error {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.log.Error(ErrTxBegin.Error(), "err", err.Error())

		return fmt.Errorf("%w:%w", ErrTxBegin, err)
	}

	tx.Begin(ctx)
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`
	DELETE FROM %s
	WHERE %s = ($1)
	`, PeopleTable,
		IdColumn,
	)

	commandTag, err := s.conn.Exec(ctx, query, id)
	if err != nil {
		s.log.Error(ErrQuery.Error(), "err", err.Error())
		s.log.Debug(ErrQuery.Error(), "err", err.Error(), "query", query)

		return fmt.Errorf("%s:%w", ErrQuery, err)
	}

	if commandTag.RowsAffected() == 0 {
		return storage.ErrIDNotFound
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error(ErrTxCommit.Error(), "err", err.Error())
		s.log.Debug(ErrTxCommit.Error(), "err", err.Error(), "query", query)

		return fmt.Errorf("%s:%w", ErrTxCommit, err)
	}

	return nil
}

func (s *PostgreStorage) FindByID(ctx context.Context, id int64) (*models.Person, error) {

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.log.Error(ErrTxBegin.Error(), "err", err.Error())

		return nil, fmt.Errorf("%w:%w", ErrTxBegin, err)
	}

	tx.Begin(ctx)
	defer tx.Rollback(ctx)

	result := StoragePerson{}

	query := fmt.Sprintf(`
	SELECT %s, %s, %s, %s, %s, %s FROM %s
	WHERE %s = ($1)
	`, NameColumn,
		SurnameColumn,
		PatronymicColumn,
		AgeColumn,
		GenderColumn,
		NationalityColumn,
		PeopleTable,
		IdColumn,
	)

	err = s.conn.QueryRow(ctx, query, id).Scan(
		&result.Name,
		&result.Surname,
		&result.Patronymic,
		&result.Age,
		&result.Gender,
		&result.Nationality,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.log.Debug("ID was not found")
			return nil, storage.ErrIDNotFound
		}
		s.log.Error(ErrQuery.Error(), "err", err.Error())
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error(ErrTxCommit.Error(), "err", err.Error())
		s.log.Debug(ErrTxCommit.Error(), "err", err.Error(), "query", query)

		return nil, fmt.Errorf("%s:%w", ErrTxCommit, err)
	}

	personModel := models.Person{
		Name:        result.Name,
		Surname:     result.Surname,
		Patronymic:  result.Patronymic,
		Age:         result.Age,
		Gender:      result.Gender,
		Nationality: result.Nationality,
	}

	return &personModel, nil
}

func (s *PostgreStorage) Update(ctx context.Context, entity *models.Person, id int64) error {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		s.log.Error(ErrTxBegin.Error(), "err", err.Error())

		return fmt.Errorf("%w:%w", ErrTxBegin, err)
	}

	tx.Begin(ctx)
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`
	   UPDATE %s
        SET
            %s = ($1),
			%s = ($2),
			%s = ($3),
			%s = ($4),
			%s = ($5),
			%s = ($6)
        WHERE %s = ($7)
		RETURNING %s;
		`,
		PeopleTable,
		NameColumn,
		SurnameColumn,
		PatronymicColumn,
		AgeColumn,
		GenderColumn,
		NationalityColumn,
		IdColumn,
		IdColumn,
	)

	_, err = s.conn.Exec(ctx, query,
		entity.Name,
		entity.Surname,
		entity.Patronymic,
		entity.Age,
		entity.Gender,
		entity.Nationality,
		id)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.log.Debug("ID was not found")
			return storage.ErrIDNotFound
		}
		s.log.Error(storage.ErrIDNotFound.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error(ErrTxCommit.Error(), "err", err.Error())
		s.log.Debug(ErrTxCommit.Error(), "err", err.Error(), "query", query)

		return fmt.Errorf("%s:%w", ErrTxCommit, err)
	}
	return nil
}

func (s *PostgreStorage) FilteredPages(ctx context.Context, offset int, limit int, options *filters.Options) ([]*models.Person, int, error) {

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.log.Error(ErrTxBegin.Error(), "err", err.Error())

		return nil, 0, fmt.Errorf("%w:%w", ErrTxBegin, err)
	}

	tx.Begin(ctx)
	defer tx.Rollback(ctx)

	list := []*models.Person{}

	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s 
		FROM %s `,
		NameColumn, SurnameColumn, PatronymicColumn, AgeColumn, GenderColumn, NationalityColumn,
		PeopleTable,
	)

	query, args := filter(query, options)

	count, err := s.countPeople(ctx, options, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Error("no people found")

		}
		s.log.Error(ErrQuery.Error(), "err", err.Error())

		return nil, 0, fmt.Errorf("%w:%w", ErrQuery, err)
	}

	query += fmt.Sprintf(" OFFSET ($%d) LIMIT ($%d)", len(args)+1, len(args)+2)

	args = append(args, offset, limit)

	s.log.Debug("Query statment enriched by filters:", "query", query)

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Error("no people found")

		}
		s.log.Error(ErrQuery.Error(), "err", err.Error())

		return nil, 0, fmt.Errorf("%w:%w", ErrQuery, err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Person

		err := rows.Scan(
			&p.Name,
			&p.Surname,
			&p.Patronymic,
			&p.Age,
			&p.Gender,
			&p.Nationality,
		)

		if err != nil {
			s.log.Error("can't scan row", "err", err.Error())
			return nil, 0, fmt.Errorf("can't scan row: %w", err)
		}
		list = append(list, &p)
	}

	if err := rows.Err(); err != nil {
		s.log.Error("rows error", "err", err.Error())
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return list, count, nil

}

func (s *PostgreStorage) countPeople(ctx context.Context, options *filters.Options, args []interface{}) (int, error) {

	var count int

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", PeopleTable)

	query, args = filter(query, options)

	err := s.conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Error("no people found")

		}
		s.log.Error(ErrQuery.Error(), "err", err.Error())

		return 0, fmt.Errorf("%w:%w", ErrQuery, err)
	}

	return count, nil

}

func filter(query string, options *filters.Options) (string, []interface{}) {
	var whereClauses []string
	var args []interface{}
	argNum := 1

	if options.Name != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", NameColumn, argNum))
		args = append(args, *options.Name)
		argNum++
	}
	if options.Surname != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", SurnameColumn, argNum))
		args = append(args, *options.Surname)
		argNum++
	}
	if options.Patronymic != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", PatronymicColumn, argNum))
		args = append(args, *options.Patronymic)
		argNum++
	}
	if options.Gender != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", GenderColumn, argNum))
		args = append(args, *options.Gender)
		argNum++
	}
	if options.Nationality != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", NationalityColumn, argNum))
		args = append(args, *options.Nationality)
		argNum++
	}
	if options.Age != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", AgeColumn, argNum))
		args = append(args, *options.Age)
		argNum++
	}
	if options.Age == nil && options.MinAge != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s => $%d", AgeColumn, argNum))
		args = append(args, *options.MinAge)
		argNum++
	}
	if options.Age == nil && options.MaxAge != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s <= $%d", AgeColumn, argNum))
		args = append(args, *options.MaxAge)
		argNum++
	}

	// Добавляем WHERE если есть условия
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return query, args
}
