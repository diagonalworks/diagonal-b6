import { Map } from '@/components/Map';
import { useStartup } from '@/lib/api/startup';
import { useWorkspaceStore } from '@/stores/workspace';
import { World as WorldT } from '@/stores/worlds';
import { FeatureIDProto, FeatureType } from '@/types/generated/api';
import { useMemo } from 'react';
import GeoJsonLayer from './GeoJsonLayer';
import OutlinersLayer from './OutlinersLayer';

const getWorldFeatureId = (
    worldId: string,
    namespace?: string,
    collection?: string
): FeatureIDProto => {
    const featureId = {
        type: 'FeatureTypeCollection' as unknown as FeatureType,
        namespace: namespace ?? 'diagonal.works',
        value: 0,
    };

    if (worldId === 'baseline') {
        const baselineValue = collection?.match(/(?<=\/)\d*$/)?.[0];
        featureId.value = parseInt(baselineValue ?? '0');
    } else {
        featureId.namespace = `${featureId.namespace}/scenario`;
        featureId.value = +worldId;
    }
    return featureId;
};

export default function World({
    id,
    side,
}: {
    id: WorldT['id'];
    side: 'left' | 'right';
}) {
    const root = useWorkspaceStore((state) => state.root);
    const startup = useStartup();

    const featureId = useMemo(() => {
        return getWorldFeatureId(id, startup.data?.root?.namespace, root);
    }, [id, startup.data?.root?.namespace, root]);

    const mapRoot = useMemo(() => {
        return `collection/${featureId.namespace}/${featureId.value}`;
    }, [featureId.namespace, featureId.value]);

    return (
        <div className=" w-full h-full absolute top-0 left-0">
            <Map root={mapRoot} side={side} world={id}>
                <OutlinersLayer world={id} />
                <GeoJsonLayer world={id} side={side} />
            </Map>
        </div>
    );
}
