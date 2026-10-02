import GoogleIcon from '../../../assets/icons/google.svg';
import SncfIcon from '../../../assets/icons/SNCF.svg';
import '../subs.css'

// type Sub = {
//     id: string;
//     name: string;
//     connected: boolean;
// };

function SubscriptionPage() {
    return (
        <div className='subs-page'>
            <h1 className='subs-title'>Choose your subscriptions</h1>
            <div className='subs-grid'>

                {/* GOOGLE */}
                <div className='subs-card'>
                    <img className="sub-icon" src={GoogleIcon} alt="google icon" />
                    <p className='subs-name'>Google</p>
                    <button className='subs-connect'>Connect</button>
                </div>

                {/* SNCF */}
                <div className='subs-card'>
                     <img className="sub-icon" src={SncfIcon} alt="SNCF icon" />
                    <p className='subs-name'>SNCF</p>
                    <button className='subs-connect'>Connect</button>
                </div>
            </div>
        </div>
    );
}

export default SubscriptionPage;