import { AppStore, appAtom, initialAppStore } from '@/atoms/app';
import { startupQueryAtom } from '@/atoms/startup';
import { MapLayerProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { useAtom, useAtomValue } from 'jotai';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
} from 'react';
import { MapRef } from 'react-map-gl/maplibre';
import { OutlinerSpec, OutlinerStore } from './outliner';

/**
 * The app context that provides the app state and the methods to update it.
 */
export const AppContext = createContext<{
    app: AppStore;
    setApp: (fn: (draft: AppStore) => void) => void;
    createOutliner: (outliner: OutlinerStore) => void;
    draggableOutliners: OutlinerStore[];
    dockedOutliners: OutlinerStore[];
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
    getVisibleMarkers: (map: MapRef) => $FixMe[];
    queryLayers: Array<{
        layer: MapLayerProto;
        histogram: OutlinerStore['histogram'];
    }>;
    closeOutliner: (id: keyof AppStore['outliners']) => void;
}>({
    app: initialAppStore,
    setApp: () => {},
    createOutliner: () => {},
    draggableOutliners: [],
    dockedOutliners: [],
    setFixedOutliner: () => {},
    setActiveOutliner: () => {},
    moveOutliner: () => {},
    getVisibleMarkers: () => [],
    closeOutliner: () => {},
    queryLayers: [],
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
                        tab: 'baseline',
                        docked: true,
                        transient: false,
                        coordinates: { x: 0, y: 0 },
                    },
                    data: d,
                };
            });
        });
    }, [startupQuery.data?.docked]);

    const dockedOutliners = useMemo(() => {
        return Object.values(app.outliners).filter(
            (outliner) => outliner.properties.docked
        );
    }, [app.outliners]);

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

    const queryLayers = useMemo(() => {
        return Object.values(app.outliners).flatMap((outliner) => {
            return (
                outliner.data?.proto.layers?.map((l) => ({
                    layer: l,
                    histogram: outliner.histogram,
                })) || []
            );
        });
    }, [app.outliners]);

    const getVisibleMarkers = useCallback(
        (map: MapRef) => {
            const features = Object.values(app.outliners)
                .flatMap((outliner) => outliner.data?.geoJSON || [])
                .flat()
                .filter((f: $FixMe) => {
                    f.geometry.type === 'Point' &&
                        map
                            ?.getBounds()
                            ?.contains(
                                f.geometry.coordinates as [number, number]
                            );
                    return true;
                });
            return features;
        },
        [app.outliners]
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

    const draggableOutliners = useMemo(() => {
        return Object.values(app.outliners).filter(
            (outliner) => !outliner.properties.docked
        );
    }, [app.outliners]);

    const value = {
        app,
        setApp,
        createOutliner,
        draggableOutliners,
        dockedOutliners,
        setActiveOutliner,
        setFixedOutliner,
        moveOutliner,
        getVisibleMarkers,
        queryLayers,
        closeOutliner,
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};
