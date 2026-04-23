package mailer

import "fmt"

func VerifyAccountBody(username, verifyURL string) string {
	return fmt.Sprintf(`
			<h2>Hi %s,</h2>
			<p>Please verify your account by clicking the link below:</p>
			<a href="%s">Verify Account</a>
			<p>Link expires in 24 hours.</p>
			`, username, verifyURL,
	)
}

func ForgotPasswordBody(username, resetURL string) string {
	return fmt.Sprintf(`
			<h2>Hi %s,</h2>
			<p>Click the link below to reset your password:</p>
			<a href="%s">Reset Password</a>
			<p>Link expires in 1 hour.</p>
			`, username, resetURL,
	)
}
