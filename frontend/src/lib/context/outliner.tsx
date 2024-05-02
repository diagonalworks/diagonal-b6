import { viewAtom } from '@/atoms/location';
import { startupQueryAtom } from '@/atoms/startup';
import { Event } from '@/types/events';
import { FeatureIDProto, NodeProto } from '@/types/generated/api';
import { Chip, StackResponse } from '@/types/stack';
import { $FixMe } from '@/utils/defs';
import { useQuery } from '@tanstack/react-query';
import { ScaleOrdinal } from 'd3-scale';
import { useAtomValue } from 'jotai';
import { isUndefined } from 'lodash';
import { MapGeoJSONFeature } from 'maplibre-gl';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
} from 'react';
import { useMap } from 'react-map-gl/maplibre';
import { match } from 'ts-pattern';
import { useImmer } from 'use-immer';
import { fetchB6 } from '../b6';
import { useAppContext } from './app';

export type OutlinerSpec = {
    id: string;
    active?: boolean;
    properties: {
        docked: boolean;
        transient: boolean;
        coordinates: { x: number; y: number };
        tab: string;
    };
    request?: {
        eventType: Event;
        node?: NodeProto;
        expression?: string;
        locked: boolean;
        root?: FeatureIDProto;
    };
};

export type OutlinerStore = OutlinerSpec & {
    query?: ReturnType<typeof useQuery<StackResponse, Error>>;
    data?: StackResponse;
    histogram?: {
        selected?: string;
        colorScale: ScaleOrdinal<string, string>;
    };
};

const OutlinerContext = createContext<{
    outliner: OutlinerStore;
    setProperty: <K extends keyof OutlinerStore['properties']>(
        key: K,
        value: OutlinerStore['properties'][K]
    ) => void;
    highlightedFeatures: Array<{ feature: MapGeoJSONFeature; layer: string }>;
    setRequest: (request: OutlinerStore['request']) => void;
    setHistogramColorScale: (scale: ScaleOrdinal<string, string>) => void;
    setHistogramBucket: (bucket?: string) => void;
    choiceChips: Record<number, Chip>;
    setChoiceChipValue: (index: number, value: number) => void;
    close: () => void;
}>({
    outliner: {} as OutlinerStore,
    setProperty: () => {},
    highlightedFeatures: [],
    setRequest: () => {},
    setHistogramColorScale: () => {},
    setHistogramBucket: () => {},
    choiceChips: {},
    setChoiceChipValue: () => {},
    close: () => {},
});

export const useOutlinerContext = () => {
    return useContext(OutlinerContext);
};

