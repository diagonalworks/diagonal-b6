import React, { useCallback, useEffect, useState } from 'react';
import { match } from 'ts-pattern';

import { useStack } from '@/api/stack';
import { OutlinerChangeWrapper } from '@/features/scenarios/components/OutlinerChangeWrapper';
import { useHighlight } from '@/hooks/useHighlight';
import { StackContextProvider } from '@/lib/context/stack';
import { CollectionLayer, HistogramLayer, useMapStore } from '@/stores/map';
import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { useWorldStore } from '@/stores/worlds';

import { ConditionalWrap } from './ConditionalWrap';
import { SubstackAdapter } from './adapters/SubstackAdapter';
import { Line } from './system/Line';
import { Stack } from './system/Stack';

function Outliner({
    outliner,
    side,
    origin,
}: {
    outliner: OutlinerSpec;
    side?: 'left' | 'right';
    origin?: OutlinerSpec;
}) {
    const outlinerActions = useOutlinersStore((state) => state.actions);
    const stackData = useStack(outliner.world, outliner.request, outliner.data);
    const [open, setOpen] = useState(outliner.properties.docked ? false : true);
    const mapActions = useMapStore((state) => state.actions);
    const tileLayers = useMapStore((state) => state.layers.tiles);
    const world = useWorldStore((state) => state.worlds[outliner.world]);

    useHighlight({
        outliner,
        features: stackData.data?.proto.highlighted,
    });

    useEffect(() => {
        if (outliner.properties.show && stackData.data?.geoJSON) {
            mapActions.setGeoJsonLayer(outliner.id, {
                world: outliner.world,
                features: stackData.data.geoJSON,
            });
        }

        if (outliner.properties.show && stackData.data?.proto.layers) {
            for (const layer of stackData.data.proto.layers) {
                if (tileLayers[`${outliner.id}-${layer.path}`]) {
                    continue;
                }
                const tileLayer = match(layer.path)
                    .with('histogram', () => {
                        const l: HistogramLayer = {
                            world: outliner.world,
                            outliner: outliner.id,
                            type: 'histogram',
                            spec: {
                                tiles: `tiles/${layer.path}/{z}/{x}/{y}.mvt?q=${layer.q}&r=collection/${world.featureId.namespace}/${world.featureId.value}`,
                                selected: undefined,
                                showOnMap: outliner.properties.show,
                            },
                        };
                        return l;
                    })
                    .with('collection', () => {
                        const l: CollectionLayer = {
                            world: outliner.world,
                            outliner: outliner.id,
                            type: 'collection',
                            spec: {
                                tiles: `tiles/${layer.path}/{z}/{x}/{y}.mvt?q=${layer.q}&r=collection/${world.featureId.namespace}/${world.featureId.value}`,
                                showOnMap: outliner.properties.show,
                            },
                        };
                        return l;
                    })
                    .otherwise(() => null);
                if (tileLayer) {
                    mapActions.setTileLayer(
                        `${outliner.id}-${layer.path}`,
                        tileLayer
                    );
                }
            }
        }
    }, [
        outliner.id,
        outliner.world,
        outliner.properties.show,
        stackData.data?.geoJSON,
        mapActions,
        stackData.data?.proto.layers,
        world.featureId.namespace,
        world.featureId.value,
    ]);

    useEffect(() => {
        if (outliner.properties.show) {
            mapActions.showOutlinerLayers(outliner.id);
        } else {
            mapActions.hideOutlinerLayers(outliner.id);
        }
    }, [outliner.properties.show, mapActions, outliner.id]);

    useEffect(() => {
        return () => {
            mapActions.removeGeoJsonLayer(outliner.id);
            mapActions.removeOutlinerLayers(outliner.id);
        };
    }, []);

    const handleOpenChange = useCallback(
        (open: boolean) => {
            if (outliner.properties.docked) {
                outlinerActions.setActive(outliner.id, open);
                outlinerActions.setVisibility(outliner.id, open);
                setOpen(open);
            }
        },
        [outlinerActions, outliner.id]
    );

    if (stackData.isLoading) {
        return <LoadingStack expression={outliner.request?.expression || ''} />;
    }

    if (!stackData.data) return null;

    const substacks = stackData.data.proto.stack?.substacks;

    const firstSubstack = substacks?.[0];
    const otherSubstacks = substacks?.slice(1);

    return (
        <StackContextProvider outliner={outliner} origin={origin}>
            <ConditionalWrap
                condition={
                    side === 'right' &&
                    outliner.properties.type !== 'comparison'
                }
                wrap={(children) => (
                    <OutlinerChangeWrapper
                        id={outliner.world}
                        outliner={outliner}
                        stack={stackData.data?.proto.stack}
                    >
                        {children}
                    </OutlinerChangeWrapper>
                )}
            >
                <Stack
                    collapsible={outliner.properties.docked}
                    open={open}
                    onOpenChange={handleOpenChange}
                >
                    {firstSubstack && (
                        <Stack.Trigger className="w-80">
                            <SubstackAdapter
                                substack={firstSubstack}
                                collapsible={firstSubstack.collapsable}
                                close={!outliner.properties.docked}
                                show={!outliner.properties.docked}
                                copy
                            />
                        </Stack.Trigger>
                    )}
                    {otherSubstacks && (
                        <Stack.Content className="w-80">
                            {otherSubstacks.map((substack, i) => {
                                return (
                                    <SubstackAdapter
                                        key={i}
                                        substack={substack}
                                        collapsible={substack.collapsable}
                                    />
                                );
                            })}
                        </Stack.Content>
                    )}
                </Stack>
            </ConditionalWrap>
        </StackContextProvider>
    );
}

const memorizedOutliner = React.memo(Outliner);
export default memorizedOutliner;

const LoadingStack = ({ expression }: { expression: string }) => {
    return (
        <Stack>
            <Stack.Trigger>
                <Line className="flex flex-nowrap w-80 ">
                    <div className="loader shrink-0" />
                    <div className="text-graphite-60 italic text-nowrap overflow-hidden overflow-ellipsis">
                        {expression}
                    </div>
                </Line>
            </Stack.Trigger>
        </Stack>
    );
};
