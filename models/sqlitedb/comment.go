package sqlitedb

import (
	"time"

	"github.com/rdmyldz/clone-hn/models"
)

func (s *SqliteHN) CreateComment(c *models.Comment) (int, error) {
	stmt, err := s.db.Prepare("INSERT INTO comments (text, created_at, owner, parent_id, post_id) values(?, ?, ?, ?, ?)")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(c.Text, time.Now(), c.Owner, c.ParentID, c.PostID)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (s *SqliteHN) GetComments(query string, pid int) ([]models.Comment, error) {
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.Text, &c.CreatedAt, &c.Owner, &c.ParentID, &c.PostID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *SqliteHN) GetComment(query string, id string) (*models.Comment, error) {
	row := s.db.QueryRow(query, id)
	var c models.Comment
	err := row.Scan(&c.ID, &c.Text, &c.CreatedAt, &c.Owner, &c.ParentID, &c.PostID)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
