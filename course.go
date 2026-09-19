package main

import (
	"fmt"
	"os"
	"strconv"
)

type Course struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
}

type CoursesResponse []Course

type Compiler struct {
	ID int `json:"id"`
}

type CompilersResponse []Compiler

type Problem struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

type ProblemsResponse []Problem

func (c *Client) SetCourseId(courseId int) error {
	c.session.CourseID = &courseId
	return c.store.Save(&c.session)
}

func (c *Client) Courses() ([]Course, error) {
	courseReq, err := c.NewRequest[any]("GET", "/my/courses", nil)
	if err != nil {
		return nil, err
	}

	courseResp, err := c.Do[CoursesResponse](courseReq)
	if err != nil {
		return nil, err
	}

	return courseResp, nil
}

func (c *Client) Compilers(courseId int) ([]Compiler, error) {
	endpoint := fmt.Sprintf("/my/courses/%d/compilers", courseId)
	compilersReq, err := c.NewRequest[any]("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	compilersResp, err := c.Do[CompilersResponse](compilersReq)
	if err != nil {
		return nil, err
	}

	return compilersResp, nil
}

func (c *Client) Problems(courseId int) ([]Problem, error) {
	endpoint := fmt.Sprintf("/my/courses/%d/problems", courseId)
	problemsReq, err := c.NewRequest[any]("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	problemsResp, err := c.Do[ProblemsResponse](problemsReq)
	if err != nil {
		return nil, err
	}

	return problemsResp, nil
}

func CourseCommand(c *Client) {
	courses, err := c.Courses()
	if err != nil {
		fmt.Println("Failed to retrieve courses:", err)
		return
	}

	if len(os.Args) < 3 {
		for idx, course := range courses {
			fmt.Print(idx+1, ". ")
			fmt.Println(course.Name, "-", course.Subject)
		}
		return
	}

	selectedIdx, err := strconv.Atoi(os.Args[2])
	if err != nil || selectedIdx <= 0 || selectedIdx > len(courses) {
		fmt.Println("Invalid course index")
		return
	}

	course := courses[selectedIdx-1]
	if err := c.SetCourseId(course.ID); err != nil {
		fmt.Println("Failed to save course:", err)
		return
	}

	fmt.Println("Selected course:", course.Name, "-", course.Subject)
}
