import { useEffect, useState } from 'react';
import GoogleIcon from '../../../assets/icons/google.svg';
import SpotifyIcon from '../../../assets/icons/spotify.svg';
import steamIcon from '../../../assets/icons/steam.svg';
import { api, API_URL } from '../../shared/api';

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
        <div className="min-h-screen px-6 py-8">
            <h1 className="mb-6 text-center text-2xl font-medium text-(--text-primary)">Choose your subscriptions</h1>
            <div className="mx-auto grid max-w-270 grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">

                {/* GOOGLE */}
                <div className="flex aspect-square flex-col items-center justify-center gap-2 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                    <img className="h-15 w-15 object-contain" src={GoogleIcon} alt="google icon" />
                    <p className="text-[25px] font-medium text-(--text-primary)">Google</p>
                    <button
                        className={`mt-1.5 rounded-lg border border-(--border) px-3.5 py-1.5 text-[13px] font-medium disabled:opacity-100 ${googleConnected ? 'cursor-default bg-[mediumseagreen]' : 'cursor-pointer bg-(--accent)'} text-(--background)`}
                        type="button"
                        onClick={connectGoogle}
                        disabled={googleConnected}>
                        {googleConnected ? 'Connected' : 'Connect'}
                    </button>
                </div>

                {/* SPOTIFY */}
                <div className="flex aspect-square flex-col items-center justify-center gap-2 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                    <img className="h-15 w-15 object-contain" src={SpotifyIcon} alt="Spotify icon" />
                    <p className="text-[25px] font-medium text-(--text-primary)">Spotify</p>
                    <button
                        className={`mt-1.5 rounded-lg border border-(--border) px-3.5 py-1.5 text-[13px] font-medium disabled:opacity-100 ${spotifyConnected ? 'cursor-default bg-[mediumseagreen]' : 'cursor-pointer bg-(--accent)'} text-(--background)`}
                        type="button"
                        onClick={connectSpotify}
                        disabled={spotifyConnected}>
                        {spotifyConnected ? 'Connected' : 'Connect'}
                    </button>
                </div>

                {/* STEAM */}
                <div className="flex aspect-square flex-col items-center justify-center gap-2 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                    <img className="h-15 w-15 object-contain" src={steamIcon} alt="Steam icon" />
                    <p className="text-[25px] font-medium text-(--text-primary)">Steam</p>
                    <button
                        className={`mt-1.5 rounded-lg border border-(--border) px-3.5 py-1.5 text-[13px] font-medium disabled:opacity-100 ${steamConnected ? 'cursor-default bg-[mediumseagreen]' : 'cursor-pointer bg-(--accent)'} text-(--background)`}
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
