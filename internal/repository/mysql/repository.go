package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/normie7/nore/internal/noiseremover"
)

// http://go-database-sql.org/accessing.html
// https://tutorialedge.net/golang/golang-mysql-tutorial/

type File struct {
	Id           string
	InternalName string
	UploadedName string
	CreatedAt    time.Time
	Progress     noiseremover.Progress
}

func (f *File) TableName() string {
	return "files"
}

// todo timeouts
type mysqlRepo struct {
	db *sql.DB
}

// use postgresql update... returning... instead
func (m *mysqlRepo) QueueFiles(counter int64) ([]noiseremover.File, error) {
	files := make([]noiseremover.File, 0)

	Tx, err := m.db.Begin()
	if err != nil {
		return files, err
	}

	// selecting with queued status, so we don't have to select again after update
	q := fmt.Sprintf("SELECT id, internal_name, uploaded_name,'"+string(noiseremover.ProgressQueued)+"' FROM %s "+
		"WHERE progress = ? ORDER BY created_at ASC LIMIT ? FOR UPDATE", (&File{}).TableName())
	results, err := Tx.Query(q, noiseremover.ProgressNew, counter)
	if err != nil {
		_ = Tx.Rollback()
		return files, err
	}
	for results.Next() {
		var file noiseremover.File
		err = results.Scan(&file.Id, &file.InternalName, &file.UploadedName, &file.Progress)
		if err != nil {
			return make([]noiseremover.File, 0), err
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		err = Tx.Commit()
		if err != nil {
			return make([]noiseremover.File, 0), err
		}
		return files, nil
	}

	args := make([]interface{}, len(files))
	for i, f := range files {
		args[i] = f.Id
	}
	q = fmt.Sprintf("update %s set progress = 'queued' where id in (?"+strings.Repeat(",?", len(files)-1)+")",
		(&File{}).TableName())
	_, err = Tx.Exec(q, args...)
	if err != nil {
		_ = Tx.Rollback()
		return make([]noiseremover.File, 0), err
	}

	err = Tx.Commit()
	if err != nil {
		return make([]noiseremover.File, 0), err
	}
	return files, nil
}

func (m *mysqlRepo) SetProgress(fileId string, progress noiseremover.Progress) error {
	_, err := m.db.Exec("UPDATE files SET progress = ? where id = ?",
		progress,
		fileId,
	)
	return err
}

func (m *mysqlRepo) GetFilesToProcess() ([]noiseremover.File, error) {

	q := fmt.Sprintf("SELECT id, internal_name, uploaded_name, progress FROM %s where progress = ?", (&File{}).TableName())

	results, err := m.db.Query(q, noiseremover.ProgressNew)
	if err != nil {
		return make([]noiseremover.File, 0), err
	}

	files := make([]noiseremover.File, 0)
	for results.Next() {
		var file noiseremover.File
		err = results.Scan(&file.Id, &file.InternalName, &file.UploadedName, &file.Progress)
		if err != nil {
			return make([]noiseremover.File, 0), err
		}
		files = append(files, file)
	}

	return files, nil
}

func (m *mysqlRepo) GetInfo(fileId string) (*noiseremover.File, error) {
	f := File{}
	err := m.db.QueryRow("SELECT id, internal_name, uploaded_name, created_at, progress FROM files where id = ?", fileId).
		Scan(&f.Id, &f.InternalName, &f.UploadedName, &f.CreatedAt, &f.Progress)
	if err != nil {
		return &noiseremover.File{}, err
	}

	return &noiseremover.File{
		Id:           f.Id,
		InternalName: f.InternalName,
		UploadedName: f.UploadedName,
		Progress:     f.Progress,
	}, nil
}

func (m *mysqlRepo) Add(file *noiseremover.File) error {
	_, err := m.db.Exec("INSERT INTO files (id, internal_name, uploaded_name, created_at, progress)VALUES ( ?, ?, ?, ?, ? )",
		file.Id,
		file.InternalName,
		file.UploadedName,
		time.Now().UTC(),
		file.Progress,
	)
	return err
}

func (m *mysqlRepo) Close() error {
	return m.db.Close()
}

func NewMysqlRepo(user, password, host, port, dbName string) *mysqlRepo {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName))
	if err != nil {
		// todo move to main?
		log.Fatal(err)
	}
	return &mysqlRepo{db: db}
}
