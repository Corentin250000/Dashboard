import './dashboard.css'
import SncfWidget from './widgets/sncf/SncfWidget';

function Dashboard() {
    return (
        <SncfWidget station='Mulhouse'/>
    );
}

export default Dashboard