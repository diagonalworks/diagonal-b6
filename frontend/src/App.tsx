import { ReaderIcon } from '@radix-ui/react-icons';
import { useAtomValue } from 'jotai';
import { MapProvider } from 'react-map-gl';
import { twMerge } from 'tailwind-merge';
import { appAtom } from './atoms/app';
import { Map } from './components/Map';

function App() {
    return (
        <MapProvider>
            <div className="h-screen flex flex-col">
                <Tabs />
                <div className="flex-grow">
                    <Map id="baseline" />
                </div>
            </div>
        </MapProvider>
    );
}

const Tabs = () => {
    const { scenarios, tabs } = useAtomValue(appAtom);

    return (
        <div className="w-full px-1 pt-2">
            <div
                className={twMerge(
                    tabs?.right ? 'grid grid-cols-2' : 'grid grid-cols-1'
                )}
            >
                <div className="text-sm bg-graphite-20 rounded w-fit flex gap-2 items-center border rounded-b-none border-graphite-30 px-2 py-1">
                    <ReaderIcon />
                    {scenarios[tabs.left].name}
                </div>
            </div>
        </div>
    );
};

export default App;
