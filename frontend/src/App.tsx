import { ScenarioTab } from '@/components/ScenarioTab';
import { AppProvider, useAppContext } from '@/lib/context/app';
import { PlusIcon, ReaderIcon } from '@radix-ui/react-icons';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { Provider } from 'jotai';
import { queryClientAtom } from 'jotai-tanstack-query';
import { useHydrateAtoms } from 'jotai/react/utils';
import {
    HTMLAttributes,
    PropsWithChildren,
    useDeferredValue,
    useEffect,
    useState,
} from 'react';
import { MapProvider } from 'react-map-gl';
import { twMerge } from 'tailwind-merge';

import diagonalScenarioStyle from '@/components/diagonal-map-style-orange.json';
import diagonalBasemapStyle from '@/components/diagonal-map-style.json';
import { StyleSpecification } from 'maplibre-gl';
import { Scenario } from './atoms/app';
import { ScenarioProvider } from './lib/context/scenario';

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
    const {
        app: { tabs },
    } = useAppContext();
    return (
        <div className="h-screen max-h-screen flex flex-col">
            <Tabs />
            <div className="flex-grow">
                {tabs.left && (
                    <ScenarioProvider id={tabs.left} tab="left">
                        <ScenarioTab
                            id={tabs.left}
                            className={twMerge(
                                tabs.right && 'w-1/2 inline-block'
                            )}
                            mapStyle={
                                diagonalBasemapStyle as StyleSpecification
                            }
                        />
                    </ScenarioProvider>
                )}
                {tabs.right && (
                    <ScenarioProvider id={tabs.right} tab="right">
                        <ScenarioTab
                            id={tabs.right}
                            className="w-1/2 inline-block"
                            mapStyle={
                                diagonalScenarioStyle as StyleSpecification
                            }
                        />
                    </ScenarioProvider>
                )}
            </div>
        </div>
    );
};

const Tabs = () => {
    const {
        setApp,
        app: { scenarios, tabs },
    } = useAppContext();

    const handleAddScenario = () => {
        const id = `scenario-${scenarios.length}`;
        setApp((draft) => {
            draft.scenarios[id] = {
                id: id,
                name: 'Untitled Scenario',
            };
            if (!draft.tabs.right) draft.tabs.right = id;
        });
    };

    return (
        <div className="w-full px-1 pt-2">
            <div
                className={twMerge(
                    tabs?.right ? 'grid grid-cols-2' : 'grid grid-cols-1'
                )}
            >
                <div className="flex items-end justify-between">
                    <TabButton scenario={scenarios[tabs.left]} />
                    {!tabs.right && (
                        <button
                            onClick={handleAddScenario}
                            aria-label="add scenario"
                            className="text-sm flex gap-2 items-center bg-orange-10 rounded w-fit border border-b-0 hover:bg-orange-20 rounded-b-none border-orange-30 text-orange-60 px-2 py-1"
                        >
                            <PlusIcon />
                            scenario
                        </button>
                    )}
                </div>
                {tabs.right && (
                    <TabButton
                        scenario={scenarios[tabs.right]}
                        className=" bg-orange-10 border-orange-30"
                        editable
                    />
                )}
            </div>
        </div>
    );
};

const TabButton = ({
    editable = false,
    scenario,
    ...props
}: {
    editable?: boolean;
    scenario: Scenario;
} & HTMLAttributes<HTMLDivElement>) => {
    const { setApp } = useAppContext();

    const [inputValue, setInputValue] = useState(scenario.name);
    const deferredInput = useDeferredValue(inputValue);

    useEffect(() => {
        setApp((draft) => {
            draft.scenarios[scenario.id].name = deferredInput;
        });
    }, [deferredInput]);

    const handleInputChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(evt.target.value);
    };

    return (
        <div
            {...props}
            className={twMerge(
                'text-sm w-fit flex gap-2 items-center bg-graphite-20 rounded rounded-b-none border border-b-0 border-graphite-30 px-2 py-1',
                props.className
            )}
        >
            <ReaderIcon />
            <input
                onChange={handleInputChange}
                disabled={!editable}
                className="bg-transparent border-none text-sm focus:outline-none focus:text-graphite-80 transition-colors  caret-orange-60 "
                value={scenario.name}
                //onClick={props.onClick}
            />
        </div>
    );
};

export default App;
