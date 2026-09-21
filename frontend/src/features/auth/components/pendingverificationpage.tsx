import "../auth.css"

function PendingVerificationPage() {
    return (
        <div className="auth-page">
            <div className="auth-card">
                <h1 className="auth-title">Check your emails!</h1>
                <div className="auth-field">
                    <p> A verification email has been sent to you.</p>
                    <p>Check your inbox and don’t forget to check your spam folder.</p>
                </div>
            </div>
        </div>
    );
}

export default PendingVerificationPage;