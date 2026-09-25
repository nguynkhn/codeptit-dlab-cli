package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SubmitRequest struct {
	CourseID   int    `json:"course_id"`
	QuestionID int    `json:"question_id"`
	CompilerID int    `json:"compiler_id"`
	SourceCode string `json:"source_code"`
}

type SubmitResponse struct {
	ID int `json:"id"`
}

const (
	StatusPending = "pending"
	StatusDone    = "done"
	StatusError   = "error"
)

type Submission struct {
	Status  string `json:"status"`
	Verdict string `json:"verdict"`
}

type SubmissionResponse struct {
	Submission Submission `json:"submission"`
}

func (c *Client) Submit(courseID, questionID, compilerID int, sourceCode string) (int, error) {
	submitReq, err := c.NewRequest("POST", "/submissions", &SubmitRequest{
		CourseID: courseID, QuestionID: questionID, CompilerID: compilerID,
		SourceCode: sourceCode,
	})
	if err != nil {
		return 0, err
	}

	submitResp, err := c.Do[SubmitResponse](submitReq)
	if err != nil {
		return 0, err
	}

	return submitResp.ID, err
}

func (c *Client) Submission(submissionID int) (Submission, error) {
	var submission Submission

	endpoint := fmt.Sprintf("/submissions/%d", submissionID)
	submissionReq, err := c.NewRequest[any]("GET", endpoint, nil)
	if err != nil {
		return submission, err
	}

	submissionResp, err := c.Do[SubmissionResponse](submissionReq)
	if err != nil {
		return submission, err
	}

	submission = submissionResp.Submission
	return submission, err
}

func pollSubmission(c *Client, submissionId int) {
	const (
		INTERVAL = 1500 * time.Millisecond
		TIMEOUT  = 10 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	for {
		submission, err := c.Submission(submissionId)
		if err != nil {
			fmt.Println("Failed to retrieve submission:", err)
			fmt.Println("Retrying...")
		} else if submission.Status == StatusDone {
			fmt.Println("Done! Verdict:", submission.Verdict)
			return
		} else if submission.Status == StatusError {
			fmt.Println("Submission failed")
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("Polling timed out:", ctx.Err())
			return

		case <-ticker.C:
		}
	}
}

func indexProblemByCode(problems []Problem) map[string]int {
	m := make(map[string]int, len(problems))
	for _, problem := range problems {
		m[strings.ToLower(problem.Code)] = problem.ID
	}
	return m
}

func SubmitCommand(c *Client) {
	courseID, ok := c.GetCourse()
	if !ok {
		fmt.Println("Please select a course first")
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Please enter a file")
		return
	}

	problems, err := c.Problems(courseID)
	if err != nil {
		fmt.Println("Failed to retrieve problems:", err)
		return
	}
	problemMap := indexProblemByCode(problems)

	for _, path := range os.Args[2:] {
		srcBytes, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Failed to read file:", err)
			continue
		}
		src := string(srcBytes)

		lang, ok := DetectLanguage(path)
		if !ok {
			fmt.Println("No available language detected")
			continue
		}

		fileName := filepath.Base(path)
		problemCode, ok := ExtractFirstComment(src, lang)
		if !ok {
			problemCode = strings.TrimSuffix(fileName, filepath.Ext(fileName))
		}

		problemID, ok := problemMap[strings.ToLower(problemCode)]
		if !ok {
			fmt.Println("No problem with this code found")
			continue
		}

		compilerID := int(lang)
		submissionID, err := c.Submit(courseID, problemID, compilerID, src)
		if err != nil {
			fmt.Println("Failed to submit code:", err)
			continue
		}

		fmt.Println(fileName, "submitted! Retrieving status...")
		pollSubmission(c, submissionID)
	}
}
