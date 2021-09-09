package sqlitedb

import (
	"log"

	"github.com/rdmyldz/clone-hn/models"
)

func (s *SqliteHN) CreatePost(p *models.Post) (int, error) {
	stmt, err := s.db.Prepare(`INSERT INTO posts (link, title, domain, owner, points, parent_id,
		 main_post_id, comment_num, title_summary, created_at) 
		 values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	if p.MainPostID != 0 {
		ts, err := s.GetTitleSum(p.MainPostID)
		if err != nil {
			log.Printf("error while getting title_sum: %v\n", err)
			return -1, err
		}
		p.TitleSummary = ts
	}

	res, err := stmt.Exec(p.Link, p.Title, p.Domain, p.Owner, p.Points, p.ParentID, p.MainPostID, p.CommentNum, p.TitleSummary, p.CreatedAt)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (s *SqliteHN) GetTitleSum(id int) (string, error) {
	row := s.db.QueryRow("SELECT substr(title,1,5) FROM posts WHERE post_id = ?", id)
	var ts string
	err := row.Scan(&ts)
	if err != nil {
		return "", err
	}

	return ts, nil
}

func (s *SqliteHN) UpdateCommentNum(id int) error {
	row := s.db.QueryRow("SELECT comment_num FROM posts WHERE post_id=?", id)
	var cn int
	err := row.Scan(&cn)
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare("UPDATE posts SET comment_num=? WHERE post_id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(cn+1, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteHN) UpdatePoints(id int) error {
	row := s.db.QueryRow("SELECT points FROM posts WHERE post_id=?", id)
	var p int
	err := row.Scan(&p)
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare("UPDATE posts SET points=? WHERE post_id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(p+1, id)
	if err != nil {
		return err
	}
	return nil
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
		err := rows.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.MainPostID, &p.CommentNum, &p.TitleSummary, &p.CreatedAt)
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
	err := row.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.MainPostID, &p.CommentNum, &p.TitleSummary, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
