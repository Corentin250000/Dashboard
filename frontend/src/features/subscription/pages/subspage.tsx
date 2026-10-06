import GoogleIcon from '../../../assets/icons/google.svg';
import SncfIcon from '../../../assets/icons/SNCF.svg';

// type Sub = {
//     id: string;
//     name: string;
//     connected: boolean;
// };

function SubscriptionPage() {
    return (
        <div className="min-h-screen px-6 py-8">
            <h1 className="mb-6 text-center text-2xl font-medium text-(--text-primary)">Choose your subscriptions</h1>
            <div className="mx-auto grid max-w-270 grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">

                {/* GOOGLE */}
                <div className="flex aspect-square flex-col items-center justify-center gap-2 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                    <img className="h-10 w-15" src={GoogleIcon} alt="google icon" />
                    <p className="text-[25px] font-medium text-(--text-primary)">Google</p>
                    <button type="button" className="mt-1.5 cursor-pointer rounded-lg border border-(--border) bg-(--accent) px-3.5 py-1.5 text-[13px] font-medium text-(--background)">Connect</button>
                </div>

                {/* SNCF */}
                <div className="flex aspect-square flex-col items-center justify-center gap-2 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                    <img className="h-10 w-15" src={SncfIcon} alt="SNCF icon" />
                    <p className="text-[25px] font-medium text-(--text-primary)">SNCF</p>
                    <button type="button" className="mt-1.5 cursor-pointer rounded-lg border border-(--border) bg-(--accent) px-3.5 py-1.5 text-[13px] font-medium text-(--background)">Connect</button>
                </div>
            </div>
        </div>
    );
}

export default SubscriptionPage;
