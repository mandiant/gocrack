package notifications

import "html/template"

type templateCrackedPassword struct {
	TaskID    string
	TaskName  string
	CaseCode  string
	Email     string
	PublicURL string
}

type templateTaskStatusChanged struct {
	TaskName  string
	TaskID    string
	NewStatus string
	Email     string
	PublicURL string
	CaseCode  string
}

var emailCrackedTemplate = template.Must(template.New("cracked_password").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
        <p>Your task, {{.TaskName}} (Case: {{.CaseCode}}) has new password(s) that have recently cracked. You may view the results <a href="{{.PublicURL}}/tasks/details/{{.TaskID}}">here</a>.</p>
        <p>Sincerely,<br /> Your friendly neighborhood password cracking server.</p>
	</body>
</html>`))

var emailStatusChanged = template.Must(template.New("task_status_changed").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
		<p>Your task, {{.TaskName}} (Case: {{.CaseCode}}) changed status to {{.NewStatus}}. You may view your task <a href="{{.PublicURL}}/tasks/details/{{.TaskID}}">here</a>.</p>
		<p>Sincerely,<br /> Your friendly neighborhood password cracking server.</p>
	</body>
</html>`))
