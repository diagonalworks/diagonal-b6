import { ScenarioTab } from '@/components/ScenarioTab';
import { AppProvider, useAppContext } from '@/lib/context/app';
import { ReaderIcon } from '@radix-ui/react-icons';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { Provider } from 'jotai';
import { queryClientAtom } from 'jotai-tanstack-query';
import { useHydrateAtoms } from 'jotai/react/utils';
import { PropsWithChildren } from 'react';
import { MapProvider } from 'react-map-gl';
import { twMerge } from 'tailwind-merge';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: Infinity,
        },
    },
});

const HydrateAtoms = ({ children }: PropsWithChildren) => {
    useHydrateAtoms([[queryClientAtom, queryClient]]);
    return children;
};

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <Provider>
                <HydrateAtoms>
                    <MapProvider>
                        <AppProvider>
                            <Workspace />
                        </AppProvider>
                    </MapProvider>
                </HydrateAtoms>
                <ReactQueryDevtools initialIsOpen={false} />
            </Provider>
        </QueryClientProvider>
    );
}

const Workspace = () => {
    return (
        <div className="h-screen max-h-screen flex flex-col">
            <Tabs />
            <div className="flex-grow">
                <ScenarioTab id="baseline" />
            </div>
        </div>
    );
};

const Tabs = () => {
    const {
        app: { scenarios, tabs },
    } = useAppContext();

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
