package sqlitedb

import (
	"context"
	"fmt"

	"github.com/rdmyldz/clone-hn/models"
)

func (s *SqliteHN) CreatePost(ctx context.Context, p *models.Post) (int, error) {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO posts (link, title, domain, owner, points, parent_id,
		 main_post_id, comment_num, title_summary, created_at) 
		 values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return -1, fmt.Errorf("in CreatePost, error while preparing statement: %w", err)
	}
	defer stmt.Close()

	if p.MainPostID != 0 {
		ts, err := s.GetTitleSum(ctx, p.MainPostID)
		if err != nil {
			return -1, fmt.Errorf("in CreatePost, error while getting title_sum: %w", err)
		}
		p.TitleSummary = ts
	}

	res, err := stmt.ExecContext(ctx, p.Link, p.Title, p.Domain, p.Owner, p.Points, p.ParentID, p.MainPostID, p.CommentNum, p.TitleSummary, p.CreatedAt)
	if err != nil {
		return -1, fmt.Errorf("in CreatePost, error while executing statement: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("in CreatePost, error while getting last inserted id: %w", err)
	}
	return int(id), nil
}

func (s *SqliteHN) GetTitleSum(ctx context.Context, id int) (string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT substr(title,1,25) FROM posts WHERE post_id = ?", id)
	var ts string
	err := row.Scan(&ts)
	if err != nil {
		return "", fmt.Errorf("in GetTitleSum, error while scanning row: %w", err)
	}

	return ts, nil
}

func (s *SqliteHN) UpdateCommentNum(ctx context.Context, id int) error {
	row := s.db.QueryRowContext(ctx, "SELECT comment_num FROM posts WHERE post_id=?", id)
	var cn int
	err := row.Scan(&cn)
	if err != nil {
		return fmt.Errorf("in UpdateCommentNum, error while querying row: %w", err)
	}

	stmt, err := s.db.PrepareContext(ctx, "UPDATE posts SET comment_num=? WHERE post_id=?")
	if err != nil {
		return fmt.Errorf("in UpdateCommentNum, error while preparing statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx, cn+1, id)
	if err != nil {
		return fmt.Errorf("in UpdateCommentNum, error while executing statement: %w", err)
	}
	return nil
}

func (s *SqliteHN) UpdatePoints(ctx context.Context, id int) error {
	row := s.db.QueryRowContext(ctx, "SELECT points FROM posts WHERE post_id=?", id)
	var p int
	err := row.Scan(&p)
	if err != nil {
		return fmt.Errorf("in UpdatePoints, error while querying row: %w", err)
	}

	stmt, err := s.db.PrepareContext(ctx, "UPDATE posts SET points=? WHERE post_id=?")
	if err != nil {
		return fmt.Errorf("in UpdatePoints, error while preparing statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx, p+1, id)
	if err != nil {
		return fmt.Errorf("in UpdatePoints, error while executing statement: %w", err)
	}
	return nil
}

func (s *SqliteHN) GetPosts(ctx context.Context, query string) ([]models.Post, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("in GetPosts, error while querying rows: %w", err)
	}
	defer rows.Close()
	var ret []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.MainPostID, &p.CommentNum, &p.TitleSummary, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("in GetPosts, error while scaning row: %w", err)
		}
		ret = append(ret, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *SqliteHN) GetPost(ctx context.Context, query, id string) (*models.Post, error) {
	row := s.db.QueryRowContext(ctx, query, id)
	var p models.Post
	err := row.Scan(&p.ID, &p.Link, &p.Title, &p.Domain, &p.Owner, &p.Points, &p.ParentID, &p.MainPostID, &p.CommentNum, &p.TitleSummary, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("in GetPost, error while querying row: %w", err)
	}

	return &p, nil
}
