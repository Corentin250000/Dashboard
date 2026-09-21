import "../auth.css"

function PendingVerificationPage() {
    return (
        <div className="auth-page">
            <h1 className="auth-title">Checks your mails!</h1>
            <div className="auth-field">
                <p> A verification email has been sent to you. 
                    Check your inbox and don’t forget to check your spam folder.
                </p>
            </div>
        </div>
    );
}

export default PendingVerificationPage;