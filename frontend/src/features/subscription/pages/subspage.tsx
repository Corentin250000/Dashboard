import { useEffect, useState } from 'react';
import GoogleIcon from '../../../assets/icons/google.svg';
import SpotifyIcon from '../../../assets/icons/spotify.svg';
import steamIcon from '../../../assets/icons/steam.svg';
import { api, API_URL } from '../../shared/api';
import '../subs.css';

// type Sub = {
//     id: string;
//     name: string;
//     connected: boolean;
// };

function SubscriptionPage() {
    const [spotifyConnected, setSpotifyConnected] = useState(false);
    const [googleConnected, setGoogleConnected] = useState(false);
    const [steamConnected, setSteamConnected] = useState(false);

    useEffect(() => {
        api.getServices()
            .then(({ services }) => {
                const spotify = services.find((service) => service.name === 'spotify');
                setSpotifyConnected(spotify?.connected ?? false);
                const google = services.find((service) => service.name === 'google');
                setGoogleConnected(google?.connected ?? false);
                const steam = services.find((service) => service.name === 'steam');
                setSteamConnected(steam?.connected ?? false);
            })
            .catch(() => {
                setSpotifyConnected(false);
                setGoogleConnected(false);
                setSteamConnected(false);
            });
    }, []);

    const connectSpotify = () => {
        window.location.assign(`${API_URL}/api/me/services/spotify/connect`);
    };

    const connectGoogle = () => {
        window.location.assign(`${API_URL}/api/me/services/google/connect`);
    };

    const connectSteam = () => {
        window.location.assign(`${API_URL}/api/me/services/steam/connect`);
    };

    return (
        <div className='subs-page'>
            <h1 className='subs-title'>Choose your subscriptions</h1>
            <div className='subs-grid'>

                {/* GOOGLE */}
                <div className='subs-card'>
                    <img className="sub-icon" src={GoogleIcon} alt="google icon" />
                    <p className='subs-name'>Google</p>
                    <button
                        className={googleConnected ? 'subs-connected' : 'subs-connect'}
                        type="button"
                        onClick={connectGoogle}
                        disabled={googleConnected}>
                        {googleConnected ? 'Connected' : 'Connect'}
                    </button>
                </div>

                {/* SPOTIFY */}
                <div className='subs-card'>
                    <img className="sub-icon" src={SpotifyIcon} alt="Spotify icon" />
                    <p className='subs-name'>Spotify</p>
                    <button
                        className={spotifyConnected ? 'subs-connected' : 'subs-connect'}
                        type="button"
                        onClick={connectSpotify}
                        disabled={spotifyConnected}>
                        {spotifyConnected ? 'Connected' : 'Connect'}
                    </button>
                </div>

                {/* STEAM */}
                <div className='subs-card'>
                    <img className="sub-icon" src={steamIcon} alt="Steam icon" />
                    <p className='subs-name'>Steam</p>
                    <button
                        className={steamConnected ? 'subs-connected' : 'subs-connect'}
                        type="button"
                        onClick={connectSteam}
                        disabled={steamConnected}>
                        {steamConnected ? 'Connected' : 'Connect'}
                    </button>
                </div>
            </div>
        </div>
    );
}

export default SubscriptionPage;
