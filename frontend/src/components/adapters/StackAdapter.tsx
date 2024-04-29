import { AppStore } from '@/atoms/app';
import { StackContextProvider } from '@/lib/context/stack';
import { useEffect, useMemo, useState } from 'react';
import { useMap } from 'react-map-gl/maplibre';
import { match } from 'ts-pattern';
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

        return stack.proto.highlighted.namespaces.flatMap((namespace, i) => {
            console.log('namespace', namespace);
            const nsType = namespace.match(/(?<=^\/)[a-z]+(?=\/)/)?.[0];
            return match(nsType)
                .with('path', () => {
                    return stack.proto.highlighted?.ids[i].ids.flatMap((id) => {
                        console.log('id', id);
                        const queryFeatures = map?.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'road',
                                filter: ['all'],
                            }
                        );

                        // this is not ideal for performance, but it's fine for now
                        const feature = queryFeatures?.find((f) => {
                            return parseInt(f.properties.id, 16) == id;
                        });
                        console.log('feature', feature);
                        return feature
                            ? [
                                  {
                                      feature,
                                      layer: 'road',
                                  },
                              ]
                            : [];
                    });
                })
                .with('area', () => {
                    return stack.proto.highlighted?.ids[i].ids.flatMap((id) => {
                        console.log('id', id);
                        const queryFeatures = map?.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'building',
                                filter: ['all'],
                            }
                        );

                        // this is not ideal for performance, but it's fine for now
                        const feature = queryFeatures?.find((f) => {
                            return parseInt(f.properties.id, 16) == id;
                        });
                        return feature
                            ? [
                                  {
                                      feature,
                                      layer: 'building',
                                  },
                              ]
                            : [];
                    });
                })
                .otherwise(() => []);
        });
    }, [stack.proto.highlighted]);

    useEffect(() => {
        console.log('highlightedFeatures', highlightedFeatures);
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
    }, [stack.proto.highlighted]);

    if (!stack.proto.stack) return null;

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
