import { ReaderIcon } from '@radix-ui/react-icons';
import {
    QueryClient,
    QueryClientProvider,
    useQuery,
} from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { useEffect } from 'react';
import { MapProvider } from 'react-map-gl';
import { twMerge } from 'tailwind-merge';
import { appAtom, collectionAtom } from './atoms/app';
import { viewAtom } from './atoms/location';
import { Map } from './components/Map';
import { StartupResponse } from './types/startup';

const queryClient = new QueryClient();

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <MapProvider>
                <Workspace />
            </MapProvider>
            <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
    );
}

const Workspace = () => {
    const [viewState, setViewState] = useAtom(viewAtom);
    const setApp = useSetAtom(appAtom);
    const collection = useAtomValue(collectionAtom);

    const startup = useQuery({
        queryKey: ['startup', collection],
        queryFn: () =>
            fetch(
                '/api/startup?' +
                    new URLSearchParams({
                        z: viewState.zoom.toString(),
                        ll: `${viewState.latitude},${viewState.longitude}`,
                        r: collection,
                    })
            ).then((res) => res.json() as Promise<StartupResponse>),
    });

    useEffect(() => {
        if (startup.data) {
            setViewState({
                ...viewState,
                ...(startup.data.mapCenter && {
                    latitude: startup.data.mapCenter.LatE7 / 1e7,
                    longitude: startup.data.mapCenter.LngE7 / 1e7,
                }),
                ...(startup.data.mapZoom && { zoom: startup.data.mapZoom }),
            });
            setApp((draft) => {
                draft.session = startup.data.session;
            });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [startup.data]);

    return (
        <div className="h-screen flex flex-col">
            <Tabs />
            <div className="flex-grow">
                <Map id="baseline" />
            </div>
        </div>
    );
};

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
