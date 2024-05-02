import { Feature, GeoJsonProperties, Geometry } from 'geojson';
import { MapGeoJSONFeature } from 'maplibre-gl';
import { match } from 'ts-pattern';

export const isSamePositionPoints = (
    f1: Feature<Geometry, GeoJsonProperties>,
    f2: Feature<Geometry, GeoJsonProperties>,
    precision: number = 6
) => {
    if (f1.geometry.type !== 'Point' || f2.geometry.type !== 'Point')
        return false;
    return (
        f1.geometry.coordinates[0].toFixed(precision) ===
            f2.geometry.coordinates[0].toFixed(precision) &&
        f1.geometry.coordinates[1].toFixed(precision) ===
            f2.geometry.coordinates[1].toFixed(precision)
    );
};

export const getFeaturePath = (feature: MapGeoJSONFeature) => {
    const { ns, id } = feature.properties;
    const type = match(feature.geometry.type)
        .with('Point', () => 'point')
        .with('LineString', () => 'path')
        .with('Polygon', () => 'area')
        .with('MultiPolygon', () => 'area')
        .otherwise(() => null);
    if (ns && id && type) {
        return `/${type}/${ns}/${BigInt(`0x${id}`)}`;
    }
};

export const getRoadWidth = (type: string) => {
    return match(type)
        .with('motorway', 'trunk', () => 1.5)
        .with('primary', () => 1.2)
        .with('secondary', 'tertiary', 'street', () => 0.8)
        .with('unclassified', 'residential', 'service', () => 1)
        .with('cycleway', 'footway', 'path', () => 0.8)
        .otherwise(() => 1);
};
