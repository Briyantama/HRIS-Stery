package infrastructure

import (
	"fmt"
	"strings"
)

// TemplateEngine renders notification templates with variable substitution
type TemplateEngine struct {
	templates map[string]Template
}

type Template struct {
	Subject      string
	Body         string
	RequiredVars []string
}

// NewTemplateEngine creates a new template engine with Phase 1 templates
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: map[string]Template{
			"leave.requested": {
				Subject:      "Leave Request Submitted",
				Body:         "Your leave request for {{start_date}} to {{end_date}} has been submitted for approval.",
				RequiredVars: []string{"start_date", "end_date"},
			},
			"leave.approved": {
				Subject:      "Leave Request Approved",
				Body:         "Your leave request for {{start_date}} to {{end_date}} has been approved.",
				RequiredVars: []string{"start_date", "end_date"},
			},
			"leave.rejected": {
				Subject:      "Leave Request Rejected",
				Body:         "Your leave request for {{start_date}} to {{end_date}} has been rejected. Reason: {{reason}}",
				RequiredVars: []string{"start_date", "end_date", "reason"},
			},
			"employee.welcome": {
				Subject:      "Welcome to HRIS Stery",
				Body:         "Welcome {{first_name}} {{last_name}}! Your account has been created. You can log in with: {{email}}",
				RequiredVars: []string{"first_name", "last_name", "email"},
			},
			"employee.termination": {
				Subject:      "Employee Termination",
				Body:         "Employee {{first_name}} {{last_name}} has been terminated.",
				RequiredVars: []string{"first_name", "last_name"},
			},
			"auth.activation": {
				Subject:      "Account Activation",
				Body:         "Welcome {{first_name}} {{last_name}}! Please verify your email: {{email}}",
				RequiredVars: []string{"first_name", "last_name", "email"},
			},
		},
	}
}

// Render renders a template with the given variables
func (te *TemplateEngine) Render(templateKey string, variables map[string]string) (subject, body string, err error) {
	template, exists := te.templates[templateKey]
	if !exists {
		return "", "", fmt.Errorf("template not found: %s", templateKey)
	}

	// Validate required variables
	for _, requiredVar := range template.RequiredVars {
		if _, exists := variables[requiredVar]; !exists {
			return "", "", fmt.Errorf("missing required variable: %s", requiredVar)
		}
	}

	// Substitute variables
	subject = te.substituteVariables(template.Subject, variables)
	body = te.substituteVariables(template.Body, variables)

	return subject, body, nil
}

func (te *TemplateEngine) substituteVariables(text string, variables map[string]string) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
