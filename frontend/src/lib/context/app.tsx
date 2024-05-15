import { AppStore, Scenario, appAtom, initialAppStore } from '@/atoms/app';
import { startupQueryAtom } from '@/atoms/startup';
import {
    EvaluateRequestProto,
    EvaluateResponseProto,
    FeatureIDProto,
} from '@/types/generated/api';
import { $FixMe } from '@/utils/defs';
import { useQuery } from '@tanstack/react-query';
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
import { b6 } from '../b6';
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
    changes: Array<{ label?: string; id: FeatureIDProto }>;
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
    changes: [],
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

    const changesQuery = useQuery<EvaluateResponseProto, Error>({
        queryKey: [
            'evaluate',
            'expressions',
            JSON.stringify(startupQuery.data?.root),
        ],
        queryFn: () => {
            if (!startupQuery.data?.root) return Promise.resolve({ data: {} });
            return b6.evaluate({
                root: startupQuery.data?.root,
                request: {
                    call: {
                        function: {
                            symbol: 'list-feature',
                        },
                        args: [
                            {
                                literal: {
                                    featureIDValue: {
                                        type: 'FeatureTypeCollection',
                                        namespace:
                                            'diagonal.works/skyline-demo-05-2024',
                                        value: 1,
                                    },
                                },
                            },
                        ],
                    },
                },
            } as unknown as EvaluateRequestProto);
        },
    });

    const changes = useMemo(() => {
        const changes = changesQuery.data?.result?.literal?.collectionValue;
        if (!changes) return [];
        return (
            changes.values?.flatMap((v, i) => {
                if (!v.featureIDValue || !changes.keys?.[i].stringValue)
                    return [];
                return {
                    label: changes.keys?.[i].stringValue,
                    id: v.featureIDValue,
                };
            }) ?? []
        );
    }, [changesQuery.data]);

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
                  c.request.scenarios?.includes(app.tabs?.right as $FixMe)
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
        const id = uniqueId();
        setApp((draft) => {
            draft.scenarios[id] = {
                id: id,
                name: 'Untitled Scenario',
            };
            draft.tabs.right = id;
        });
    }, [setApp, changedWorldScenarios, startupQuery.data?.root?.namespace]);

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
        changes,
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};
