import { ScenarioTab } from '@/components/ScenarioTab';
import { AppProvider, useAppContext } from '@/lib/context/app';
import { Cross1Icon, PlusIcon, ReaderIcon } from '@radix-ui/react-icons';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { Provider } from 'jotai';
import { queryClientAtom } from 'jotai-tanstack-query';
import { useHydrateAtoms } from 'jotai/react/utils';
import {
    HTMLAttributes,
    PropsWithChildren,
    useCallback,
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
                            tab="left"
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
                            tab="right"
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
        app: { scenarios, tabs },
        changedWorldScenarios,
        addScenario,
    } = useAppContext();

    return (
        <div className="w-full px-1 pt-2 z-50 -mb-[1px]">
            <div
                className={twMerge(
                    tabs?.right ? 'grid grid-cols-2' : 'grid grid-cols-1'
                )}
            >
                <div className="flex items-end justify-between">
                    <TabButton
                        scenario={scenarios[tabs.left]}
                        tab="left"
                        active
                    />
                    {!tabs.right && (
                        <button
                            onClick={addScenario}
                            aria-label="add scenario"
                            className="text-sm flex gap-2 mb-[1px] items-center bg-orange-10 rounded w-fit border border-b-0 hover:bg-orange-20 rounded-b-none border-orange-30 text-orange-60 px-2 py-1"
                        >
                            <PlusIcon />
                            scenario
                        </button>
                    )}
                </div>
                {tabs.right && (
                    <div className="flex gap-1">
                        {changedWorldScenarios.map((scenario) => (
                            <TabButton
                                scenario={scenario}
                                editable
                                tab="right"
                                active={tabs.right === scenario.id}
                            />
                        ))}
                        <button
                            className="bg-orange-10 hover:bg-orange-20  border border-b border-b-orange-40 border-orange-30 text-orange-70 hover:text-orange-90 px-2 rounded-t"
                            aria-label="create new scenario"
                            onClick={addScenario}
                        >
                            <PlusIcon />
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

const TabButton = ({
    editable = false,
    scenario,
    tab,
    active = false,
    ...props
}: {
    editable?: boolean;
    scenario: Scenario;
    tab?: 'left' | 'right';
    active?: boolean;
} & HTMLAttributes<HTMLDivElement>) => {
    const { setApp, removeScenario, setActiveScenario } = useAppContext();

    const [inputValue, setInputValue] = useState(scenario.name);
    const deferredInput = useDeferredValue(inputValue);
    const [isHovered, setIsHovered] = useState(false);

    useEffect(() => {
        setApp((draft) => {
            draft.scenarios[scenario.id].name = deferredInput;
        });
    }, [deferredInput]);

    const handleInputChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(evt.target.value);
    };

    const handleDeleteScenario = useCallback(() => {
        removeScenario(scenario.id);
    }, [removeScenario, scenario.id]);

    const handleClick = () => {
        setActiveScenario(scenario.id);
    };

    return (
        <div
            {...props}
            className={twMerge(
                'text-sm w-fit border-b-2  flex gap-2  items-center transition-colors bg-graphite-20 rounded rounded-b-none border  border-graphite-40 px-2 py-1',
                tab === 'right' && 'bg-orange-20 border-orange-40',
                active &&
                    (tab === 'right'
                        ? 'border-b-orange-30'
                        : 'border-b-graphite-30'),
                tab === 'right' ? 'hover:bg-orange-30' : 'hover:bg-graphite-30',
                active && (tab === 'right' ? 'bg-orange-30' : 'bg-graphite-30'),
                props.className
            )}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            onClick={handleClick}
        >
            <ReaderIcon />
            <input
                onChange={handleInputChange}
                disabled={!editable || !active}
                className="bg-transparent border-none text-sm focus:outline-none focus:text-graphite-80 transition-colors  caret-orange-60 "
                value={scenario.name}
            />

            {scenario.id !== 'baseline' && (
                <div className="w-4 flex">
                    <button
                        aria-label="close tab "
                        onClick={handleDeleteScenario}
                        className={twMerge(
                            'text-orange-70 hover:text-orange-90 transition-colors',
                            !isHovered && ' hidden ',
                            isHovered && ' visible'
                        )}
                    >
                        <Cross1Icon />
                    </button>
                </div>
            )}
        </div>
    );
};

export default App;
