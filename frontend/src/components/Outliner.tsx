import { OutlinerChangeWrapper } from '@/features/scenarios/components/OutlinerChangeWrapper';
import { useHighlight } from '@/hooks/useHighlight';
import { useStack } from '@/lib/api/stack';
import { StackContextProvider } from '@/lib/context/stack';
import { useMapStore } from '@/stores/map';
import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import React, { useCallback, useEffect, useState } from 'react';
import { ConditionalWrap } from './ConditionalWrap';
import { SubstackAdapter } from './adapters/SubstackAdapter';
import { Line } from './system/Line';
import { Stack } from './system/Stack';

function Outliner({
    outliner,
    origin,
}: {
    outliner: OutlinerSpec;
    origin?: OutlinerSpec;
}) {
    const outlinerActions = useOutlinersStore((state) => state.actions);
    const stackData = useStack(outliner.world, outliner.request, outliner.data);
    const [open, setOpen] = useState(outliner.properties.docked ? false : true);
    const mapActions = useMapStore((state) => state.actions);
    useHighlight({
        world: outliner.world,
        features: stackData.data?.proto.highlighted,
    });

    useEffect(() => {
        if (
            (outliner.properties.active || outliner.properties.transient) &&
            stackData.data?.geoJSON
        ) {
            mapActions.setGeoJsonLayer(outliner.id, {
                world: outliner.world,
                features: stackData.data.geoJSON,
            });
        }
    }, [
        outliner.id,
        outliner.world,
        outliner.properties.active,
        outliner.properties.transient,
        stackData.data?.geoJSON,
        mapActions,
    ]);

    useEffect(() => {
        return () => {
            mapActions.removeGeoJsonLayer(outliner.id);
        };
    }, []);

    const handleOpenChange = useCallback(
        (open: boolean) => {
            outlinerActions.setActive(outliner.id, open);
            setOpen(open);
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
                    outliner.world !== 'baseline' &&
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
                        <Stack.Trigger>
                            <SubstackAdapter
                                substack={firstSubstack}
                                collapsible={firstSubstack.collapsable}
                                close={!outliner.properties.docked}
                            />
                        </Stack.Trigger>
                    )}
                    {otherSubstacks && (
                        <Stack.Content>
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
                <Stack.Trigger>
                    <Line className="flex flex-nowrap ">
                        <div className="loader shrink-0" />
                        <div className="text-graphite-60 italic text-nowrap overflow-hidden overflow-ellipsis">
                            {expression}
                        </div>
                    </Line>
                </Stack.Trigger>
            </Stack.Trigger>
        </Stack>
    );
};
