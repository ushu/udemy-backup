package lister

import (
	"context"
	"github.com/ushu/udemy-backup/client"
)

const DefaultPageSize = 1400

type Lister client.Client

func New(c *client.Client) *Lister {
	return (*Lister)(c)
}

func (l *Lister) LoadFullCurriculum(ctx context.Context, courseID int) ([]*client.Lecture, error) {
	var res []*client.Lecture
	opt := &client.PaginationOptions{
		Page:     1,
		PageSize: DefaultPageSize,
	}
	for {
		// load page info
		cur, err := (*client.Client)(l).LoadCurriculum(ctx, courseID, opt)
		if err != nil {
			return res, err
		}
		res = append(res, cur.Results...)

		// last page ?
		if cur.Next == "" {
			break
		}
		opt.Page++
	}

	return res, nil
}

func (l *Lister) ListAllCourses(ctx context.Context) ([]*Course, error) {
	var cc []*client.Course
	opt := &client.PaginationOptions{
		Page:     1,
		PageSize: DefaultPageSize,
	}
	for {
		// load page info
		res, err := (*client.Client)(l).ListCourses(ctx, opt)
		if err != nil {
			return cc, err
		}
		cc = append(cc, res.Results...)

		// last page ?
		if res.Next == "" {
			break
		}
		opt.Page++
	}
	return cc, nil
}
