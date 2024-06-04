import Workspace from '@/components/Workspace';

import { AppProvider } from './provider';

export function App() {
    return (
        <AppProvider>
            <Workspace />
        </AppProvider>
    );
}
