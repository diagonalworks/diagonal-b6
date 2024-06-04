import { Feature, GeoJsonProperties, Geometry } from 'geojson';
import { MapGeoJSONFeature, StyleSpecification } from 'maplibre-gl';
import { match } from 'ts-pattern';

/**
 * Check if two features are the same point, with a given precision.
 * @param f1 First feature
 * @param f2 Second feature
 * @param precision Precision to compare coordinates
 * @returns Boolean if the features are the same point
 */
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

/**
 * Get the b6 feature path from a MapGeoJSONFeature.
 * @param feature MapGeoJSONFeature
 * @returns Feature path
 */
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

/**
 * Get the road width based on the road type.
 * @param type Road type
 * @returns Road width
 */
export const getRoadWidth = (type: string) => {
    return match(type)
        .with('motorway', 'trunk', () => 1.5)
        .with('primary', () => 1.2)
        .with('secondary', 'tertiary', 'street', () => 0.8)
        .with('unclassified', 'residential', 'service', () => 1)
        .with('cycleway', 'footway', 'path', () => 0.8)
        .otherwise(() => 1);
};

/**
 * Returns a copy of the map style with the tiles property for the diagonal source changed.
 * @param mapStyle Original diagonal map style,
 * @param source New source URL for the diagonal tiles
 * @returns Updated map style
 */
export const changeMapStyleSource = (
    mapStyle: StyleSpecification,
    source: string
): StyleSpecification => {
    return {
        ...mapStyle,
        sources: {
            ...mapStyle.sources,
            diagonal: {
                ...mapStyle.sources.diagonal,
                tiles: [source],
            },
        },
    } as StyleSpecification;
};

/**
 * Get the tile source URL for the diagonal map.
 * @param root Optional root parameter for the tile source
 * @returns Tile source URL
 */
export const getTileSource = (root?: string) => {
    return `${window.location.origin}/tiles/base/{z}/{x}/{y}.mvt${
        root ? `?r=${root}` : ''
    }`;
};
