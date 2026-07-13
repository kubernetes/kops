package linodego

import (
	"context"
)

type SecurityQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Response string `json:"response"`
}

type SecurityQuestionsListResponse struct {
	SecurityQuestions []SecurityQuestion `json:"security_questions"`
}

type SecurityQuestionsAnswerQuestion struct {
	QuestionID int    `json:"question_id"`
	Response   string `json:"response"`
}

type SecurityQuestionsAnswerOptions struct {
	SecurityQuestions []SecurityQuestionsAnswerQuestion `json:"security_questions"`
}

// SecurityQuestionsList returns a collection of security questions and their responses, if any, for your User Profile.
func (c *Client) SecurityQuestionsList(ctx context.Context) (*SecurityQuestionsListResponse, error) {
	return doGETRequest[SecurityQuestionsListResponse](ctx, c, "profile/security-questions")
}

// SecurityQuestionsAnswer adds security question responses for your User.
func (c *Client) SecurityQuestionsAnswer(ctx context.Context, opts SecurityQuestionsAnswerOptions) error {
	return doPOSTRequestNoResponseBody(ctx, c, "profile/security-questions", opts)
}
