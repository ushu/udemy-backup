package client

import "context"

var DefaultClient = New()

func Login(username, password string) (ID, accessToken string, err error) {
	return DefaultClient.Login(context.Background(), username, password)
}

func GetUser() (*User, error) {
	return DefaultClient.GetUser(context.Background())
}

func ListCourses(opt *PaginationOptions) (*Courses, error) {
	return DefaultClient.ListCourses(context.Background(), opt)
}

func GetCourse(ID int) (*Course, error) {
	return DefaultClient.GetCourse(context.Background(), ID)
}

func LoadCurriculum(courseID int, opt *PaginationOptions) (*Curriculum, error) {
	return DefaultClient.LoadCurriculum(context.Background(), courseID, opt)
}
