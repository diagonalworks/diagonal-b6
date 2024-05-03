import { AppStore, Scenario, appAtom, initialAppStore } from '@/atoms/app';
import { startupQueryAtom } from '@/atoms/startup';
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
import { OutlinerSpec, OutlinerStore } from './outliner';

/**
 * The app context that provides the app state and the methods to update it.
 */
export const AppContext = createContext<{
    app: AppStore;
    setApp: (fn: (draft: AppStore) => void) => void;
    createOutliner: (outliner: OutlinerStore) => void;
    setFixedOutliner: (id: keyof AppStore['outliners']) => void;
    setActiveOutliner: (
        id: keyof AppStore['outliners'],
        value: AppStore['outliners'][string]['active']
    ) => void;
    moveOutliner: (
        id: keyof AppStore['outliners'],
        dx: number,
        dy: number
    ) => void;
    closeOutliner: (id: keyof AppStore['outliners']) => void;
    changedWorldScenarios: Scenario[];
    addScenario: () => void;
    removeScenario: (id: string) => void;
}>({
    app: initialAppStore,
    setApp: () => {},
    createOutliner: () => {},
    setFixedOutliner: () => {},
    setActiveOutliner: () => {},
    moveOutliner: () => {},
    closeOutliner: () => {},
    changedWorldScenarios: [],
    addScenario: () => {},
    removeScenario: () => {},
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
            startupQuery.data?.docked?.forEach((d, i) => {
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

    const _removeTransientStacks = useCallback(() => {
        setApp((draft) => {
            for (const id in draft.outliners) {
                if (
                    draft.outliners[id].properties.transient &&
                    !draft.outliners[id].properties.docked
                ) {
                    delete draft.outliners[id];
                }
            }
        });
    }, [setApp]);

    const createOutliner = useCallback(
        (outliner: OutlinerSpec) => {
            _removeTransientStacks();
            setApp((draft) => {
                draft.outliners[outliner.id] = outliner;
            });
        },
        [setApp]
    );

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
            };
            if (!draft.tabs.right) draft.tabs.right = id;
        });
    }, [setApp, changedWorldScenarios]);

    const removeScenario = useCallback(
        (id: string) => {
            setApp((draft) => {
                delete draft.scenarios[id];
            });
        },
        [setApp]
    );

    const value = {
        app,
        setApp,
        createOutliner,
        setActiveOutliner,
        setFixedOutliner,
        moveOutliner,
        closeOutliner,
        changedWorldScenarios,
        addScenario,
        removeScenario,
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};
