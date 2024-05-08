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
}>({
    app: initialAppStore,
    setApp: () => {},
    setFixedOutliner: () => {},
    setActiveOutliner: () => {},
    moveOutliner: () => {},
    closeOutliner: () => {},
    changedWorldScenarios: [],
    addScenario: () => {},
    removeScenario: () => {},
    setActiveScenario: () => {},
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
        console.log('changedWorldScenarios');
        return Object.values(app.scenarios).filter((o) => o.id !== 'baseline');
    }, [app.scenarios]);

    const addScenario = useCallback(() => {
        const id = uniqueId('scenario-');
        setApp((draft) => {
            draft.scenarios[id] = {
                id: id,
                name: 'Untitled Scenario',
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
                console.log('removeScenario', id);
                delete draft.scenarios[id];
                const newTab = changedWorldScenarios.find(
                    (s) => s.id !== id
                )?.id;
                setActiveScenario(newTab);
                console.log('newTab', newTab);
            });
        },
        [setApp, changedWorldScenarios, setActiveScenario]
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
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};
