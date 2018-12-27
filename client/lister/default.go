package lister

import (
	"context"
	"github.com/ushu/udemy-backup/client"
)

func ListAllCourses() ([]*client.Course, error) {
	l := (*Lister)(client.DefaultClient)
	return l.ListAllCourses(context.Background())
}

func LoadFullCurriculum(courseID int) ([]*client.Lecture, error) {
	l := (*Lister)(client.DefaultClient)
	return l.LoadFullCurriculum(context.Background(), courseID)
}
