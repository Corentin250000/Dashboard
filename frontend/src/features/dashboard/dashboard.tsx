import './dashboard.css'
import SncfWidget from './widgets/sncf/SncfWidget';

function Dashboard() {
    return (
        <SncfWidget station='Marseille Saint-Charles'/>
    );
}

export default Dashboard