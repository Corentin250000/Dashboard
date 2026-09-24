package mail

import (
	"bytes"
	"html/template"

	"dashboard-backend/internal/theme"
)

const confirmationRegistration = `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>Confirm your email</title>
	</head>
	<body style="margin:0; padding:0; background-color:{{.Background}}; font-family:Arial, Helvetica, sans-serif;">
		<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:{{.Background}}; padding:40px 0;">
			<tr>
				<td align="center">
					<table role="presentation" width="480" cellpadding="0" cellspacing="0" style="background-color:{{.Card}}; border-radius:8px; overflow:hidden; box-shadow:0 2px 6px rgba(0,0,0,0.05);">
						<tr>
							<td style="background-color:{{.Accent}}; padding:24px; text-align:center;">
								<h1 style="color:#ffffff; margin:0; font-size:20px;">Dashboard</h1>
							</td>
						</tr>
						<tr>
							<td style="padding:32px 24px;">
								<h2 style="color:{{.TextPrimary}}; font-size:20px; margin-top:0;">Confirm your email address</h2>
								<p style="color:{{.TextSecondary}}; font-size:15px; line-height:1.5;">
									Thanks for signing up! Please confirm your email address by clicking the button below.
									This link will expire in 24 hours.
								</p>
								<table role="presentation" cellpadding="0" cellspacing="0" style="margin:32px auto;">
									<tr>
										<td style="border-radius:6px; background-color:{{.Accent}};">
											<a href="{{.ConfirmationLink}}"
											style="display:inline-block; padding:12px 28px; font-size:15px; color:#ffffff; text-decoration:none; font-weight:bold; border-radius:6px;">
												Confirm Email
											</a>
										</td>
									</tr>
								</table>
								<p style="color:{{.TextSecondary}}; font-size:13px; line-height:1.5;">
									If the button doesn't work, copy and paste this link into your browser:<br />
									<a href="{{.ConfirmationLink}}" style="color:{{.Accent}}; word-break:break-all;">{{.ConfirmationLink}}</a>
								</p>
								<p style="color:#9ca3af; font-size:13px; margin-top:32px;">
									If you didn't create an account, you can safely ignore this email.
								</p>
							</td>
						</tr>
						<tr>
							<td style="background-color:#f9fafb; padding:16px 24px; text-align:center;">
								<p style="color:#9ca3af; font-size:12px; margin:0;">
									&copy; 2026 Dashboard. All rights reserved.
								</p>
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</body>
</html>`

type confirmationData struct {
	Background       string
	Card             string
	TextPrimary      string
	TextSecondary    string
	Accent           string
	ConfirmationLink string
}

var confirmationTemplate = template.Must(template.New("confirmation").Parse(confirmationRegistration))

func RenderConfirmationEmail(link string) (string, error) {
	var buf bytes.Buffer
	data := confirmationData{
		Background:       theme.ColorBackground,
		Card:             theme.ColorCard,
		TextPrimary:      theme.ColorTextPrimary,
		TextSecondary:    theme.ColorTextSecondary,
		Accent:           theme.ColorAccent,
		ConfirmationLink: link,
	}
	if err := confirmationTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
