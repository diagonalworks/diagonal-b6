import { AppStore } from '@/atoms/app';
import { StackContextProvider } from '@/lib/context/stack';
import { useEffect, useMemo, useState } from 'react';
import { useMap } from 'react-map-gl/maplibre';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

export const StackAdapter = ({
    stack,
    docked = false,
    mapId,
}: {
    stack: AppStore['stacks'][string];
    docked?: boolean;
    mapId: string;
}) => {
    const [open, setOpen] = useState(docked ? false : true);
    const { [mapId]: map } = useMap();

    const highlightedFeatures = useMemo(() => {
        if (!stack.proto.highlighted?.ids) return [];

        return stack.proto.highlighted.ids.flatMap(({ ids }) => {
            return ids.flatMap((id) => {
                const queryFeatures = map?.querySourceFeatures('diagonal', {
                    sourceLayer: 'building',
                    filter: ['all'],
                });

                // this is not ideal for performance, but it's fine for now
                const feature = queryFeatures?.find((f) => {
                    return parseInt(f.properties.id, 16) == id;
                });
                return feature ? [feature] : [];
            });
        });
    }, [stack.proto.highlighted]);

    useEffect(() => {
        highlightedFeatures.forEach((feature) => {
            map?.setFeatureState(
                {
                    source: 'diagonal',
                    sourceLayer: 'building',
                    id: feature.id,
                },
                {
                    highlighted: true,
                }
            );
        });
        return () => {
            highlightedFeatures.forEach((feature) => {
                map?.setFeatureState(
                    {
                        source: 'diagonal',
                        sourceLayer: 'building',
                        id: feature.id,
                    },
                    {
                        highlighted: false,
                    }
                );
            });
        };
    }, [stack.proto.highlighted]);

    if (!stack.proto.stack) return null;
    //console.log(stack.proto);

    const firstSubstack = stack.proto.stack.substacks[0];
    const otherSubstacks = stack.proto.stack.substacks.slice(1);

    return (
        <div>
            <StackContextProvider stack={stack} mapId={mapId}>
                <Stack collapsible={docked} open={open} onOpenChange={setOpen}>
                    <Stack.Trigger>
                        <SubstackAdapter
                            substack={firstSubstack}
                            collapsible={firstSubstack.collapsable}
                        />
                    </Stack.Trigger>
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
                </Stack>
            </StackContextProvider>
        </div>
    );
};
