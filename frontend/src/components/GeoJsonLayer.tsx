import * as circleIcons from '@/assets/icons/circle';
import { useStacks } from '@/lib/api/stack';
import { useOutlinersStore } from '@/stores/outliners';
import { World } from '@/stores/worlds';
import { $FixMe } from '@/utils/defs';
import { DotIcon } from '@radix-ui/react-icons';
import React, { useMemo } from 'react';
import { Marker, useMap as useMapLibre } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';

const Icon = ({
    b6Icon,
    side,
}: {
    b6Icon: $FixMe | undefined;
    side: 'left' | 'right';
}) => {
    const icon = match(b6Icon)
        .with('dot', () => {
            return (
                <DotIcon
                    className={twMerge(
                        'fill-graphite-80',
                        side === 'right' && 'fill-rose-80'
                    )}
                />
            );
        })
        .otherwise(() => {
            const icon = b6Icon;
            if (!icon)
                return (
                    <div
                        className={twMerge(
                            'w-2 h-2 rounded-full bg-ultramarine-50 border border-ultramarine-80',
                            side === 'right' && 'bg-rose-50 border-rose-80'
                        )}
                    />
                );
            const iconComponentName = `${icon
                .charAt(0)
                .toUpperCase()}${icon.slice(1)}`;
            if (circleIcons[iconComponentName as keyof typeof circleIcons]) {
                const Icon =
                    circleIcons[iconComponentName as keyof typeof circleIcons];
                return <Icon />;
            }
            return <DotIcon />;
        });
    return icon;
};

function GeoJsonLayer({
    world,
    side,
}: {
    world: World['id'];
    side: 'left' | 'right';
}) {
    const { [world]: map } = useMapLibre();
    const outliners = useOutlinersStore((state) =>
        state.actions.getByWorld(world)
    );

    const activeOutliners = useMemo(() => {
        return Object.values(outliners).filter(
            (outliner) =>
                outliner.world === world &&
                (outliner.properties.active || outliner.properties.transient)
        );
    }, [outliners, world]);

    const queries = useStacks(
        activeOutliners.map((outliner) => ({
            request: outliner?.request,
            fallback: outliner.data,
        }))
    );

    const features = useMemo(() => {
        return queries
            .flatMap((query) => {
                return query?.data?.geoJSON ?? [];
            })
            .flatMap((g: $FixMe) => {
                if (g.type === 'FeatureCollection') return g.features;
                if (g.type === 'Feature') return [g];
                return [];
            })
            .flat();
    }, [outliners]);

    const markers = useMemo(() => {
        return features.filter(
            (f: $FixMe) =>
                f.geometry?.type === 'Point' &&
                map
                    ?.getBounds()
                    ?.contains(f.geometry.coordinates as [number, number])
        );
    }, [features, map]);

    return (
        <>
            {markers.map((marker, i) => {
                return (
                    <Marker
                        key={i}
                        longitude={marker.geometry.coordinates[0]}
                        latitude={marker.geometry.coordinates[1]}
                        className={twMerge(
                            '[&>svg]:fill-graphite-80',
                            side === 'right' && '[&>svg]:fill-rose-80'
                        )}
                    >
                        <Icon
                            b6Icon={marker.properties?.['-b6-icon']}
                            side={side}
                        />
                    </Marker>
                );
            })}
        </>
    );
}

const memoizedGeoJsonLayer = React.memo(GeoJsonLayer);
export default memoizedGeoJsonLayer;
