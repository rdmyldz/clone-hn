package sqlitedb

import (
	"time"

	"github.com/rdmyldz/clone-hn/models"
)

func (s *SqliteHN) CreatePost(p *models.Post) (int, error) {
	stmt, err := s.db.Prepare("INSERT INTO posts (link, title, domain, owner, points, parent_id, created_at) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(p.Link, p.Title, p.Domain, p.Owner, p.Points, p.ParentID, time.Now())
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (s *SqliteHN) GetPosts(query string) ([]models.Post, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *SqliteHN) GetPost(query, id string) (*models.Post, error) {
	row := s.db.QueryRow(query, id)
	var p models.Post
	err := row.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