export const OutlinerProvider = ({
    outliner,
    children,
}: PropsWithChildren & {
    outliner: OutlinerStore;
}) => {
    const { request } = outliner;
    const { setApp, closeOutliner } = useAppContext();
    const viewState = useAtomValue(viewAtom);
    const { data } = useAtomValue(startupQueryAtom);
    const { [outliner.properties.tab]: map } = useMap();

    const close = useCallback(() => {
        closeOutliner(outliner.id);
    }, [closeOutliner, outliner.id]);

    const [choiceChips, setChoiceChips] = useImmer<Record<number, Chip>>({});

    const query = useQuery({
        queryKey: [
            'outliner',
            request?.expression,
            request?.eventType,
            request?.locked,
            JSON.stringify(request?.node),
        ],
        queryFn: () => {
            if (!request) return Promise.reject('No request');
            return fetchB6('stack', {
                expression: request.expression || '',
                logEvent: request.eventType,
                locked: request.locked,
                node: request?.node,
                root: request?.root,
                logMapCenter: {
                    latE7: Math.round(viewState.latitude * 1e7),
                    lngE7: Math.round(viewState.longitude * 1e7),
                },
                logMapZoom: 0,
                session: data?.session || 0,
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled: !!request,
    });

    useEffect(() => {
        // Which substack is the choice line in? should substacks have their own context?
        const allLines =
            outliner.data?.proto.stack?.substacks.flatMap(
                (substack) => substack.lines
            ) ?? [];
        const choiceLines = allLines.flatMap((line) => line?.choice ?? []);

        choiceLines.forEach((line) => {
            line.chips.forEach((atom) => {
                if (isUndefined(atom.chip?.index)) {
                    console.warn(`Chip index is undefined`, { line, atom });
                }
                const chipIndex = atom.chip?.index ?? 0; // unsafe fallback
                choiceChips[chipIndex] = {
                    atom: {
                        labels: atom.chip?.labels ?? [],
                        index: chipIndex,
                    },
                    value: 0,
                };
            });
        });
    }, [outliner.data?.proto.stack?.substacks]);

    const setChoiceChipValue = useCallback(
        (index: number, value: number) => {
            setChoiceChips((draft) => {
                if (!draft[index]) return;
                draft[index].value = value;
            });
        },
        [setChoiceChips]
    );

    const setProperty = <K extends keyof OutlinerStore['properties']>(
        key: K,
        value: OutlinerStore['properties'][K]
    ) => {
        setApp((draft) => {
            draft.outliners[outliner.id].properties[key] = value;
        });
    };

    const setRequest = (request: OutlinerStore['request']) => {
        setApp((draft) => {
            draft.outliners[outliner.id].request = request;
        });
    };

    const setHistogramColorScale = useCallback(
        (scale: ScaleOrdinal<string, string>) => {
            setApp((draft) => {
                const histogram = draft.outliners[outliner.id].histogram;
                if (histogram) {
                    histogram.colorScale = scale;
                } else {
                    draft.outliners[outliner.id].histogram = {
                        selected: undefined,
                        colorScale: scale,
                    };
                }
            });
        },
        [setApp, outliner.id]
    );

    const setHistogramBucket = useCallback(
        (bucket?: string) => {
            setApp((draft) => {
                const histogram = draft.outliners[outliner.id].histogram;
                if (histogram) {
                    histogram.selected = bucket;
                } else {
                    draft.outliners[outliner.id].histogram = {
                        selected: bucket,
                        colorScale: undefined as $FixMe,
                    };
                }
            });
        },
        [setApp, outliner.id]
    );

    useEffect(() => {
        setApp((draft) => {
            draft.outliners[outliner.id].query = query;
            draft.outliners[outliner.id].data = query.data;
        });
    }, [query.data]);

    const highlightedFeatures = useMemo(() => {
        const highlighted = outliner.data?.proto.highlighted;
        if (!highlighted?.ids) return [];

        return highlighted.namespaces.flatMap((ns, i) => {
            const nsType = ns.match(/(?<=^\/)[a-z]+(?=\/)/)?.[0];
            return match(nsType)
                .with('path', () => {
                    return highlighted.ids[i].ids.flatMap((id) => {
                        const queryFeatures = map?.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'road',
                                filter: ['all'],
                            }
                        );

                        const feature = queryFeatures?.find(
                            (f) => parseInt(f.properties.id, 16) == id
                        );
                        return feature ? [{ feature, layer: 'road' }] : [];
                    });
                })
                .with('area', () => {
                    return highlighted.ids[i].ids.flatMap((id) => {
                        const queryFeatures = map?.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'building',
                                filter: ['all'],
                            }
                        );

                        const feature = queryFeatures?.find(
                            (f) => parseInt(f.properties.id, 16) == id
                        );
                        return feature ? [{ feature, layer: 'building' }] : [];
                    });
                })
                .otherwise(() => []);
        });
    }, [outliner.data?.proto.highlighted]);

    useEffect(() => {
        highlightedFeatures.forEach((f) => {
            if (!f) return;
            const { feature, layer } = f;
            map?.setFeatureState(
                {
                    source: 'diagonal',
                    sourceLayer: layer,
                    id: feature.id,
                },
                {
                    highlighted: true,
                }
            );
        });
        return () => {
            highlightedFeatures.forEach((f) => {
                if (!f) return;
                const { feature, layer } = f;
                map?.setFeatureState(
                    {
                        source: 'diagonal',
                        sourceLayer: layer,
                        id: feature.id,
                    },
                    {
                        highlighted: false,
                    }
                );
            });
        };
    }, [outliner.data?.proto.highlighted]);

    return (
        <OutlinerContext.Provider
            value={{
                outliner,
                setProperty,
                highlightedFeatures,
                setRequest,
                setHistogramColorScale,
                setHistogramBucket,
                choiceChips,
                setChoiceChipValue,
                close,
            }}
        >
            {children}
        </OutlinerContext.Provider>
    );
};
