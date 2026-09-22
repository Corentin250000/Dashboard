import '../auth.css'

function ConfirmAuthentificationPage() {
    return(
        <div className='auth-page'>
            <div className='auth-card'>
                <h1 className='auth-title'> Authentication Successful </h1>
                <div className='auth-field'>
                    <p>You have been successfully authenticated. You will be redirected shortly.</p>
                </div>
            </div>
        </div>
    );
}

export default ConfirmAuthentificationPage;