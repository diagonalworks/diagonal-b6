import { AppStore, Scenario, appAtom, initialAppStore } from '@/atoms/app';
import { startupQueryAtom } from '@/atoms/startup';
import { $FixMe } from '@/utils/defs';
import { useAtom, useAtomValue } from 'jotai';
import { uniqueId } from 'lodash';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
} from 'react';
import { Comparator } from './comparator';
import { OutlinerStore } from './outliner';

/**
 * The app context that provides the app state and the methods to update it.
 */
export const AppContext = createContext<{
    app: AppStore;
    setApp: (fn: (draft: AppStore) => void) => void;
    setFixedOutliner: (id: keyof AppStore['outliners']) => void;
    setActiveOutliner: (
        id: keyof AppStore['outliners'],
        value: AppStore['outliners'][string]['active']
    ) => void;
    createOutliner: (outliner: OutlinerStore) => void;
    moveOutliner: (
        id: keyof AppStore['outliners'],
        dx: number,
        dy: number
    ) => void;
    closeOutliner: (id: keyof AppStore['outliners']) => void;
    changedWorldScenarios: Scenario[];
    addScenario: () => void;
    removeScenario: (id: string) => void;
    setActiveScenario: (id: string) => void;
    addComparator: (req: Comparator['request']) => void;
    activeComparator?: Comparator;
}>({
    app: initialAppStore,
    setApp: () => {},
    setFixedOutliner: () => {},
    setActiveOutliner: () => {},
    createOutliner: () => {},
    moveOutliner: () => {},
    closeOutliner: () => {},
    changedWorldScenarios: [],
    addScenario: () => {},
    removeScenario: () => {},
    setActiveScenario: () => {},
    addComparator: () => {},
});

/**
 *
 * Hook to access the app context.
 * Use this hook to access the app state and the methods to update it.
 */
export const useAppContext = () => {
    return useContext(AppContext);
};

/**
 * The app provider that provides the app context to the app.
 */
export const AppProvider = ({ children }: PropsWithChildren) => {
    const [app, setApp] = useAtom(appAtom);
    const startupQuery = useAtomValue(startupQueryAtom);

    const addComparator = useCallback(
        (request: Comparator['request']) => {
            if (!request) return;
            const id = uniqueId('comparator-');
            setApp((draft) => {
                draft.comparators[id] = {
                    id,
                    request,
                };
            });
        },
        [setApp]
    );

    const activeComparator = useMemo(() => {
        return Object.values(app.comparators).find((c) =>
            !c.request
                ? false
                : c.request.baseline === (app.tabs.left as $FixMe) &&
                  c.request.scenarios.includes(app.tabs?.right as $FixMe)
        );
    }, [app.comparators, app.tabs]);

    const createOutliner = useCallback(
        (outliner: OutlinerStore) => {
            setApp((draft) => {
                draft.outliners[outliner.id] = outliner;
            });
        },
        [setApp]
    );

    const closeOutliner = useCallback(
        (id: keyof AppStore['outliners']) => {
            setApp((draft) => {
                delete draft.outliners[id];
            });
        },
        [setApp]
    );

    useEffect(() => {
        setApp((draft) => {
            startupQuery.data?.docked?.forEach((d: $FixMe, i: number) => {
                draft.outliners[`docked-${i}`] = {
                    id: `docked-${i}`,
                    properties: {
                        scenario: 'baseline',
                        docked: true,
                        transient: false,
                        coordinates: { x: 0, y: 0 },
                    },
                    data: d,
                };
            });
        });
    }, [startupQuery.data?.docked]);

    const setFixedOutliner = useCallback(
        (id: keyof AppStore['outliners']) => {
            setApp((draft) => {
                draft.outliners[id].properties.transient = false;
            });
        },
        [setApp]
    );

    const moveOutliner = useCallback(
        (id: keyof AppStore['outliners'], dx: number, dy: number) => {
            setApp((draft) => {
                const { coordinates } = draft.outliners[id].properties;
                if (!coordinates) return;
                draft.outliners[id].properties.coordinates = {
                    x: coordinates.x + dx,
                    y: coordinates.y + dy,
                };
            });
        },
        [setApp]
    );

    const setActiveOutliner = useCallback(
        (
            id: keyof AppStore['outliners'],
            value: AppStore['outliners'][string]['active']
        ) => {
            setApp((draft) => {
                draft.outliners[id].active = value;
            });
        },
        [setApp]
    );

    const changedWorldScenarios = useMemo(() => {
        return Object.values(app.scenarios).filter((o) => o.id !== 'baseline');
    }, [app.scenarios]);

    const addScenario = useCallback(() => {
        const id = uniqueId('scenario-');
        setApp((draft) => {
            draft.scenarios[id] = {
                id: id,
                name: 'Untitled Scenario',
                worldId: undefined,
                change: {
                    features: [],
                    function: '',
                },
            };
            draft.tabs.right = id;
        });
    }, [setApp, changedWorldScenarios]);

    const setActiveScenario = useCallback(
        (id?: string) => {
            setApp((draft) => {
                draft.tabs.right = id;
            });
        },
        [setApp]
    );

    const removeScenario = useCallback(
        (id: string) => {
            setApp((draft) => {
                delete draft.scenarios[id];
                const newTab = changedWorldScenarios.find(
                    (s) => s.id !== id
                )?.id;
                draft.tabs.right = newTab;
            });
        },
        [setApp, changedWorldScenarios]
    );

    const value = {
        app,
        setApp,
        setActiveOutliner,
        setFixedOutliner,
        moveOutliner,
        closeOutliner,
        changedWorldScenarios,
        setActiveScenario,
        addScenario,
        removeScenario,
        addComparator,
        activeComparator,
        createOutliner,
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};
